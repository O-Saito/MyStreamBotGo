package plugin

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/services"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

var (
	mu         sync.RWMutex
	registered = map[string]Plugin{}
)

func Register(p Plugin) {
	name := p.Name()
	if name == "" {
		helpers.Logf(helpers.ERROR, "[PLUGIN] Register: plugin with empty Name() rejected")
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if _, exists := registered[name]; exists {
		helpers.Logf(helpers.WARN, "[PLUGIN] Register: %s already registered, overwriting", name)
	}
	registered[name] = p
}

func InitAll() {
	mu.RLock()
	plugins := copyPlugins()
	mu.RUnlock()

	for _, p := range plugins {
		ctx := &Context{Name: p.Name()}
		if err := p.Init(ctx); err != nil {
			helpers.Logf(helpers.ERROR, "[PLUGIN] Init %s: %v", p.Name(), err)
		} else {
			helpers.Logf(helpers.DEBUG, "[PLUGIN] Init %s: ok", p.Name())
		}
	}
}

func StartAll() {
	mu.RLock()
	plugins := copyPlugins()
	mu.RUnlock()

	for _, p := range plugins {
		if err := p.Start(); err != nil {
			helpers.Logf(helpers.ERROR, "[PLUGIN] Start %s: %v", p.Name(), err)
		} else {
			helpers.Logf(helpers.DEBUG, "[PLUGIN] Start %s: ok", p.Name())
		}
	}
}

func StopAll() {
	mu.RLock()
	plugins := copyPlugins()
	mu.RUnlock()

	for _, p := range plugins {
		if err := p.Stop(); err != nil {
			helpers.Logf(helpers.ERROR, "[PLUGIN] Stop %s: %v", p.Name(), err)
		} else {
			helpers.Logf(helpers.DEBUG, "[PLUGIN] Stop %s: ok", p.Name())
		}
	}
}

func DispatchChat(msg *globals.MessageFromStream) {
	mu.RLock()
	plugins := copyPlugins()
	mu.RUnlock()

	for _, p := range plugins {
		p.OnChat(msg)
	}
}

func DispatchEvent(event *globals.Event) {
	mu.RLock()
	plugins := copyPlugins()
	mu.RUnlock()

	for _, p := range plugins {
		p.OnEvent(event)
	}
}

func RegisterLuaActions(L *lua.LState) {
	mu.RLock()
	plugins := copyPlugins()
	mu.RUnlock()

	for _, p := range plugins {
		actions := p.Actions()
		if len(actions) == 0 {
			continue
		}
		services.ExposeToLua(L, "plugin_"+p.Name(), actions)
	}
}

func copyPlugins() []Plugin {
	out := make([]Plugin, 0, len(registered))
	for _, p := range registered {
		out = append(out, p)
	}
	return out
}
