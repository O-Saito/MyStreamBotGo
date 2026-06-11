package keylistener

import (
	"MyStreamBot/helpers"
	"MyStreamBot/plugin"
	"MyStreamBot/services"
	"sync"
)

type Plugin struct {
	plugin.BasePlugin
	ctx         *plugin.Context
	subscribers map[string]struct{}
	mu          sync.RWMutex
}

func init() {
	plugin.Register(&Plugin{
		subscribers: make(map[string]struct{}),
	})
}

func (p *Plugin) Name() string { return "keylistener" }

func (p *Plugin) Init(ctx *plugin.Context) error {
	p.ctx = ctx
	ctx.Logf(helpers.DEBUG, "keylistener initialized")
	return nil
}

func (p *Plugin) Start() error {
	return startHook(p.onKeyEvent)
}

func (p *Plugin) Stop() error {
	stopHook()
	p.mu.Lock()
	p.subscribers = make(map[string]struct{})
	p.mu.Unlock()
	return nil
}

func (p *Plugin) Actions() []services.LuaFunction {
	return []services.LuaFunction{
		{
			Name: "listen",
			Fn: func(moduleName string) {
				if moduleName == "" {
					return
				}
				p.mu.Lock()
				p.subscribers[moduleName] = struct{}{}
				p.mu.Unlock()
				p.ctx.Logf(helpers.DEBUG, "%s subscribed to key events", moduleName)
			},
		},
		{
			Name: "stop_listen",
			Fn: func(moduleName string) {
				p.mu.Lock()
				delete(p.subscribers, moduleName)
				p.mu.Unlock()
				p.ctx.Logf(helpers.DEBUG, "%s unsubscribed from key events", moduleName)
			},
		},
	}
}

func (p *Plugin) onKeyEvent(vkCode uint32, flags uint32, wParam uintptr) {
	if wParam != WM_KEYDOWN && wParam != WM_SYSKEYDOWN {
		return
	}

	ev := buildKeyEvent(vkCode, wParam)

	evMap := map[string]any{
		"vk_code": ev.VkCode,
		"vk_name": ev.VkName,
		"key":     ev.Key,
		"is_down": ev.IsDown,
		"modifiers": map[string]any{
			"shift": ev.Modifiers["shift"],
			"ctrl":  ev.Modifiers["ctrl"],
			"alt":   ev.Modifiers["alt"],
			"caps":  ev.Modifiers["caps"],
		},
	}

	p.mu.RLock()
	modules := make([]string, 0, len(p.subscribers))
	for m := range p.subscribers {
		modules = append(modules, m)
	}
	p.mu.RUnlock()

	for _, module := range modules {
		p.ctx.EmitEvent("keypress", evMap, module)
	}
}
