package plugin

import (
	"MyStreamBot/globals"
	"MyStreamBot/plugin/contract"
	"MyStreamBot/services"
)

type Plugin = contract.Plugin
type Context = contract.Context

var PluginAPIVersion = contract.PluginAPIVersion

type BasePlugin struct{}

func (BasePlugin) Init(ctx *Context) error                          { return nil }
func (BasePlugin) Start() error                                     { return nil }
func (BasePlugin) Stop() error                                      { return nil }
func (BasePlugin) Actions() []services.LuaFunction                  { return nil }
func (BasePlugin) OnChat(msg *globals.MessageFromStream)            {}
func (BasePlugin) OnEvent(event *globals.Event)                     {}
