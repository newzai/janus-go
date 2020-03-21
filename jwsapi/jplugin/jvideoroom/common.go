package jvideoroom

import (
	"github.com/newzai/janus-go/jwsapi"
)

//UserType user type , publisher, subscriber
type UserType int

const (
	//UserTypePublisher publisher
	UserTypePublisher UserType = iota + 1
	//UserTypeSubscriber subscriber
	UserTypeSubscriber
)

func (ut UserType) String() string {
	switch ut {
	case UserTypePublisher:
		return "publisher"
	case UserTypeSubscriber:
		return "subscriber"
	default:
		return "n/a"
	}
}

//WithMessageOptionRoom set room id
func WithMessageOptionRoom(room uint64) jwsapi.MessageOption {
	return func(param jwsapi.Message) {
		param["room"] = room
	}
}

//WithMessageOptionPermanet set room permanent
func WithMessageOptionPermanet(permanent bool) jwsapi.MessageOption {
	return func(param jwsapi.Message) {
		param["permanent"] = permanent
	}
}

//WithMessageOptionDescription set room description
func WithMessageOptionDescription(description string) jwsapi.MessageOption {
	return func(param jwsapi.Message) {
		param["description"] = description
	}
}

//WithMessageOptionSecret set secret
func WithMessageOptionSecret(secret string) jwsapi.MessageOption {
	return func(param jwsapi.Message) {
		param["secret"] = secret
	}
}

//CreateRoom create room from janus-gateway videoroom plugin
//jwsapi.WithMessageOption("publishers",10) to set publishers parm
//see https://janus.conf.meetecho.com/docs/videoroom.html create room param
func CreateRoom(h *jwsapi.Handle, opts ...jwsapi.MessageOption) (uint64, error) {
	msg := jwsapi.Message{
		jwsapi.AttrRequest: "create",
	}

	for _, opt := range opts {
		opt(msg)
	}
	rsp, err := h.Request(msg)
	if err != nil {
		return 0, err
	}
	pluginData := rsp.PluginData()
	data := pluginData.Data()
	room, ok := data.Uint64("room")
	if !ok {
		return 0, data.PluginDataError()
	}
	return room, nil
}

//DestroyRoom destroy room
func DestroyRoom(h *jwsapi.Handle, room uint64, opts ...jwsapi.MessageOption) error {

	msg := jwsapi.Message{
		jwsapi.AttrRequest: "destroy",
		"room":             room,
	}
	for _, opt := range opts {
		opt(msg)
	}
	rsp, err := h.Request(msg)
	if err != nil {
		return err
	}

	return rsp.PluginDataError()
}

//Exists check room is exists
func Exists(h *jwsapi.Handle, room uint64) (bool, error) {
	msg := jwsapi.Message{
		jwsapi.AttrRequest: "exists",
		"room":             room,
	}

	rsp, err := h.Request(msg)
	if err != nil {
		return false, err
	}
	plugindata := rsp.PluginData()
	data := plugindata.Data()

	return data.Bool("exists"), nil
}

//List list all rooms in janus-gateway videoroom
func List(h *jwsapi.Handle) ([]Room, error) {

	msg := jwsapi.Message{
		jwsapi.AttrRequest: "list",
	}

	rsp, err := h.Request(msg)
	if err != nil {
		return nil, err
	}
	plugindata := rsp.PluginData()
	data := plugindata.Data()
	rooms := data.Array("list")
	if rooms == nil {
		rooms = data.Array("rooms")
	}
	jrooms := make([]Room, 0, len(rooms))
	for _, r := range rooms {
		rr := jwsapi.Message(r.(map[string]interface{}))
		jrooms = append(jrooms, Room{rr})
	}
	return jrooms, nil
}

//Listparticipants get all room publishers
func Listparticipants(h *jwsapi.Handle, room uint64) ([]Participant, error) {
	msg := jwsapi.Message{
		jwsapi.AttrRequest: "listparticipants",
		"room":             room,
	}

	rsp, err := h.Request(msg)
	if err != nil {
		return nil, err
	}

	pluginData := rsp.PluginData()
	data := pluginData.Data()

	arParts := data.Array("participants")
	parts := make([]Participant, 0, len(arParts))
	for _, part := range arParts {
		parts = append(parts, Participant{jwsapi.Message(part.(map[string]interface{}))})
	}

	return parts, nil
}
