package mlua

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	msql "MyStreamBot/sql"

	lua "github.com/yuin/gopher-lua"
)

type DynamicEvent struct {
	Name      string
	Path      string
	LState    *lua.LState
	OnStart   *lua.LFunction
	OnTick    *lua.LFunction
	OnEvent   *lua.LFunction
	OnMessage *lua.LFunction
	OnCommand *lua.LFunction
	OnRequest *lua.LFunction
	LastTick  time.Time
	NextTick  time.Time
	Interval  time.Duration
	Paused    bool
	mu        sync.RWMutex
	stateMu   sync.RWMutex
	db        *sql.DB
}

type DynamicEventInfo struct {
	Name       string         `json:"name"`
	Paused     bool           `json:"paused"`
	Interval   time.Duration  `json:"interval"`
	ModuleData map[string]any `json:"moduleData"`
}

var (
	globalRegister     = make([]func(*lua.LState), 0)
	dynamicEvents      = make(map[string]*DynamicEvent)
	dynamicEventsMutex sync.RWMutex

	globalLoopOnce sync.Once
	stopGlobalLoop chan struct{}
)

func (dev *DynamicEvent) ProcessChat(evm *globals.MessageFromStream) {
	dev.stateMu.RLock()
	paused := dev.Paused
	dev.stateMu.RUnlock()

	dev.mu.RLock()
	LState := dev.LState
	f := dev.OnMessage
	dev.mu.RUnlock()

	if f == nil || paused {
		return
	}

	tbl := LState.NewTable()
	tbl = ToLTable(LState, evm, tbl)

	dev.mu.RLock()
	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, tbl); err != nil {
		helpers.Logf(helpers.Red, "[LUA CHAT ERROR] %s: %v", dev.Name, err)
	}
	dev.mu.RUnlock()
}

func (dev *DynamicEvent) ProcessCommand(evm *globals.LuaCommand) {
	dev.stateMu.RLock()
	paused := dev.Paused
	dev.stateMu.RUnlock()

	dev.mu.RLock()
	f := dev.OnCommand
	LState := dev.LState
	dev.mu.RUnlock()

	if f == nil || paused {
		return
	}

	lname := lua.LString(evm.Name)
	tbl := LState.NewTable()
	tbl = ToLTableCommand(LState, evm, tbl)

	dev.mu.RLock()
	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, lname, tbl); err != nil {
		helpers.Logf(helpers.Red, "[LUA COMMAND ERROR] %s: %v", dev.Name, err)
	}
	dev.mu.RUnlock()
}

func (dev *DynamicEvent) ProcessEvent(evm *globals.LuaEvent) {
	dev.stateMu.RLock()
	paused := dev.Paused
	dev.stateMu.RUnlock()

	dev.mu.RLock()
	LState := dev.LState
	f := dev.OnEvent
	dev.mu.RUnlock()

	if f == nil || paused {
		return
	}

	evName := lua.LString(evm.Type)
	ntbl := ToLValue(LState, evm.Data)

	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, evName, ntbl); err != nil {
		helpers.Logf(helpers.Red, "[LUA EVENT ERROR] %s: %v", dev.Name, err)
	}
}

func (dev *DynamicEvent) ProcessRequest(evm *globals.SocketMessage) {
	dev.mu.RLock()
	if dev.Name != evm.Filter || dev.OnRequest == nil {
		dev.mu.RUnlock()
		return
	}
	f := dev.OnRequest
	LState := dev.LState
	dev.mu.RUnlock()

	tbl := ToLValue(LState, evm.Data)
	dev.mu.RLock()
	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, lua.LString(evm.Type), tbl); err != nil {
		helpers.Logf(helpers.Red, "[LUA EVENT ERROR] %s: %v", dev.Name, err)
	}
	dev.mu.RUnlock()
}

func ListDynamicEvents() []DynamicEventInfo {
	dynamicEventsMutex.RLock()
	defer dynamicEventsMutex.RUnlock()
	events := make([]DynamicEventInfo, 0, len(dynamicEvents))
	for name, val := range dynamicEvents {
		data := FromLValue(val.LState, val.LState.GetGlobal("ev")).(map[string]any)
		d := map[string]any{}
		if data["data"] != nil {
			d = data["data"].(map[string]any)
		}
		events = append(events, DynamicEventInfo{
			Name:       name,
			Paused:     val.Paused,
			Interval:   val.Interval,
			ModuleData: d,
		})
	}

	return events
}

func UpdateDynamicEvent(event DynamicEventInfo) error {
	dynamicEventsMutex.Lock()
	defer dynamicEventsMutex.Unlock()
	if ev, exists := dynamicEvents[event.Name]; exists {
		ev.mu.Lock()
		ev.Paused = event.Paused
		ev.Interval = time.Duration(float64(event.Interval) * float64(time.Second))
		ev.mu.Unlock()
		return nil
	}
	return os.ErrNotExist
}

// Chamado de loadAllModules
func LoadDyEvents(baseDir string) {
	helpers.Logf(helpers.Lua, "[DYNAMIC] Carregando eventos dinâmicos de %s", baseDir)

	files, err := os.ReadDir(baseDir)
	if err != nil {
		helpers.Logf(helpers.Red, "[DYNAMIC] Erro ao ler diretório: %v", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".lua" {
			continue
		}
		fullPath := filepath.Join(baseDir, file.Name())
		name := file.Name()

		L := lua.NewState()
		L.DoString(fmt.Sprintf(`package.path = package.path .. ";%s/modules/?.lua;"`, baseDir))
		RegisterGlobalState(L)

		dynamicEventsMutex.RLock()
		for _, f := range globalRegister {
			f(L)
		}
		dynamicEventsMutex.RUnlock()

		eventTable := L.NewTable()
		L.SetGlobal("ev", eventTable)

		fn, err := L.LoadFile(fullPath)
		if err != nil {
			helpers.Logf(helpers.Red, "[DYNAMIC] Erro ao carregar %s: %v", name, err)
			continue
		}

		if err := L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}); err != nil {
			helpers.Logf(helpers.Red, "[DYNAMIC] Erro executando %s: %v", name, err)
			continue
		}

		ev := &DynamicEvent{
			Name:      name,
			Path:      fullPath,
			LState:    L,
			OnStart:   getGlobalFunction(L, "on_start"),
			OnTick:    getGlobalFunction(L, "on_tick"),
			OnEvent:   getGlobalFunction(L, "on_event"),
			OnMessage: getGlobalFunction(L, "on_message"),
			OnCommand: getGlobalFunction(L, "on_command"),
			OnRequest: getGlobalFunction(L, "on_request"),
			NextTick:  time.Now().Add(time.Second),
			Interval:  time.Second, // padrão
			Paused:    true,        // padrão
			db:        nil,
		}

		setFunctionOnTable(ev, eventTable)

		oldEvent := dynamicEvents[name]
		if oldEvent != nil {
			data := FromLValue(oldEvent.LState, oldEvent.LState.GetGlobal("ev")).(map[string]any)
			d := map[string]any{}
			if data["data"] != nil {
				d = data["data"].(map[string]any)
			}
			oldData := ToLValue(ev.LState, d)
			eventTable.RawSetString("data", oldData)
			ev.db = oldEvent.db
		}

		dynamicEventsMutex.Lock()
		dynamicEvents[name] = ev
		dynamicEventsMutex.Unlock()

		helpers.Logf(helpers.Green, "[DYNAMIC] Evento carregado: %s", name)

		// Executa on_start
		if ev.OnStart != nil {
			if err := L.CallByParam(lua.P{Fn: ev.OnStart, NRet: 0, Protect: true}); err != nil {
				helpers.Logf(helpers.Red, "[DYNAMIC] Erro no on_start de %s: %v", name, err)
			}
		}
	}

	// Inicia o loop global apenas uma vez
	globalLoopOnce.Do(func() {
		stopGlobalLoop = make(chan struct{})
		go globalEventLoop()
	})
}

func setFunctionOnTable(ev *DynamicEvent, tbl *lua.LTable) {
	if ev.LState == nil {
		return
	}

	ev.LState.SetField(tbl, "socket_send", ev.LState.NewFunction(func(L *lua.LState) int {
		if L.Get(1) == lua.LNil || L.Get(2) == lua.LNil {
			return -1
		}

		t := L.CheckString(1)
		d := L.CheckTable(2)

		globals.WsBroadcast <- globals.SocketMessage{
			Filter: ev.Name,
			Type:   t,
			Data:   TableToMap(d),
		}

		return 0
	}))

	ev.LState.SetField(tbl, "setInterval", ev.LState.NewFunction(func(L *lua.LState) int {
		val := L.CheckNumber(1)
		ev.stateMu.Lock()
		ev.Interval = time.Duration(float64(val) * float64(time.Second))
		ev.stateMu.Unlock()
		return 0
	}))

	ev.LState.SetField(tbl, "setPaused", ev.LState.NewFunction(func(L *lua.LState) int {
		val := L.CheckBool(1)
		ev.stateMu.Lock()
		ev.Paused = val
		ev.stateMu.Unlock()
		return 0
	}))

	ev.LState.SetField(tbl, "getInterval", ev.LState.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ev.Interval.Seconds()))
		return 1
	}))

	ev.LState.SetField(tbl, "isPaused", ev.LState.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LBool(ev.Paused))
		return 1
	}))

	ev.LState.SetField(tbl, "useDB", ev.LState.NewFunction(func(L *lua.LState) int {
		createdDb := false
		ev.stateMu.Lock()
		if ev.db == nil {
			db, _ := msql.OpenModuleDB(ev.Name)
			ev.db = db
			createdDb = true
		}
		ev.stateMu.Unlock()
		L.Push(lua.LBool(createdDb))
		return 1
	}))

	ev.LState.SetField(tbl, "db_exec", ev.LState.NewFunction(func(L *lua.LState) int {
		if L.Get(1) == lua.LNil {
			return 0
		}
		query := L.CheckString(1)
		ev.stateMu.RLock()
		if ev.db == nil {
			ev.stateMu.RUnlock()
			return 0
		}
		_, err := ev.db.Exec(query)
		ev.stateMu.RUnlock()
		if err != nil {
			helpers.Logf("[DYEVENTS DB_EXEC] %s", err.Error())
		}
		return 0
	}))

	ev.LState.SetField(tbl, "db_query", ev.LState.NewFunction(func(L *lua.LState) int {
		if L.Get(1) == lua.LNil {
			return 0
		}
		query := L.CheckString(1)
		ev.stateMu.RLock()
		if ev.db == nil {
			ev.stateMu.RUnlock()
			return 0
		}
		r, _ := ev.db.Query(query)
		ev.stateMu.RUnlock()
		cols, _ := r.Columns()
		helpers.Logf(helpers.Red, "COLS %v", cols)
		result := []map[string]any{}

		for r.Next() {
			d := map[string]any{}
			items := make([]any, len(cols))
			for i := range items {
				// http://go-database-sql.org/varcols.html
				items[i] = new(sql.RawBytes)
			}
			r.Scan(items...)

			for i, v := range cols {
				if sb, ok := items[i].(*sql.RawBytes); ok {
					d[v] = string(*sb)
				}
			}
			result = append(result, d)
		}
		r.Close()
		L.Push(ToLValue(ev.LState, result))
		return 1
	}))

}

// Loop único para todos os eventos
func globalEventLoop() {
	helpers.Logf(helpers.Lua, "[DYNAMIC] Loop global iniciado")
	ticker := time.NewTicker((1 * time.Second) / 60)
	defer ticker.Stop()

	for {
		select {
		case <-stopGlobalLoop:
			helpers.Logf(helpers.Lua, "[DYNAMIC] Loop global parado")
			return

		case now := <-ticker.C:
			dynamicEventsMutex.RLock()
			for _, ev := range dynamicEvents {
				ev.mu.RLock()
				if ev.Paused || ev.OnTick == nil {
					//helpers.Logf(helpers.Yellow, "[DYNAMIC] Evento %s está pausado ou sem on_tick", ev.Name)
					continue
				}
				shouldRun := now.After(ev.NextTick)
				ev.mu.RUnlock()

				if !shouldRun {
					continue
				}

				ev.mu.RLock()
				LState := ev.LState
				f := ev.OnTick
				ev.mu.RUnlock()
				if f != nil {
					err := LState.CallByParam(lua.P{
						Fn:      f,
						NRet:    0,
						Protect: true,
					})
					if err != nil {
						helpers.Logf(helpers.Red, "[DYNAMIC] Erro no on_tick de %s: %v", ev.Name, err)
					}
				}
				ev.stateMu.Lock()
				ev.LastTick = now
				ev.NextTick = now.Add(ev.Interval)
				ev.stateMu.Unlock()
			}
			dynamicEventsMutex.RUnlock()
		}
	}
}

// Permite eventos do websocket interno
func HandleDyEventWebsocket(msg any) {
	dynamicEventsMutex.RLock()
	defer dynamicEventsMutex.RUnlock()

	for _, ev := range dynamicEvents {
		if ev.OnEvent == nil { //|| ev.Paused {
			continue
		}

		ev.mu.RLock()
		tbl := ev.LState.NewTable()
		ev.LState.SetField(tbl, "payload", ToLValue(ev.LState, msg))
		if err := ev.LState.CallByParam(lua.P{
			Fn:      ev.OnEvent,
			NRet:    0,
			Protect: true,
		}, tbl); err != nil {
			helpers.Logf(helpers.Red, "[DYNAMIC] Erro no on_event de %s: %v", ev.Name, err)
		}
		ev.mu.RUnlock()
	}
}

func getGlobalFunction(L *lua.LState, name string) *lua.LFunction {
	f := L.GetGlobal(name)
	if fn, ok := f.(*lua.LFunction); ok {
		return fn
	}
	return nil
}
