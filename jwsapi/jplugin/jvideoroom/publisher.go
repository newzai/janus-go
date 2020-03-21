package jvideoroom

import (
	"context"

	"github.com/newzai/janus-go/jwsapi"
	"github.com/newzai/janus-go/logging"
	"github.com/pkg/errors"
)

//PublisherState publisher state
type PublisherState int

const (
	//PublisherStateUnjoin un join or leave
	PublisherStateUnjoin PublisherState = iota
	//PublisherStateJoined has join, unpublished
	PublisherStateJoined
	//PublisherStatePublished published
	PublisherStatePublished
)

//Publisher publisher
type Publisher struct {
	ctx     context.Context
	handle  *jwsapi.Handle
	room    uint64
	id      uint64
	display string
	parts   map[uint64]Participant
	tasks   chan func(*Publisher)
	state   PublisherState
	//callback
	onNewPublisher func(Participant)
	onUnpublished  func(uint64)
	onLeaved       func(uint64)
	onStateChanged func(PublisherState)
}

//PublisherOption 参数化
type PublisherOption func(*Publisher)

//WithPublisherOptionID client set publisher id, optional
func WithPublisherOptionID(id uint64) PublisherOption {
	return func(p *Publisher) {
		p.id = id
	}
}

//WithPublisherOptionDisplay set display for this publisher, optional
func WithPublisherOptionDisplay(display string) PublisherOption {
	return func(p *Publisher) {
		p.display = display
	}
}

//WithPublisherOptionNewPublisher set other on newpublisher callback
func WithPublisherOptionNewPublisher(callback func(Participant)) PublisherOption {
	return func(p *Publisher) {
		p.onNewPublisher = callback
	}
}

//WithPublisherOptionUnpublished set other unpublished callback.
func WithPublisherOptionUnpublished(callback func(uint64)) PublisherOption {
	return func(p *Publisher) {
		p.onUnpublished = callback
	}
}

//WithPublisherOptionLeaved set other publisher leaved callback
func WithPublisherOptionLeaved(callback func(uint64)) PublisherOption {
	return func(p *Publisher) {
		p.onLeaved = callback
	}
}

//NewPublisher create new publisher
func NewPublisher(ctx context.Context, h *jwsapi.Handle, room uint64, opts ...PublisherOption) *Publisher {

	p := &Publisher{
		ctx:    ctx,
		handle: h,
		room:   room,
		parts:  make(map[uint64]Participant),
		tasks:  make(chan func(*Publisher), 128),
		state:  PublisherStateUnjoin,
	}

	for _, opt := range opts {
		opt(p)
	}

	go p.execLoop()

	return p
}

//ID get id
func (p *Publisher) ID() uint64 {
	return p.id
}

//Display get display
func (p *Publisher) Display() string {
	return p.display
}

//Room get room
func (p *Publisher) Room() uint64 {
	return p.room
}

//Handle return handle
func (p *Publisher) Handle() *jwsapi.Handle {
	return p.handle
}

//SetOption set param, callback, eg: WithPublisherOptionNewPublisher
func (p *Publisher) SetOption(opts ...PublisherOption) {
	for _, opt := range opts {
		opt(p)
	}
}

//Join join to janus
func (p *Publisher) Join(opts ...jwsapi.MessageOption) error {

	body := jwsapi.Message{
		jwsapi.AttrRequest: "join",
		"ptype":            UserTypePublisher.String(),
		"room":             p.room,
	}
	if p.id > 0 {
		body["id"] = p.id
	}
	if p.display != "" {
		body["display"] = p.display
	}
	for _, opt := range opts {
		opt(body)
	}

	rsp, err := p.handle.Message(body)
	if err != nil {
		return err
	}
	pluginData := rsp.PluginData()
	data := pluginData.Data()
	p.id, _ = data.Uint64("id")

	publishers := data.Array("publishers")
	p.onPublishers(publishers)

	p.state = PublisherStateJoined

	return nil
}

//Publish start offer
//return answer, error
func (p *Publisher) Publish(audio bool, video bool, data bool, offer string, trickle bool, opts ...jwsapi.MessageOption) (string, error) {
	body := jwsapi.Message{
		jwsapi.AttrRequest: "configure",
		"audio":            audio,
		"video":            video,
		"data":             data,
	}
	for _, opt := range opts {
		opt(body)
	}
	jsep := jwsapi.Message{
		"type":    "offer",
		"sdp":     offer,
		"trickle": trickle,
	}

	rsp, err := p.handle.JsepMessage(body, jsep)
	if err != nil {
		return "", err
	}
	jsep, ok := rsp.SubMessage("jsep")
	if !ok {
		return "", errors.New("not jsep")
	}
	if jtype, ok := jsep.String("type"); !ok || jtype != "answer" {
		return "", errors.New("jsep type error")
	}
	sdp, ok := jsep.String("sdp")
	if !ok {
		return "", errors.New("not sdp")
	}
	return sdp, nil

}

//Unpublish tell janus-gateway close PeerConnection
func (p *Publisher) Unpublish() error {

	body := jwsapi.Message{
		jwsapi.AttrRequest: "unpublish",
	}

	_, err := p.handle.Message(body)
	return err
}

//Leave leave the room
func (p *Publisher) Leave() error {
	body := jwsapi.Message{
		jwsapi.AttrRequest: "leave",
	}

	_, err := p.handle.Message(body)
	return err
}

func (p *Publisher) onPublishers(publishers []interface{}) {
	for _, pub := range publishers {
		part := Participant{jwsapi.Message(pub.(map[string]interface{}))}
		id := part.ID()
		if id > 0 {
			p.parts[id] = part
			if p.onNewPublisher != nil {
				p.onNewPublisher(part)
			}
		}

	}
}

func (p *Publisher) onPluginEvent(event jwsapi.Message) {

	if publishers := event.Array("publishers"); publishers != nil {
		p.onPublishers(publishers)
	} else if unpublished, ok := event.Uint64("unpublished"); ok {
		delete(p.parts, unpublished)
		if p.onUnpublished != nil {
			p.onUnpublished(unpublished)
		}

	} else if leaving, ok := event.Uint64("leaving"); ok {
		delete(p.parts, leaving)
		if p.onUnpublished != nil {
			p.onUnpublished(leaving)
		}
		if p.onLeaved != nil {
			p.onLeaved(leaving)
		}
	} else if leaving, ok := event.String("leaving"); ok && leaving == "ok" {
		//itself leave,use call Leave()
		p.state = PublisherStateUnjoin
		if p.onStateChanged != nil {
			p.onStateChanged(p.state)
		}
	}
}

func (p *Publisher) onEvent(event *jwsapi.Message) {

	logging.Infof("[%d] recv event %v", p.id, event)
	switch event.Type() {
	case "event":
		pluginData := event.PluginData()
		data := pluginData.Data()
		p.onPluginEvent(data)
		//被动挂断..
	default:
		logging.Warnf("[%d] unknown msg %s", p.id, event.Type())
	}
}
func (p *Publisher) execLoop() {

	defer func() {
		logging.Infof("[%d] Publisher End", p.id)
		close(p.tasks)
	}()
	for {
		select {
		case <-p.ctx.Done():
			return
		case event, ok := <-p.handle.Events:
			if !ok {
				return
			}

			p.onEvent(event)
		case t, ok := <-p.tasks:
			if !ok {
				return
			}
			t(p)
		}
	}
}
