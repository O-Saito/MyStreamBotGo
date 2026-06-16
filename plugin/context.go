package plugin

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
)

type Context struct {
	Name string
}

func (c *Context) EmitEvent(eventType string, data map[string]any, target ...string) {
	if len(target) > 0 {
		globals.LuaRequest <- globals.SocketMessage{
			Filter: target[0] + ".lua",
			Type:   eventType,
			Data:   data,
		}
	} else {
		globals.EventQueue <- globals.Event{Type: eventType, Data: data}
	}
}

func (c *Context) Logf(level helpers.Level, format string, a ...any) {
	helpers.Logf(level, "[PLUGIN:"+c.Name+"] "+format, a...)
}
