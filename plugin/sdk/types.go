package sdk

import (
	"encoding/json"
)

// HostAPIVersion must match plugin/contract.PluginAPIVersion.
// Native plugins compiled against this version are rejected at load
// if they don't match the host's expected version.
const HostAPIVersion uint32 = 0x00020000

type Context struct {
	Name string `json:"name"`
}

type MessageFromStream struct {
	Source    string         `json:"source"`
	Channel   string         `json:"channel"`
	UserId    string         `json:"userId"`
	User      string         `json:"user"`
	MessageId string         `json:"messageId"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata"`
}

type Event struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type LuaAction struct {
	Name string `json:"name"`
}

type Plugin interface {
	Name() string
	Init(ctx *Context) error
	Start() error
	Stop() error
	OnChat(msg *MessageFromStream)
	OnEvent(evt *Event)
	Actions() []LuaAction
}

func MarshalContext(ctx *Context) string {
	data, _ := json.Marshal(ctx)
	return string(data)
}

func UnmarshalContext(jsonStr string) (*Context, error) {
	var ctx Context
	err := json.Unmarshal([]byte(jsonStr), &ctx)
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}

func MarshalMessage(msg *MessageFromStream) string {
	data, _ := json.Marshal(msg)
	return string(data)
}

func UnmarshalMessage(jsonStr string) (*MessageFromStream, error) {
	var msg MessageFromStream
	err := json.Unmarshal([]byte(jsonStr), &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func MarshalEvent(evt *Event) string {
	data, _ := json.Marshal(evt)
	return string(data)
}

func UnmarshalEvent(jsonStr string) (*Event, error) {
	var evt Event
	err := json.Unmarshal([]byte(jsonStr), &evt)
	if err != nil {
		return nil, err
	}
	return &evt, nil
}
