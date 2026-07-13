package loader

/*
#include <stdlib.h>
#include "hostapi.h"

extern bot_host_api_t* create_host_api(void);
*/
import "C"
import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"MyStreamBot/plugin/contract"
	"MyStreamBot/services"
	"encoding/json"
	"fmt"
	"runtime"
	"unsafe"

	lua "github.com/yuin/gopher-lua"
)

type sharedLibPlugin struct {
	path    string
	name    string
	lib     *LoadedLibrary
	hostAPI *C.bot_host_api_t
}

var _ contract.Plugin = (*sharedLibPlugin)(nil)

func LoadLibraryAsPlugin(lib *LoadedLibrary) (contract.Plugin, error) {
	name := CallPluginName(lib.Funcs["plugin_name"], lib.Funcs["plugin_free"])
	if name == "" {
		closeLibrary(lib.Handle)
		return nil, fmt.Errorf("%s: plugin_name returned empty", lib.Path)
	}

	ver := CallPluginAPIVersion(lib.Funcs["plugin_api_version"])
	if ver != contract.PluginAPIVersion {
		closeLibrary(lib.Handle)
		return nil, fmt.Errorf("%s: API version mismatch (plugin=%08x, host=%08x)",
			lib.Path, ver, contract.PluginAPIVersion)
	}

	api := C.create_host_api()
	if api == nil {
		closeLibrary(lib.Handle)
		return nil, fmt.Errorf("%s: failed to create host API", lib.Path)
	}
	CallPluginSetHost(lib.Funcs["plugin_set_host"], unsafe.Pointer(api))

	helpers.Logf(helpers.DEBUG, "[LOADER] loaded plugin %q from %s", name, lib.Path)
	return &sharedLibPlugin{path: lib.Path, name: name, lib: lib, hostAPI: api}, nil
}

func (p *sharedLibPlugin) Name() string { return p.name }

func (p *sharedLibPlugin) Init(ctx *contract.Context) error {
	fn := p.lib.Funcs["plugin_init"]
	if fn == 0 {
		return nil
	}
	data, _ := json.Marshal(map[string]string{"name": ctx.Name})
	return CallPluginInit(fn, string(data), p.lib.Funcs["plugin_free"])
}

func (p *sharedLibPlugin) Start() error {
	if p.lib.Funcs["plugin_start"] != 0 {
		CallPluginStart(p.lib.Funcs["plugin_start"])
	}
	return nil
}

func (p *sharedLibPlugin) Stop() error {
	if p.lib.Funcs["plugin_stop"] != 0 {
		CallPluginStop(p.lib.Funcs["plugin_stop"])
	}
	return nil
}

func (p *sharedLibPlugin) OnChat(msg *globals.MessageFromStream) {
	fn := p.lib.Funcs["plugin_on_chat"]
	if fn == 0 {
		return
	}
	data, _ := json.Marshal(msg)
	CallPluginOnChat(fn, string(data))
}

func (p *sharedLibPlugin) OnEvent(event *globals.Event) {
	fn := p.lib.Funcs["plugin_on_event"]
	if fn == 0 {
		return
	}
	data, _ := json.Marshal(event)
	CallPluginOnEvent(fn, string(data))
}

func (p *sharedLibPlugin) Actions() []services.LuaFunction {
	fn := p.lib.Funcs["plugin_get_actions"]
	if fn == 0 {
		return nil
	}
	raw := CallPluginGetActions(fn, p.lib.Funcs["plugin_free"])
	if raw == "" {
		return nil
	}

	var descriptors []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(raw), &descriptors); err != nil {
		helpers.Logf(helpers.ERROR, "[LOADER] %s: plugin_get_actions parse error: %v", p.name, err)
		return nil
	}

	callFn := p.lib.Funcs["plugin_call_action"]
	if callFn == 0 {
		return nil
	}

	freeFn := p.lib.Funcs["plugin_free"]

	funcs := make([]services.LuaFunction, 0, len(descriptors))
	for _, d := range descriptors {
		localActionName := d.Name
		funcs = append(funcs, services.LuaFunction{
			Name: localActionName,
			Fn: func(L *lua.LState) (n int) {
				defer func() {
					if r := recover(); r != nil {
						buf := make([]byte, 4096)
						stackLen := runtime.Stack(buf, false)
						helpers.Logf(helpers.ERROR, "[PLUGIN] %s action %s panicked: %v\n%s", p.name, localActionName, r, buf[:stackLen])
						L.Push(lua.LNil)
						L.Push(lua.LString(fmt.Sprintf("plugin %s action %s panicked", p.name, localActionName)))
						n = 2
					}
				}()
				meta := "{}"
				if mt := L.GetGlobal("__plugin_meta"); mt != nil && mt.Type() == lua.LTTable {
					if m, ok := mlua.TableToMap(mt.(*lua.LTable)).(map[string]interface{}); ok && len(m) > 0 {
						b, _ := json.Marshal(m)
						meta = string(b)
					}
				}
				var jsonArgs string
				if L.GetTop() > 0 {
					data, _ := json.Marshal(mlua.FromLValue(L, L.Get(1)))
					jsonArgs = string(data)
				}
				result := CallPluginAction(callFn, localActionName, jsonArgs, meta, freeFn)
				if result == "" {
					return 0
				}
				var parsed any
				if err := json.Unmarshal([]byte(result), &parsed); err == nil {
					L.Push(mlua.ToLValue(L, parsed))
				} else {
					L.Push(lua.LString(result))
				}
				return 1
			},
		})
	}
	return funcs
}

func (p *sharedLibPlugin) CloseLibrary() {
	if p.hostAPI != nil {
		deleteHostAPI(p.hostAPI)
		p.hostAPI = nil
	}
	if p.lib != nil {
		CloseLibrary(p.lib)
	}
}

func (p *sharedLibPlugin) GetPluginName() string {
	return p.name
}
