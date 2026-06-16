package plugin

import (
	"MyStreamBot/globals"
	"MyStreamBot/services"
)

type Plugin interface {
	Name() string
	Init(ctx *Context) error
	Start() error
	Stop() error
	Actions() []services.LuaFunction
	OnChat(msg *globals.MessageFromStream)
	OnEvent(event *globals.Event)
}

type BasePlugin struct{}

func (BasePlugin) Init(ctx *Context) error                          { return nil }
func (BasePlugin) Start() error                                     { return nil }
func (BasePlugin) Stop() error                                      { return nil }
func (BasePlugin) Actions() []services.LuaFunction                  { return nil }
func (BasePlugin) OnChat(msg *globals.MessageFromStream)            {}
func (BasePlugin) OnEvent(event *globals.Event)                     {}
