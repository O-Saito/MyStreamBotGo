package mlua

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
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

type DyEventQueueData struct {
	Type              int
	InnerType         string
	MessageFromStream globals.MessageFromStream
	LuaCommand        globals.Command
	LuaEvent          globals.Event
	SocketMessage     globals.SocketMessage
}

const (
	DyEventChat    = 0
	DyEventCommand = 1
	DyEventEvent   = 2
	DyEventRequest = 3
)

const (
	folderPathCommands = "./modules/commands"
	folderPathChat     = "./modules/chat"
	folderPathEvents   = "./modules/events"
	folderPathCustom   = "./modules/customevents"
)

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

	DyEventQueue = make(chan DyEventQueueData, 1000)
)

// initialize the LState
func Init(funcs ...func(*lua.LState)) {
	LChat = lua.NewState()
	LCommands = lua.NewState()
	LEvents = lua.NewState()

	RegisterGlobalState(LChat)
	RegisterGlobalState(LCommands)
	RegisterGlobalState(LEvents)

	dynamicEventsMutex.RLock()
	for _, f := range funcs {
		f(LChat)
		f(LCommands)
		f(LEvents)
		globalRegister = append(globalRegister, f)
	}
	dynamicEventsMutex.RUnlock()
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

func createIfNotExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeAppend)
	}
}

// Load/Reload all modules
func LoadAllModules() {
	createIfNotExists(folderPathCommands)
	createIfNotExists(folderPathChat)
	createIfNotExists(folderPathEvents)
	createIfNotExists(folderPathCustom)

	loadDir(LCommands, folderPathCommands, "command")
	loadDir(LChat, folderPathChat, "chat")
	loadEvents(LEvents, folderPathEvents)
	LoadDyEvents(folderPathCustom)
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

	fn, err := L.LoadFile(path)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[LOAD ERROR] %s: %v", path, err)
		return
	}

	if err := L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}); err != nil {
		helpers.Logf(helpers.ERROR, "[EXECUTE MODULE ERROR] %s: %v", path, err)
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
	case modType == "chat":
		f := L.GetGlobal("on_message")
		if fn, ok := f.(*lua.LFunction); ok {
			chatModules[mod.NameWithoutExt()] = mod
			chatFunctions[mod.NameWithoutExt()] = fn
		}
	case len(modType) > 6 && modType[:6] == "event:":
		eventName := modType[6:]
		f := L.GetGlobal("on_event")
		if fn, ok := f.(*lua.LFunction); ok {
			eventModules[eventName] = append(eventModules[eventName], mod)
			eventFunctions[eventName+"_"+mod.NameWithoutExt()] = fn
		}
	}
	helpers.Printf(helpers.Green, "[MODULE LOADED] %s (%s)", path, modType)
}

func RegisterGlobalState(L *lua.LState) {
	mt := L.NewTypeMetatable("State")

	// __index → getters and methods
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

	// __newindex → does not allow direct set
	L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
		// nothing allowed
		ud := L.CheckUserData(1)
		key := L.CheckString(2)
		helpers.Logf(helpers.WARN, "[LUA STATE WARNING] Attempt to set State.%s directly", key)

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
func HandleCommand(name string, ev *globals.Command) {
	tbl := LCommands.NewTable()
	tbl = ToLTableCommand(LCommands, ev, tbl)
	if fn, ok := commandFunctions[name]; ok {
		if err := LCommands.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, tbl); err != nil {
			helpers.Logf(helpers.ERROR, "[LUA COMMAND ERROR] %s: %v", name, err)
		}
	}
}

func HandleChat(ev *globals.MessageFromStream) {
	tbl := LChat.NewTable()
	tbl = ToLTable(LChat, ev, tbl)
	for name, fn := range chatFunctions {
		if err := LChat.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, tbl); err != nil {
			helpers.Logf(helpers.ERROR, "[LUA CHAT ERROR] %s: %v", name, err)
		}
	}
}

func HandleEvent(eventName string, ev *globals.Event) {
	tbl := ToLValue(LEvents, ev.Data)

	for name, fn := range eventFunctions {
		if strings.Contains(name, eventName) {
			if err := LEvents.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, tbl); err != nil {
				helpers.Logf(helpers.ERROR, "[LUA EVENT ERROR] %s: %v", name, err)
			}
		}
	}
}

// Hotreload
func StartWatcher() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		helpers.Log(helpers.INFO, "Started file watcher!")
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Ext(ev.Name) != ".lua" {
					continue
				}
				folder := strings.ReplaceAll(fmt.Sprintf(".\\%s", filepath.Dir(ev.Name)), "\\", "/")
				log.Printf("[FS EVENT] %s %s %s", folder, ev.Name, ev.Op)
				time.Sleep(50 * time.Millisecond)

				if strings.Contains(folder, folderPathCustom) {
					LoadDyEventModule(folder, strings.Replace(ev.Name, fmt.Sprintf("%s\\", filepath.Dir(ev.Name)), "", 1))
					continue
				}
				if strings.Contains(folder, folderPathCommands) {
					loadModule(LCommands, ev.Name, "command")
					continue
				}
				if strings.Contains(folder, folderPathChat) {
					loadModule(LChat, ev.Name, "chat")
					continue
				}
				if strings.Contains(folder, folderPathEvents) {
					loadEvents(LEvents, folderPathEvents)
					continue
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				helpers.Logf(helpers.ERROR, "[WATCHER ERROR] %v", err)
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
