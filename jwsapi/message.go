package jwsapi

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

const (
	attrType        = "janus"
	attrTransaction = "transaction"
	attrSessionID   = "session_id"
	attrHandleID    = "handle_id"
	attrSender      = "sender"
	attrBody        = "body"
	attrPluginData  = "plugindata"
	attrPlugin      = "plugin"
	//AttrRequest request
	AttrRequest   = "request"
	attrVideoRoom = "videoroom"
	attrJSEP      = "jsep"
)

//Message janus message
type Message map[string]interface{}

//Type return janus message type
func (m *Message) Type() string {

	return (*m)[attrType].(string)
}

//Transaction get transaction from message
func (m *Message) Transaction() (string, bool) {

	t, ok := (*m)[attrTransaction]
	if ok {
		return t.(string), true
	}
	return "", false
}

//SubMessage get sub Message
func (m *Message) SubMessage(key string) (Message, bool) {
	value, ok := (*m)[key]
	if !ok {
		return nil, false
	}
	switch value.(type) {
	case map[string]interface{}:
		return Message(value.(map[string]interface{})), true
	default:
		return nil, false
	}
}

//Array get array value
func (m *Message) Array(key string) []interface{} {
	value, ok := (*m)[key]
	if !ok {
		return nil
	}
	switch value.(type) {
	case []interface{}:
		return value.([]interface{})
	default:
		return nil
	}

}

//Bool get bool value
func (m *Message) Bool(key string) bool {
	value, ok := (*m)[key]
	if !ok {
		return false
	}
	switch value.(type) {
	case bool:
		return value.(bool)
	default:
		return false
	}
}

func (m *Message) String(key string) (string, bool) {
	value, ok := (*m)[key]
	if !ok {
		return "", false
	}

	switch value.(type) {
	case string:
		return value.(string), true
	default:
		return "", false
	}
}

//Uint64 get value
func (m *Message) Uint64(key string) (uint64, bool) {
	value, ok := (*m)[key]
	if ok {
		switch val := value.(type) {
		case json.Number:
			ival, err := val.Int64()
			if err == nil {
				return uint64(ival), true
			}
		}
	}

	return 0, false
}

//Uint16 get value
func (m *Message) Uint16(key string) (uint16, bool) {
	value, ok := m.Uint64(key)
	return uint16(value), ok
}

//Uint32 get value
func (m *Message) Uint32(key string) (uint32, bool) {
	value, ok := m.Uint64(key)
	return uint32(value), ok
}

//SessionID get session id from handle
func (m *Message) SessionID() (uint64, bool) {
	return m.Uint64(attrSessionID)
}

//HandleID get handle id
func (m *Message) HandleID() (uint64, bool) {
	return m.Uint64(attrSender)
}

//IsACK is ack
func (m *Message) IsACK() bool {
	return strings.EqualFold(m.Type(), "ack")
}

//IsSuccess is success
func (m *Message) IsSuccess() bool {
	return strings.EqualFold(m.Type(), "success")
}

//IsEvent is event
func (m *Message) IsEvent() bool {
	return strings.EqualFold(m.Type(), "event")
}

//IsError check is error response
func (m *Message) IsError() bool {
	return strings.EqualFold(m.Type(), "error")
}

//Error IsError() is true ,call this
func (m *Message) Error() error {
	if m.IsError() {
		errInfo := (*m)["error"].(map[string]interface{})
		reason := errInfo["reason"].(string)
		return errors.New(reason)
	} else if pluginData, ok := m.SubMessage(attrPluginData); ok {
		return pluginData.PluginDataError()
	}
	return nil
}

//PluginDataError get plugindata error
func (m *Message) PluginDataError() error {
	if err, ok := (*m)["error"]; ok {
		switch err.(type) {
		case string:
			return errors.New(err.(string))
		case map[string]interface{}:
			errInfo := (*m)["error"].(map[string]interface{})
			reason := errInfo["reason"].(string)
			return errors.New(reason)
		default:
			return errors.New("unkwon error")
		}

	} else if data, ok := m.SubMessage("data"); ok {
		return data.PluginDataError()
	}
	return nil
}

//Data return data
func (m *Message) Data() Message {
	return Message((*m)["data"].(map[string]interface{}))
}

//Plugin get plugin...
func (m *Message) Plugin() string {
	return (*m)[attrPlugin].(string)
}

//PluginData get plugindata
func (m *Message) PluginData() Message {
	return Message((*m)[attrPluginData].(map[string]interface{}))
}

//VideoRoom get videoroom event type
func (m *Message) VideoRoom() string {
	return (*m)[attrVideoRoom].(string)
}

//Set set key value
func (m *Message) Set(key string, value interface{}) {
	(*m)[key] = value
}

//MessageOption  message option
type MessageOption func(Message)

//WithMessageOption set key value
func WithMessageOption(key string, value interface{}) MessageOption {
	return func(msg Message) {
		msg[key] = value
	}
}
