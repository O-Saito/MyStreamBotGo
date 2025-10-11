package mlua

import (
	"io/fs"
	"log"
	"path/filepath"
	"time"

	"MyStreamBot/globals"
	"MyStreamBot/helpers"

	"github.com/fsnotify/fsnotify"
	lua "github.com/yuin/gopher-lua"
)

type LuaModule struct {
	Path string
	Name string
}

func (m *LuaModule) NameWithoutExt() string {
	return m.Name[:len(m.Name)-len(filepath.Ext(m.Name))]
}

var (
	LChat     *lua.LState
	LCommands *lua.LState
	LEvents   *lua.LState

	commandFunctions = make(map[string]*lua.LFunction)
	chatFunctions    = make(map[string]*lua.LFunction)
	eventFunctions   = make(map[string]*lua.LFunction)

	commands     = make(map[string]*LuaModule)
	chatModules  = make(map[string]*LuaModule)
	eventModules = make(map[string][]*LuaModule)

	watcher   *fsnotify.Watcher
	reloadDeb = make(map[string]time.Time)
)

// inicializa o LState
func Init(funcs ...func(*lua.LState)) {
	LChat = lua.NewState()
	LCommands = lua.NewState()
	LEvents = lua.NewState()

	RegisterGlobalState(LChat)
	RegisterGlobalState(LCommands)
	RegisterGlobalState(LEvents)

	for _, f := range funcs {
		f(LChat)
		f(LCommands)
		f(LEvents)
	}
}

func ExposeServiceToLua(L *lua.LState, name string, funcs map[string]func(*lua.LState) int) {
	tbl := L.NewTable()
	lgFuncs := make(map[string]lua.LGFunction, len(funcs))
	for k, v := range funcs {
		lgFuncs[k] = lua.LGFunction(v)
	}
	L.SetFuncs(tbl, lgFuncs)
	L.SetGlobal(name, tbl)
}

// Load/Reload all modules
func LoadAllModules() {
	loadDir(LCommands, "./modules/commands", "command")
	loadDir(LChat, "./modules/chat", "chat")
	loadEvents(LEvents, "./modules/events")
}

func loadDir(L *lua.LState, dir string, modType string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".lua" {
			return nil
		}
		loadModule(L, path, modType)
		return nil
	})
}

func loadEvents(L *lua.LState, baseDir string) {
	filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".lua" {
			eventName := filepath.Base(filepath.Dir(path))
			loadModule(L, path, "event:"+eventName)
		}
		return nil
	})
}

func loadModule(L *lua.LState, path string, modType string) {
	if t, ok := reloadDeb[path]; ok && time.Since(t) < 200*time.Millisecond {
		return
	}
	reloadDeb[path] = time.Now()

	/*// Limpar módulo antigo
	var modGlobalName string
	switch modType {
	case "command":
		modGlobalName = "cmd_" + filepath.Base(path[:len(path)-len(filepath.Ext(path))])
	case "chat":
		modGlobalName = "chat_" + filepath.Base(path[:len(path)-len(filepath.Ext(path))])
	default:
		if len(modType) > 6 && modType[:6] == "event:" {
			eventName := modType[6:]
			modGlobalName = "event_" + eventName + "_" + filepath.Base(path[:len(path)-len(filepath.Ext(path))])
		}
	}

	if modGlobalName != "" {
		L.SetGlobal(modGlobalName, lua.LNil) // remove tabela antiga
	}*/

	/*if err := L.DoFile(path); err != nil {
		helpers.Logf(helpers.Red, "[LOAD ERROR] %s: %v", path, err)
		return
	}*/

	fn, err := L.LoadFile(path)
	if err != nil {
		helpers.Logf(helpers.Red, "[LOAD ERROR] %s: %v", path, err)
		return
	}

	if err := L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}); err != nil {
		helpers.Logf(helpers.Red, "[EXECUTE MODULE ERROR] %s: %v", path, err)
		return
	}

	mod := &LuaModule{Path: path, Name: filepath.Base(path)}

	switch {
	case modType == "command":
		f := L.GetGlobal("on_command")
		if fn, ok := f.(*lua.LFunction); ok {
			commands[mod.NameWithoutExt()] = mod
			commandFunctions[mod.NameWithoutExt()] = fn
		}
		//commandFunctions[mod.NameWithoutExt()] = fn
	case modType == "chat":
		f := L.GetGlobal("on_message")
		if fn, ok := f.(*lua.LFunction); ok {
			chatModules[mod.NameWithoutExt()] = mod
			chatFunctions[mod.NameWithoutExt()] = fn
		}
		//chatFunctions[mod.NameWithoutExt()] = fn
	case len(modType) > 6 && modType[:6] == "event:":
		eventName := modType[6:]
		eventModules[eventName] = append(eventModules[eventName], mod)
		eventFunctions[eventName+"_"+mod.NameWithoutExt()] = fn
	}
	/*switch {
	case modType == "command":
		commands[mod.NameWithoutExt()] = mod
	case modType == "chat":
		chatModules = append(chatModules, mod)
	case len(modType) > 6 && modType[:6] == "event:":
		eventName := modType[6:]
		eventModules[eventName] = append(eventModules[eventName], mod)
	}*/

	helpers.Logf(helpers.Green, "[MODULE LOADED] %s (%s)", path, modType)
}

func RegisterGlobalState(L *lua.LState) {
	mt := L.NewTypeMetatable("State")

	// __index → getters e métodos
	L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
		ud := L.CheckUserData(1)
		key := L.CheckString(2)

		state := ud.Value.(*globals.State)

		switch key {
		case "GetViewers":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				state.RLock()
				defer state.RUnlock()
				tbl := L.NewTable()
				for _, v := range state.GetViewerList() {
					tbl.Append(lua.LString(v))
				}
				L.Push(tbl)
				return 1
			}))
		case "AddViewer":
			L.Push(L.NewFunction(func(L *lua.LState) int {
				state.AddTwitchViewer(L.CheckString(1))
				return 0
			}))
		case "Data":
			state.RLock()
			defer state.RUnlock()
			tbl := L.NewTable()
			for k, v := range state.Data {
				switch val := v.(type) {
				case string:
					tbl.RawSetString(k, lua.LString(val))
				case int:
					tbl.RawSetString(k, lua.LNumber(val))
				case float64:
					tbl.RawSetString(k, lua.LNumber(val))
				case bool:
					tbl.RawSetString(k, lua.LBool(val))
				default:
					tbl.RawSetString(k, lua.LNil)
				}
			}
			L.Push(tbl)
		default:
			L.Push(lua.LNil)
		}

		return 1
	}))

	// __newindex → não permite set direto
	L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
		// nada permitido
		ud := L.CheckUserData(1)
		key := L.CheckString(2)
		helpers.Logf(helpers.Yellow, "[LUA STATE WARNING] Tentativa de setar State.%s diretamente", key)

		state := ud.Value.(*globals.State)

		switch key {
		case "Data":
			val := L.CheckTable(3)
			state.Lock()
			defer state.Unlock()
			val.ForEach(func(k, v lua.LValue) {
				switch v.Type() {
				case lua.LTString:
					state.Data[k.String()] = v.String()
				case lua.LTNumber:
					state.Data[k.String()] = float64(v.(lua.LNumber))
				case lua.LTBool:
					state.Data[k.String()] = bool(v.(lua.LBool))
				default:
					state.Data[k.String()] = nil
				}
			})
		}
		return 0
	}))

	ud := L.NewUserData()
	ud.Value = globals.GetState()
	L.SetMetatable(ud, L.GetTypeMetatable("State"))
	L.SetGlobal("state", ud)
}

// Dispatch events to modules
func HandleCommand(name string, ev globals.LuaCommand) {
	if fn, ok := commandFunctions[name]; ok {
		tbl := ToLTableCommand(LCommands, ev)
		if err := LCommands.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, tbl); err != nil {
			helpers.Logf(helpers.Red, "[LUA COMMAND ERROR] %s: %v", name, err)
		}
		return
	}

	helpers.Logf(helpers.Yellow, "[LUA COMMAND WARNING] Comando sem handler: %s", name)
	/*if mod, ok := commands[name]; ok {
		f := LCommands.GetGlobal("on_command")

		if f.Type() == lua.LTFunction {
			tbl := ToLTableCommand(LCommands, ev)
			if err := LCommands.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, tbl); err != nil {
				helpers.Logf(helpers.Red, "[LUA COMMAND HANDLER ERROR] %s: %v", mod.Path, err)
			}
		}
	}*/
}

func HandleChat(ev globals.MessageFromStream) {
	for name, fn := range chatFunctions {
		tbl := ToLTable(LChat, ev)
		if err := LChat.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, tbl); err != nil {
			helpers.Logf(helpers.Red, "[LUA CHAT ERROR] %s: %v", name, err)
		}
	}
	/*for _, mod := range chatModules {
		f := LChat.GetGlobal("on_message")
		if f.Type() == lua.LTFunction {
			tbl := ToLTable(LChat, ev)
			if err := LChat.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, tbl); err != nil {
				helpers.Logf(helpers.Red, "[LUA CHAT HANDLER ERROR] %s: %v", mod.Path, err)
			}
		}
	}*/
}

func HandleEvent(eventName string, ev globals.LuaEvent) {
	for name, fn := range eventFunctions {
		if len(name) > len(eventName) && name[:len(eventName)] == eventName {
			tbl := ToLTableEvent(LEvents, ev)
			if err := LEvents.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, tbl); err != nil {
				helpers.Logf(helpers.Red, "[LUA EVENT ERROR] %s: %v", name, err)
			}
		}
	}
	/*if mods, ok := eventModules[eventName]; ok {
		for _, mod := range mods {
			f := LEvents.GetGlobal("on_event")
			if f.Type() == lua.LTFunction {
				tbl := ToLTableEvent(LEvents, ev)
				if err := LEvents.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, tbl); err != nil {
					helpers.Logf(helpers.Red, "[LUA EVENT HANDLER ERROR] %s: %v", mod.Path, err)
				}
			}
		}
	}*/
}

// Hotreload
func StartWatcher() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Ext(ev.Name) != ".lua" {
					continue
				}
				log.Printf("[FS EVENT] %s %s", ev.Name, ev.Op)
				time.Sleep(50 * time.Millisecond)
				LoadAllModules()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				helpers.Logf(helpers.Red, "[WATCHER ERROR] %v", err)
			}
		}
	}()

	filepath.WalkDir("./modules", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
}

func StartEventQueues() {
	go func() {
		for ev := range globals.ChatQueue {
			HandleChat(ev)
		}
	}()

	go func() {
		for ev := range globals.CommandQueue {
			HandleCommand(ev.Name, ev)
		}
	}()

	go func() {
		for ev := range globals.EventQueue {
			HandleEvent(ev.Data["event_name"].(string), ev)
		}
	}()
}
