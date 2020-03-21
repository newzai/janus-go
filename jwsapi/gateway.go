package jwsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/newzai/janus-go/logging"
	"github.com/pkg/errors"
)

const (
	writeWait      = 3 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type onResponse func(*Message)
type cstate int

const (
	connectioing cstate = 1
	connectioned cstate = 2
	closed       cstate = 3
)

type connState struct {
	state cstate
	ts    time.Time
}

func getTID() string {
	uid := uuid.New()
	return uid.String()
}

//Connection a ws connection
// auto reConnection if ws conn is closed
// auto claim all session when reConnection is success
type Connection struct {
	ctx                 context.Context
	isDestroy           int32
	id                  int
	url                 string
	cc                  int // connection count
	conn                *websocket.Conn
	connStateChan       chan cstate
	connCtx             context.Context
	connCancel          context.CancelFunc
	recvChan            chan *Message
	sendChan            chan []byte
	tasks               chan func(*Connection)
	readTimeoutSeconds  int
	writeTimeoutSeconds int
	retrySeconds        int
	readBufferLimit     int64
	transactions        map[string]onResponse
	sessions            map[uint64]*Session
	sessionCalcels      map[uint64]context.CancelFunc
	state               connState
}

//NewConnection create new janus gateway connection
func NewConnection(ctx context.Context, url string, id int) *Connection {
	conn := &Connection{
		ctx:                 ctx,
		isDestroy:           0,
		id:                  id,
		url:                 url,
		connStateChan:       make(chan cstate, 16),
		recvChan:            make(chan *Message, 1024),
		sendChan:            make(chan []byte, 1024),
		tasks:               make(chan func(*Connection), 1024),
		transactions:        make(map[string]onResponse),
		sessions:            make(map[uint64]*Session),
		sessionCalcels:      make(map[uint64]context.CancelFunc),
		readTimeoutSeconds:  3,
		writeTimeoutSeconds: 3,
		retrySeconds:        3,
		readBufferLimit:     maxMessageSize,
		state: connState{
			state: connectioing,
			ts:    time.Now(),
		},
	}

	go conn.execLoop()

	return conn
}

//ID return this connection id
func (c *Connection) ID() string {
	return fmt.Sprintf("[%s@%d_%d]", c.url, c.id, c.cc)
}

//IsDestroy is destroy for this object
func (c *Connection) IsDestroy() bool {
	if atomic.LoadInt32(&c.isDestroy) == 1 {
		return true
	}
	return false
}

func (c *Connection) tryConnection() {

	c.cc++
	c.connCtx, c.connCancel = context.WithCancel(c.ctx)

	websocket.DefaultDialer.Subprotocols = []string{"janus-protocol"}
	websocket.DefaultDialer.HandshakeTimeout = time.Duration(c.retrySeconds) * time.Second
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {

		logging.Warnf("%s connection err:%v", c.ID(), err)
		<-time.After(time.Duration(c.retrySeconds) * time.Second)
		select {
		case <-c.ctx.Done():
		default:
		}
		return
	}

	c.conn = conn
	go c.readDump(c.conn)
	go c.writeDump(c.conn)
	c.connStateChan <- connectioned

	return
}

func (c *Connection) readDump(conn *websocket.Conn) {

	defer func() {
		logging.Infof("%s readDump End", c.ID())
		conn.Close()
	}()
	conn.SetReadLimit(c.readBufferLimit)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, data, err := conn.ReadMessage()
			if err != nil {
				logging.Warnf("%s read err %s", c.ID(), err)
				c.connStateChan <- closed
				return
			}
			logging.Infof("%s recv %s", c.ID(), string(data))
			message := bytes.TrimSpace(bytes.Replace(data, newline, space, -1))

			decoder := json.NewDecoder(bytes.NewBuffer(message))
			decoder.UseNumber()
			msg := make(Message)
			err = decoder.Decode(&msg)
			if err == nil {
				select {
				case c.recvChan <- &msg:
				default:
					t, ok := msg.Transaction()
					logging.Warnf("%s post message %s.%s(%t) failed", msg.Type(), t, ok)
				}
			} else {
				logging.Errorf("%s Decode err %v", c.ID(), err)
			}
		}
	}
}

func (c *Connection) writeDump(conn *websocket.Conn) {

	defer func() {
		logging.Infof("%s writeDump End", c.ID())
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case data := <-c.sendChan:
			conn.SetWriteDeadline(time.Now().Add(time.Duration(c.writeTimeoutSeconds) * time.Second))
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				logging.Errorf("%s write err %v", c.ID(), err)
				c.connStateChan <- closed
				return
			}
			logging.Infof("%s write ok %s", c.ID(), string(data))

		}
	}
}

func (c *Connection) execLoop() {

	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		logging.Infof("%s exec is Done", c.ID())
		atomic.StoreInt32(&c.isDestroy, 1)
		close(c.tasks)
		close(c.recvChan)
		close(c.sendChan)
		ticker.Stop()
	}()

	go c.tryConnection()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if c.state.state == closed && time.Since(c.state.ts) > time.Duration(10)*time.Second {
				//conn is disconnection to long time. release all session

				for sid := range c.sessions {
					logging.Warnf("%s delSession[%d] all for disconnection too long", c.ID(), sid)
					go c.delSession(sid)
				}
			}
		case state := <-c.connStateChan:

			logging.Infof("%s retrying %d", c.ID(), state)
			c.state.state = state
			c.state.ts = time.Now()
			switch state {
			case closed:
				c.connCancel()

				go c.tryConnection()
			case connectioned:
				//链接建立..

				for _, sess := range c.sessions {
					go sess.claim()
				}
			}

		case t, ok := <-c.tasks:
			if !ok {
				return
			}
			t(c)
		case msg, ok := <-c.recvChan:
			if !ok {
				return
			}
			tid, ok := msg.Transaction()
			if ok {
				//find tid
				trans, ok := c.transactions[tid]
				if ok {
					trans(msg)
				} else {
					logging.Warnf("%s can't find trans by %s", c.ID(), tid)
				}
			} else {
				sid, ok := msg.SessionID()
				if ok {
					//find session
					sess, ok := c.sessions[sid]
					if ok {
						sess.onMessage(c, msg)
					} else {
						logging.Warnf("%s can't find session by %d", c.ID(), sid)
					}
				} else {
					logging.Warnf("%s unknown doing msg %v", c.ID(), msg)
				}
			}
		}
	}
}

func (c *Connection) run(f func(*Connection)) {
	if c.IsDestroy() {
		return
	}

	select {
	case c.tasks <- f:
	default:
	}
}

func (c *Connection) addSession(s *Session, cancel context.CancelFunc) {
	c.run(func(cc *Connection) {
		cc.sessions[s.ID] = s
		cc.sessionCalcels[s.ID] = cancel
	})
}

func (c *Connection) delSession(sid uint64) {
	c.run(func(cc *Connection) {
		if cancel, ok := cc.sessionCalcels[sid]; ok {
			cancel()
			delete(cc.sessions, sid)
			delete(cc.sessionCalcels, sid)
		}

	})
}

func (c *Connection) delTransaction(tid string) {
	c.run(func(cc *Connection) {
		delete(c.transactions, tid)
	})
}

func (c *Connection) sendMessage(tid string, msg Message, callback onResponse) {
	if c.IsDestroy() {
		return
	}
	c.run(func(cc *Connection) {
		cc.transactions[tid] = callback
	})

	data, _ := json.Marshal(msg)
	c.sendChan <- data
}

//Request send request ,has success response
func (c *Connection) Request(request Message) (*Message, error) {
	if c.IsDestroy() {
		return nil, errors.New("conn is destroy")
	}
	var tid string
	if _, ok := request[attrTransaction]; !ok {
		tid = getTID()
		request[attrTransaction] = tid
	} else {
		tid = request[attrTransaction].(string)
	}
	defer func() {
		c.delTransaction(tid)
	}()
	result := make(chan *Message)

	c.sendMessage(tid, request, func(rsp *Message) {
		result <- rsp
	})
	select {
	case rsp := <-result:
		return rsp, rsp.Error()
	case <-time.After(3 * time.Second):
	}
	return nil, errors.New("timeout")

}

//Message send message, has ack, event response
func (c *Connection) Message(msg Message) (*Message, error) {
	if c.IsDestroy() {
		return nil, errors.New("conn is destroy")
	}

	var tid string
	if _, ok := msg[attrTransaction]; !ok {
		tid = getTID()
		msg[attrTransaction] = tid
	} else {
		tid = msg[attrTransaction].(string)
	}

	defer func() {
		c.delTransaction(tid)
	}()
	result := make(chan *Message)

	c.sendMessage(tid, msg, func(rsp *Message) {
		result <- rsp
	})
	select {
	case rsp := <-result:
		if rsp.IsError() {
			return nil, rsp.Error()
		}
	case <-time.After(3 * time.Second):
		return nil, errors.New("timeout")
	}

	select {
	case rsp := <-result:
		return rsp, rsp.Error()
	case <-time.After(3 * time.Second):
		return nil, errors.New("timeout")
	}
}

//Create ceeate New Session
func (c *Connection) Create() (*Session, error) {

	msg := Message{
		attrType: "create",
	}
	rsp, err := c.Request(msg)
	if err != nil {
		return nil, err
	}
	data := rsp.Data()
	if id, ok := data.Uint64("id"); ok {
		ctx, cancel := context.WithCancel(c.ctx)
		newSess := NewSession(ctx, id, c)
		c.addSession(newSess, cancel)
		return newSess, nil
	}
	return nil, errors.New("response err")
}
