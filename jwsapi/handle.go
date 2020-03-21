package jwsapi

import (
	"context"
	"sync/atomic"

	"github.com/newzai/janus-go/logging"
	"github.com/pkg/errors"
)

//Handle janus gateway plugin handle
type Handle struct {
	ctx       context.Context
	ID        uint64
	isDestroy int32
	s         *Session
	Events    chan *Message

	onWebrtcup func(Message)
	onMedia    func(Message)
	onSlowLink func(Message)
	onHangup   func(Message)
	onTrickle  func(Message)
}

//HandleCallbackOption setting option
type HandleCallbackOption func(*Handle)

//WithHandleWebrtcup set webrtcup callback..
func WithHandleWebrtcup(callback func(Message)) HandleCallbackOption {
	return func(h *Handle) {
		h.onWebrtcup = callback
	}
}

//WithHandleMedia set on media callback
func WithHandleMedia(callback func(Message)) HandleCallbackOption {
	return func(h *Handle) {
		h.onMedia = callback
	}
}

//WithHandleSlowLink set slow link callback
func WithHandleSlowLink(callback func(Message)) HandleCallbackOption {
	return func(h *Handle) {
		h.onSlowLink = callback
	}
}

//WithHandleHangup set hangup callback
func WithHandleHangup(callback func(Message)) HandleCallbackOption {
	return func(h *Handle) {
		h.onHangup = callback
	}
}

//WithHandleTrickle set trickle callback
func WithHandleTrickle(callback func(Message)) HandleCallbackOption {
	return func(h *Handle) {
		h.onTrickle = callback
	}
}

//NewHandle new handle.
func NewHandle(ctx context.Context, id uint64, sess *Session) *Handle {

	h := &Handle{
		ctx:    ctx,
		ID:     id,
		s:      sess,
		Events: make(chan *Message),
	}

	go h.execLoop()

	return h
}

//IsDestroy is destroy
func (h *Handle) IsDestroy() bool {
	if atomic.LoadInt32(&h.isDestroy) == 1 {
		return true
	}
	return false
}

//SetCallback set callback using WithHandleWebrtcup,WithHandleMedia...
func (h *Handle) SetCallback(opts ...HandleCallbackOption) {

	for _, opt := range opts {
		opt(h)
	}
}

//Request send request, has success response
func (h *Handle) Request(body Message) (*Message, error) {
	if h.IsDestroy() {
		return nil, errors.New("has detach")
	}

	msg := Message{
		attrType:     "message",
		attrHandleID: h.ID,
		attrBody:     body,
	}

	return h.s.Request(msg)
}

//Message send message to janus, has ack ,event resposne
func (h *Handle) Message(body Message) (*Message, error) {
	if h.IsDestroy() {
		return nil, errors.New("has detach")
	}
	msg := Message{
		attrType:     "message",
		attrHandleID: h.ID,
		attrBody:     body,
	}
	return h.s.Message(msg)
}

//JsepMessage jsep use to send Offer or Answer
func (h *Handle) JsepMessage(body Message, jsep Message) (*Message, error) {
	if h.IsDestroy() {
		return nil, errors.New("has detach")
	}

	msg := Message{
		attrType:     "message",
		attrHandleID: h.ID,
		attrBody:     body,
		attrJSEP:     jsep,
	}
	return h.s.Message(msg)

}

//Trickle send local candidae
func (h *Handle) Trickle(candidate Message) error {

	msg := Message{
		attrType:    "trickle",
		"candidate": candidate,
	}

	_, err := h.s.Request(msg)
	return err
}

//Detach release plugin handle at janus-gateway
func (h *Handle) Detach() error {
	msg := Message{
		attrType:     "detach",
		attrHandleID: h.ID,
	}
	_, err := h.s.Request(msg)

	h.s.delHandle(h.ID)

	return err
}

func (h *Handle) onMessage(msg *Message) {
	if h.IsDestroy() {
		return
	}
	switch msg.Type() {
	case "event":
		h.Events <- msg
	case "webrtcup":
		if h.onWebrtcup != nil {
			h.onWebrtcup(*msg)
		}
	case "media":
		if h.onMedia != nil {
			h.onMedia(*msg)
		}
	case "trickle":
		if h.onTrickle != nil {
			h.onTrickle(*msg)
		}
	case "hangup":
		if h.onHangup != nil {
			h.onHangup(*msg)
		}
	}

}

func (h *Handle) execLoop() {
	defer func() {
		logging.Infof("Handle[%d.%d] End", h.s.ID, h.ID)
		atomic.StoreInt32(&h.isDestroy, 1)
		close(h.Events)

	}()
	for {
		select {
		case <-h.ctx.Done():
			if h.onHangup != nil {
				h.onHangup(Message{attrType: "hangup", "reason": "ctx.Done"})
			}
			return

		}
	}
}
