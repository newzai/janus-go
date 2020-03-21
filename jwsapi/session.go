package jwsapi

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/newzai/janus-go/logging"
	"github.com/pkg/errors"
)

//Session janus gateway session
type Session struct {
	ctx           context.Context
	ID            uint64
	isDestroy     int32
	conn          *Connection
	handles       map[uint64]*Handle
	handlesCancel map[uint64]context.CancelFunc
	tasks         chan func(*Session)
}

//NewSession new session
func NewSession(ctx context.Context, id uint64, conn *Connection) *Session {

	s := &Session{
		ctx:           ctx,
		ID:            id,
		conn:          conn,
		handles:       make(map[uint64]*Handle),
		handlesCancel: make(map[uint64]context.CancelFunc),
		tasks:         make(chan func(*Session), 128),
	}

	go s.execLoop()
	return s
}

//IsDestroy session has destroy
func (s *Session) IsDestroy() bool {
	if atomic.LoadInt32(&s.isDestroy) == 1 {
		return true
	}
	return false
}

func (s *Session) addHandle(h *Handle, cancel context.CancelFunc) {
	s.run(func(ss *Session) {
		s.handles[h.ID] = h
		s.handlesCancel[h.ID] = cancel
	})

}
func (s *Session) delHandle(hid uint64) {
	s.run(func(ss *Session) {
		if cancel, ok := s.handlesCancel[hid]; ok {
			cancel()
			delete(s.handles, hid)
			delete(s.handlesCancel, hid)
		}
	})
}

func (s *Session) run(f func(*Session)) {
	if s.IsDestroy() {
		return
	}

	select {
	case s.tasks <- f:
	default:
	}
}

func (s *Session) onMessage(conn *Connection, msg *Message) {

	if hid, ok := msg.HandleID(); ok {
		if h, ok := s.handles[hid]; ok {
			h.onMessage(msg)
		} else {
			logging.Warnf("Session[%d] can't Handle[%d]", s.ID, hid)
		}
	} else {
		logging.Warnf("Session[%s] can't find handle_id at event msg", s.ID)
	}

}

//Request send request, has success response
func (s *Session) Request(msg Message) (*Message, error) {

	msg[attrSessionID] = s.ID
	return s.conn.Request(msg)
}

//Message send message to janus, has ack ,event resposne
func (s *Session) Message(msg Message) (*Message, error) {
	msg[attrSessionID] = s.ID
	return s.conn.Message(msg)
}

//Attach new handle from gateway
func (s *Session) Attach(pluginName string) (*Handle, error) {

	msg := Message{
		attrType:      "attach",
		"plugin":      pluginName,
		attrSessionID: s.ID,
	}
	rsp, err := s.conn.Request(msg)
	if err != nil {
		return nil, err
	}
	data := rsp.Data()
	if id, ok := data.Uint64("id"); ok {
		ctx, cancel := context.WithCancel(s.ctx)
		newH := NewHandle(ctx, id, s)
		s.addHandle(newH, cancel)
		return newH, nil
	}
	return nil, errors.New("rsp error")
}

//Destroy  send destroy
func (s *Session) Destroy() error {

	msg := Message{
		attrType:      "destroy",
		attrSessionID: s.ID,
	}

	_, err := s.conn.Request(msg)

	s.conn.delSession(s.ID)
	return err
}

//claim for ws close,re connection
func (s *Session) claim() {

	msg := Message{
		attrType:      "claim",
		attrSessionID: s.ID,
	}

	_, err := s.conn.Request(msg)
	if err != nil {
		logging.Errorf("[%s] claim err:%v", s.ID, err)
		s.conn.delSession(s.ID)
		return
	}

	logging.Infof("[%d] claim OK", s.ID)
}

func (s *Session) keepalive() error {
	msg := Message{
		attrType:      "keepalive",
		attrSessionID: s.ID,
	}
	_, err := s.conn.Request(msg)
	return err
}

func (s *Session) execLoop() {

	ticker := time.NewTicker(10 * time.Second)
	defer func() {

		logging.Infof("[%d] Session End", s.ID)
		atomic.StoreInt32(&s.isDestroy, 1)
		s.conn.delSession(s.ID)
		ticker.Stop()
		close(s.tasks)
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			//send keepalive...
			s.keepalive()
		case t, ok := <-s.tasks:
			if !ok {
				return
			}
			t(s)
		}
	}
}
