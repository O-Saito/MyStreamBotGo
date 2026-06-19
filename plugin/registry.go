package plugin

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/plugin/loader"
	"MyStreamBot/services"
	"runtime"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

type entry struct {
	p       Plugin
	mu      sync.RWMutex
	stopped bool
}

var (
	mu         sync.RWMutex
	registered = map[string]*entry{}
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
	registered[name] = &entry{p: p}
}

func LoadPlugins(dir string) {
	libs, err := loader.LoadDirectory(dir)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[PLUGIN] LoadPlugins: %v", err)
	}
	for _, lib := range libs {
		p, err := loader.LoadLibraryAsPlugin(lib)
		if err != nil {
			helpers.Logf(helpers.WARN, "[PLUGIN] skipping %s: %v", lib.Path, err)
			continue
		}
		Register(p)
	}
}

func InitAll() {
	mu.RLock()
	entries := copyEntries()
	mu.RUnlock()

	for _, e := range entries {
		e.mu.RLock()
		p := e.p
		stopped := e.stopped
		e.mu.RUnlock()
		if stopped {
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					helpers.Logf(helpers.ERROR, "[PLUGIN] %s panicked in Init: %v\n%s", p.Name(), r, buf[:n])
					e.mu.Lock()
					e.stopped = true
					e.mu.Unlock()
				}
			}()
			ctx := &Context{Name: p.Name()}
			if err := p.Init(ctx); err != nil {
				helpers.Logf(helpers.ERROR, "[PLUGIN] Init %s: %v", p.Name(), err)
			} else {
				helpers.Logf(helpers.DEBUG, "[PLUGIN] Init %s: ok", p.Name())
			}
		}()
	}
}

func StartAll() {
	mu.RLock()
	entries := copyEntries()
	mu.RUnlock()

	for _, e := range entries {
		e.mu.RLock()
		p := e.p
		stopped := e.stopped
		e.mu.RUnlock()
		if stopped {
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					helpers.Logf(helpers.ERROR, "[PLUGIN] %s panicked in Start: %v\n%s", p.Name(), r, buf[:n])
					e.mu.Lock()
					e.stopped = true
					e.mu.Unlock()
				}
			}()
			if err := p.Start(); err != nil {
				helpers.Logf(helpers.ERROR, "[PLUGIN] Start %s: %v", p.Name(), err)
			} else {
				helpers.Logf(helpers.DEBUG, "[PLUGIN] Start %s: ok", p.Name())
			}
		}()
	}
}

type libraryCloser interface {
	CloseLibrary()
}

func StopAll() {
	mu.RLock()
	entries := copyEntries()
	mu.RUnlock()

	for _, e := range entries {
		e.mu.Lock()
		func() {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					helpers.Logf(helpers.ERROR, "[PLUGIN] %s panicked in Stop: %v\n%s", e.p.Name(), r, buf[:n])
				}
			}()
			if err := e.p.Stop(); err != nil {
				helpers.Logf(helpers.ERROR, "[PLUGIN] Stop %s: %v", e.p.Name(), err)
			} else {
				helpers.Logf(helpers.DEBUG, "[PLUGIN] Stop %s: ok", e.p.Name())
			}
		}()
		func() {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					helpers.Logf(helpers.ERROR, "[PLUGIN] %s panicked in CloseLibrary: %v\n%s", e.p.Name(), r, buf[:n])
				}
			}()
			if closer, ok := e.p.(libraryCloser); ok {
				closer.CloseLibrary()
			}
		}()
		e.stopped = true
		e.mu.Unlock()
	}
}

func DispatchChat(msg *globals.MessageFromStream) {
	mu.RLock()
	entries := copyEntries()
	mu.RUnlock()

	for _, e := range entries {
		e.mu.RLock()
		p := e.p
		stopped := e.stopped
		e.mu.RUnlock()
		if stopped {
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					helpers.Logf(helpers.ERROR, "[PLUGIN] %s panicked in OnChat: %v\n%s", p.Name(), r, buf[:n])
					e.mu.Lock()
					e.stopped = true
					e.mu.Unlock()
				}
			}()
			p.OnChat(msg)
		}()
	}
}

func DispatchEvent(event *globals.Event) {
	mu.RLock()
	entries := copyEntries()
	mu.RUnlock()

	for _, e := range entries {
		e.mu.RLock()
		p := e.p
		stopped := e.stopped
		e.mu.RUnlock()
		if stopped {
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					helpers.Logf(helpers.ERROR, "[PLUGIN] %s panicked in OnEvent: %v\n%s", p.Name(), r, buf[:n])
					e.mu.Lock()
					e.stopped = true
					e.mu.Unlock()
				}
			}()
			p.OnEvent(event)
		}()
	}
}

func RegisterLuaActions(L *lua.LState) {
	mu.RLock()
	entries := copyEntries()
	mu.RUnlock()

	for _, e := range entries {
		e.mu.RLock()
		p := e.p
		stopped := e.stopped
		e.mu.RUnlock()
		if stopped {
			continue
		}

		actions := p.Actions()
		if len(actions) == 0 {
			continue
		}
		services.ExposeToLua(L, "plugin_"+p.Name(), actions)
	}
}

func copyEntries() []*entry {
	out := make([]*entry, 0, len(registered))
	for _, e := range registered {
		out = append(out, e)
	}
	return out
}
