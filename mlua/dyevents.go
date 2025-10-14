package mlua

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"MyStreamBot/helpers"

	lua "github.com/yuin/gopher-lua"
)

type DynamicEvent struct {
	Name       string
	Path       string
	LState     *lua.LState
	OnStart    *lua.LFunction
	OnTick     *lua.LFunction
	OnEvent    *lua.LFunction
	OnMessage  *lua.LFunction
	ModuleData map[string]any

	LastTick time.Time
	NextTick time.Time
	Interval time.Duration
	Paused   bool
	mu       sync.RWMutex
}

var (
	globalRegister     = make([]func(*lua.LState), 0)
	dynamicEvents      = make(map[string]*DynamicEvent)
	dynamicEventsMutex sync.RWMutex

	globalLoopOnce sync.Once
	stopGlobalLoop chan struct{}
)

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
			Name:       name,
			Path:       fullPath,
			LState:     L,
			OnStart:    getGlobalFunction(L, "on_start"),
			OnTick:     getGlobalFunction(L, "on_tick"),
			OnEvent:    getGlobalFunction(L, "on_event"),
			OnMessage:  getGlobalFunction(L, "on_message"),
			ModuleData: make(map[string]any),
			Interval:   time.Second, // padrão
			NextTick:   time.Now().Add(time.Second),
		}

		setFunctionOnTable(ev, eventTable)

		// Preserva estado se já existia
		dynamicEventsMutex.Lock()
		if old, exists := dynamicEvents[name]; exists {
			ev.ModuleData = old.ModuleData
		}
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

	ev.LState.SetField(tbl, "setInterval", ev.LState.NewFunction(func(L *lua.LState) int {
		val := L.CheckNumber(1)
		ev.Interval = time.Duration(float64(val) * float64(time.Second))
		return 0
	}))

	ev.LState.SetField(tbl, "setPaused", ev.LState.NewFunction(func(L *lua.LState) int {
		val := L.CheckBool(1)
		ev.Paused = val
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
				event := ev.LState.GetGlobal("ev")
				shouldRun := false
				intervalVal := ev.Interval
				if tbl, ok := event.(*lua.LTable); ok {
					interval := ev.LState.GetField(tbl, "interval")
					paused := ev.LState.GetField(tbl, "paused")

					if num, ok := interval.(lua.LNumber); ok {
						intervalVal = time.Second * time.Duration(float64(num))
					}

					pausedVal := false
					if b, ok := paused.(lua.LBool); ok {
						pausedVal = bool(b)
					}

					shouldRun = !pausedVal && ev.OnTick != nil && now.After(ev.NextTick)
				}
				ev.mu.RUnlock()

				if !shouldRun {
					continue
				}

				ev.mu.Lock()
				if ev.OnTick != nil {
					err := ev.LState.CallByParam(lua.P{
						Fn:      ev.OnTick,
						NRet:    0,
						Protect: true,
					})
					if err != nil {
						helpers.Logf(helpers.Red, "[DYNAMIC] Erro no on_tick de %s: %v", ev.Name, err)
					}
				}
				ev.LastTick = now
				ev.NextTick = now.Add(intervalVal)
				ev.mu.Unlock()
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
