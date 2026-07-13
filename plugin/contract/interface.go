package contract

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/services"
)

// 0xMMMmmmmpp where M=major, m=minor, p=patch
// current: 0.2.0
const PluginAPIVersion uint32 = 0x00020000

type Plugin interface {
	Name() string
	Init(ctx *Context) error
	Start() error
	Stop() error
	Actions() []services.LuaFunction
	OnChat(msg *globals.MessageFromStream)
	OnEvent(event *globals.Event)
}

type Context struct {
	Name string
}

func (c *Context) EmitEvent(eventType string, data map[string]any, target ...string) {
	if len(target) > 0 {
		globals.SafeSend(globals.LuaRequest, globals.SocketMessage{
			Filter: target[0] + ".lua",
			Type:   eventType,
			Data:   data,
		}, "LuaRequest")
	} else {
		globals.SafeSend(globals.EventQueue, globals.Event{Type: eventType, Data: data}, "EventQueue")
	}
}

func (c *Context) Logf(level helpers.Level, format string, a ...any) {
	helpers.Logf(level, "[PLUGIN:"+c.Name+"] "+format, a...)
}
