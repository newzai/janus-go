package jvideoroom

import (
	"context"

	"github.com/newzai/janus-go/jwsapi"
	"github.com/pkg/errors"
)

//Subscriber a subscriber
type Subscriber struct {
	ctx    context.Context
	handle *jwsapi.Handle
	room   uint64
	feed   uint64
}

//WithSubscriberPubID set publisher_id for subscriber
func WithSubscriberPubID(pubID uint64) jwsapi.MessageOption {
	return func(msg jwsapi.Message) {
		msg["publisher_id"] = pubID
	}
}

//NewSubscriber create a subscriber
func NewSubscriber(ctx context.Context, h *jwsapi.Handle, room uint64, feed uint64) *Subscriber {

	s := &Subscriber{
		ctx:    ctx,
		handle: h,
		room:   room,
		feed:   feed,
	}
	return s
}

//Room return room id
func (s *Subscriber) Room() uint64 {
	return s.room
}

//Feed return feed is
func (s *Subscriber) Feed() uint64 {
	return s.feed
}

//Handle return handle
func (s *Subscriber) Handle() *jwsapi.Handle {
	return s.handle
}

//Join join the janus, to subscriber feed
//audio,video default is true, data default is false
//optional or default param can use jwsapi.WithMessageOption to setting
//other params see  https://jwsapi.conf.meetecho.com/docs/videoroom.html VideoRoom Subscribers join
//jwsapi.WithMessageOption("video",false) to ignore video stream
//return sdp(offer),nil, or "", err
func (s *Subscriber) Join(opts ...jwsapi.MessageOption) (string, error) {

	body := jwsapi.Message{
		jwsapi.AttrRequest: "join",
		"ptype":            UserTypeSubscriber.String(),
		"room":             s.room,
		"feed":             s.feed,
		"audio":            true,
		"video":            true,
		"data":             false,
	}

	for _, opt := range opts {
		opt(body)
	}

	rsp, err := s.handle.Message(body)
	if err != nil {
		return "", err
	}

	jsep, ok := rsp.SubMessage("jsep")
	if !ok {
		return "", errors.New("not jsep")
	}
	if jtype, ok := jsep.String("type"); !ok || jtype != "offer" {
		return "", errors.New("jsep type error")
	}
	sdp, ok := jsep.String("sdp")
	if !ok {
		return "", errors.New("not sdp")
	}
	return sdp, nil
}

//Start send answer to janus
func (s *Subscriber) Start(answer string, trickle bool) error {

	body := jwsapi.Message{
		jwsapi.AttrRequest: "start",
	}
	jsep := jwsapi.Message{
		"type":    "answer",
		"sdp":     answer,
		"trickle": trickle,
	}

	_, err := s.handle.JsepMessage(body, jsep)
	return err
}

//Pause stop recv audio,video stream from janus video
func (s *Subscriber) Pause() error {
	body := jwsapi.Message{
		jwsapi.AttrRequest: "pause",
	}

	_, err := s.handle.Message(body)
	return err
}

//Play after call Pause to start recv audio,video stream
func (s *Subscriber) Play() error {
	body := jwsapi.Message{
		jwsapi.AttrRequest: "start",
	}

	_, err := s.handle.Message(body)
	return err
}

//Configure configure ..
//jwsapi.WithMessageOption("video",false) to don't recv video from janus
//jwsapi.WithMessageOption("video",true) to recv video from janus
//see https://jwsapi.conf.meetecho.com/docs/videoroom.html configure
func (s *Subscriber) Configure(opts ...jwsapi.MessageOption) error {
	body := jwsapi.Message{
		jwsapi.AttrRequest: "configure",
	}
	for _, opt := range opts {
		opt(body)
	}
	_, err := s.handle.Message(body)
	return err
}

//Switch switch stream from feed to new feed
func (s *Subscriber) Switch(newFeed uint64, opts ...jwsapi.MessageOption) error {

	body := jwsapi.Message{
		jwsapi.AttrRequest: "switch",
		"feed":             newFeed,
	}

	for _, opt := range opts {
		opt(body)
	}

	_, err := s.handle.Message(body)
	if err == nil {
		s.feed = newFeed
	}

	return err
}

//Leave leave to subscriber
func (s *Subscriber) Leave() error {

	body := jwsapi.Message{
		jwsapi.AttrRequest: "leave",
	}
	_, err := s.handle.Message(body)
	return err
}
