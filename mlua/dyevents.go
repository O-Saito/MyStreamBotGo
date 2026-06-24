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
	mu         sync.RWMutex
	Name       string
	Path       string
	Reloading  bool
	Lua        DynamicEventLua
	State      DynamicEventState
	Statistics DynamicEventStats
}

type DynamicEventLua struct {
	mu        sync.RWMutex
	LState    *lua.LState
	OnStart   *lua.LFunction
	OnTick    *lua.LFunction
	OnEvent   *lua.LFunction
	OnMessage *lua.LFunction
	OnCommand *lua.LFunction
	OnRequest *lua.LFunction
}

type DynamicEventState struct {
	mu                      sync.RWMutex
	LastTick                time.Time
	NextTick                time.Time
	Paused                  bool
	ComputeTwitchSharedChat bool
	db                      *sql.DB
	Interval                time.Duration
}

type DynamicEventStats struct {
	mu                 sync.RWMutex
	ProcessedTimes     int
	ProcessedTotalTime time.Duration
	LastProcessedTime  time.Duration
	HighestTime        time.Duration
	LowestTime         time.Duration
}

type DynamicEventInfo struct {
	Name                    string         `json:"name"`
	Paused                  bool           `json:"paused"`
	Interval                time.Duration  `json:"interval"`
	ComputeTwitchSharedChat bool           `json:"computeTwitchSharedChat"`
	ModuleData              map[string]any `json:"moduleData"`
	ProcessedTimes          int            `json:"processedTimes"`
	ProcessedTotalTime      time.Duration  `json:"ProcessedTotalTimes"`
	LastProcessedTime       time.Duration  `json:"LastProcessedTime"`
	HighestTime             time.Duration  `json:"highestTime"`
	LowestTime              time.Duration  `json:"lowestTime"`
}

var (
	globalRegister     = make([]func(*lua.LState), 0)
	dynamicEvents      = make(map[string]*DynamicEvent)
	dynamicEventsMutex sync.RWMutex

	globalLoopOnce sync.Once
	stopGlobalLoop chan struct{}
)

func GetDyEvents() map[string]*DynamicEvent {
	dynamicEventsMutex.RLock()
	defer dynamicEventsMutex.RUnlock()
	return dynamicEvents
}

func StopGlobalLoop() {
	if stopGlobalLoop != nil {
		close(stopGlobalLoop)
	}
}

func (dev *DynamicEvent) Transfer(new *DynamicEvent) {
	dev.mu.Lock()
	defer new.mu.Unlock()
	new.mu.Lock()
	defer dev.mu.Unlock()

	dev.Reloading = true

	dev.Lua.mu.Lock()
	data := FromLValue(dev.Lua.LState, dev.Lua.LState.GetGlobal("ev")).(map[string]any)
	dev.Lua.mu.Unlock()

	d := map[string]any{}
	if data["data"] != nil {
		d = data["data"].(map[string]any)
	}

	new.Lua.mu.Lock()
	oldData := ToLValue(new.Lua.LState, d)
	evTable := new.Lua.LState.GetGlobal("ev").(*lua.LTable)
	evTable.RawSetString("data", oldData)
	new.Lua.mu.Unlock()

	new.State.mu.Lock()
	dev.State.mu.Lock()
	new.State.db = dev.State.db
	dev.State.db = nil
	dev.State.mu.Unlock()
	new.State.mu.Unlock()
}

func (dev *DynamicEvent) Close() {
	dev.State.mu.Lock()
	defer dev.State.mu.Unlock()

	if dev.State.db != nil {
		dev.State.db.Close()
		dev.State.db = nil
	}

	dev.Lua.mu.Lock()
	defer dev.Lua.mu.Unlock()
	if dev.Lua.LState != nil {
		dev.Lua.LState.Close()
		dev.Lua.LState = nil
	}
}

func (dev *DynamicEvent) CanProcess() bool {
	dev.mu.RLock()
	defer dev.mu.RUnlock()

	if dev.Reloading {
		return false
	}

	dev.State.mu.RLock()
	paused := dev.State.Paused
	dev.State.mu.RUnlock()

	if paused {
		return false
	}
	return true
}

func (dev *DynamicEvent) ProcessWebsocketEvent(msg any) {
	if !dev.CanProcess() {
		return
	}

	dev.Lua.mu.Lock()
	defer dev.Lua.mu.Unlock()
	LState := dev.Lua.LState
	f := dev.Lua.OnEvent

	if f == nil {
		return
	}

	tbl := LState.NewTable()
	LState.SetField(tbl, "payload", ToLValue(LState, msg))

	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, tbl); err != nil {
		helpers.Logf(helpers.ERROR, "[DYNAMIC] Error in on_event of %s: %v", dev.Name, err)
	}
}

func (dev *DynamicEvent) ProcessStart() {
	dev.mu.Lock()
	dev.Reloading = false
	dev.mu.Unlock()

	dev.Lua.mu.Lock()
	defer dev.Lua.mu.Unlock()

	LState := dev.Lua.LState
	f := dev.Lua.OnStart

	if f == nil {
		return
	}

	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}); err != nil {
		helpers.Logf(helpers.ERROR, "[DYNAMIC] Error in on_start of %s: %v", dev.Name, err)
	}
}

func (dev *DynamicEvent) ProcessOnTick() {
	dev.Lua.mu.Lock()
	defer dev.Lua.mu.Unlock()

	LState := dev.Lua.LState
	f := dev.Lua.OnTick

	if f != nil {
		if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}); err != nil {
			helpers.Logf(helpers.ERROR, "[DYNAMIC] Error inor in on_tick of %s: %v", dev.Name, err)
		}
	}
}

func (dev *DynamicEvent) ProcessChat(evm *globals.MessageFromStream) {
	if !dev.CanProcess() {
		return
	}

	dev.State.mu.RLock()
	shouldComputeSharedChat := dev.State.ComputeTwitchSharedChat
	dev.State.mu.RUnlock()

	if !shouldComputeSharedChat &&
		evm.Source == "twitch" &&
		evm.Channel != globals.GetState().TwitchUser.UserLogin {
		return
	}

	dev.Lua.mu.Lock()
	defer dev.Lua.mu.Unlock()

	LState := dev.Lua.LState
	f := dev.Lua.OnMessage

	if f == nil {
		return
	}

	tbl := LState.NewTable()
	tbl = ToLTable(LState, evm, tbl)

	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, tbl); err != nil {
		helpers.Logf(helpers.ERROR, "[LUA CHAT ERROR] %s: %v", dev.Name, err)
	}
}

func (dev *DynamicEvent) ProcessCommand(evm *globals.Command) {
	if !dev.CanProcess() {
		return
	}

	dev.State.mu.RLock()
	shouldComputeSharedChat := dev.State.ComputeTwitchSharedChat
	dev.State.mu.RUnlock()

	if !shouldComputeSharedChat &&
		evm.Source == "twitch" &&
		evm.Channel != globals.GetState().TwitchUser.UserLogin {
		return
	}

	dev.Lua.mu.Lock()
	defer dev.Lua.mu.Unlock()

	LState := dev.Lua.LState
	f := dev.Lua.OnCommand

	if f == nil {
		return
	}

	lname := lua.LString(evm.Name)
	tbl := LState.NewTable()
	tbl = ToLTableCommand(LState, evm, tbl)

	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, lname, tbl); err != nil {
		helpers.Logf(helpers.ERROR, "[LUA COMMAND ERROR] %s: %v", dev.Name, err)
	}
}

func (dev *DynamicEvent) ProcessEvent(evm *globals.Event) {
	if !dev.CanProcess() {
		return
	}

	dev.State.mu.RLock()
	shouldComputeSharedChat := dev.State.ComputeTwitchSharedChat
	dev.State.mu.RUnlock()

	if !shouldComputeSharedChat {
		if payload, ok := evm.Data["payload"].(map[string]any); ok {
			if eventData, ok := payload["event"].(map[string]any); ok {
				if broadcasterId, ok := eventData["broadcaster_user_id"].(string); ok {
					if broadcasterId != globals.GetState().TwitchUser.UserID {
						return
					}
				}
			}
		}
	}

	dev.Lua.mu.Lock()
	defer dev.Lua.mu.Unlock()

	LState := dev.Lua.LState
	f := dev.Lua.OnEvent

	if f == nil {
		return
	}

	evName := lua.LString(evm.Type)
	ntbl := ToLValue(LState, evm.Data)

	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, evName, ntbl); err != nil {
		helpers.Logf(helpers.ERROR, "[LUA EVENT ERROR] %s: %v", dev.Name, err)
	}
}

func (dev *DynamicEvent) ProcessRequest(evm *globals.SocketMessage) {
	if dev.Name != evm.Filter {
		return
	}

	if !dev.CanProcess() {
		return
	}

	dev.Lua.mu.Lock()
	defer dev.Lua.mu.Unlock()

	LState := dev.Lua.LState
	f := dev.Lua.OnRequest

	if f == nil {
		return
	}

	tbl := ToLValue(LState, evm.Data)
	if tbl == lua.LNil {
		tbl = LState.NewTable()
	}

	socketTag := evm.SocketTag
	responseMessageID := evm.ResponseMessageID
	filter := evm.Filter
	t := evm.Type
	LState.SetField(tbl, "respond", LState.NewFunction(func(L *lua.LState) int {
		if L.Get(1) == lua.LNil {
			helpers.Printf(helpers.Lua, "[LUA] OnRequest no data passed!")
			return -1
		}

		d := L.CheckTable(1)

		if socketTag == 0 || responseMessageID == "" {
			helpers.Printf(helpers.Lua, "[LUA] OnRequest nowhere to respond [%v:%v]", socketTag, responseMessageID)
			return -1
		}
		data := TableToMap(d)

		globals.SafeSend(globals.WsBroadcast, globals.SocketMessage{
			SocketTag:         socketTag,
			ResponseMessageID: responseMessageID,
			Filter:            filter,
			Type:              fmt.Sprintf("return-%s", t),
			Data:              data,
		}, "WsBroadcast")

		return 0
	}))

	if err := LState.CallByParam(lua.P{Fn: f, NRet: 0, Protect: true}, lua.LString(evm.Type), tbl); err != nil {
		helpers.Logf(helpers.ERROR, "[LUA EVENT ERROR] %s: %v", dev.Name, err)
	}
}

func ListDynamicEvents() []DynamicEventInfo {
	dynamicEventsMutex.RLock()
	defer dynamicEventsMutex.RUnlock()
	events := make([]DynamicEventInfo, 0, len(dynamicEvents))

	for name, val := range dynamicEvents {
		val.Lua.mu.RLock()
		ldata := FromLValue(val.Lua.LState, val.Lua.LState.GetGlobal("ev"))
		val.Lua.mu.RUnlock()

		data := map[string]any{}
		if ldata != nil {
			data = ldata.(map[string]any)
		}
		d := map[string]any{}
		if data["data"] != nil {
			d = data["data"].(map[string]any)
		}
		val.State.mu.RLock()
		val.Statistics.mu.RLock()
		events = append(events, DynamicEventInfo{
			Name:                    name,
			Paused:                  val.State.Paused,
			ComputeTwitchSharedChat: val.State.ComputeTwitchSharedChat,
			Interval:                val.State.Interval,
			ModuleData:              d,
			ProcessedTimes:          val.Statistics.ProcessedTimes,
			ProcessedTotalTime:      val.Statistics.ProcessedTotalTime,
			LastProcessedTime:       val.Statistics.LastProcessedTime,
			HighestTime:             val.Statistics.HighestTime,
			LowestTime:              val.Statistics.LowestTime,
		})
		val.State.mu.RUnlock()
		val.Statistics.mu.RUnlock()
	}

	return events
}

func UpdateDynamicEvent(event *DynamicEventInfo) error {
	dynamicEventsMutex.Lock()
	defer dynamicEventsMutex.Unlock()
	if ev, exists := dynamicEvents[event.Name]; exists {
		ev.mu.RLock()
		defer ev.mu.RUnlock()
		ev.State.mu.Lock()
		defer ev.State.mu.Unlock()
		ev.State.Paused = event.Paused
		ev.State.ComputeTwitchSharedChat = event.ComputeTwitchSharedChat
		ev.State.Interval = time.Duration(float64(event.Interval) * float64(time.Second))
		return nil
	}
	return os.ErrNotExist
}

func SaveDynamicEvents() {
	dynamicEventsMutex.RLock()
	defer dynamicEventsMutex.RUnlock()
	for _, val := range dynamicEvents {
		val.State.mu.Lock()
		val.Lua.mu.Lock()
		if val.State.db != nil {
			val.State.db.Close()
		}
		if val.Lua.LState != nil {
			val.Lua.LState.Close()
		}
		val.State.mu.Unlock()
		val.Lua.mu.Unlock()
	}
}

func (des *DynamicEventStats) AddTiming(elapsed time.Duration) {
	des.mu.Lock()
	des.ProcessedTimes++
	des.ProcessedTotalTime += elapsed
	des.LastProcessedTime = elapsed
	if elapsed > des.HighestTime {
		des.HighestTime = elapsed
	}
	if des.LowestTime == 0 || elapsed < des.LowestTime {
		des.LowestTime = elapsed
	}
	des.mu.Unlock()
}

// Called from loadAllModules
func LoadDyEvents(baseDir string) {
	helpers.Logf(helpers.DEBUG, "[DYNAMIC] Loading dynamic events from %s", baseDir)

	files, err := os.ReadDir(baseDir)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[DYNAMIC] Error reading directory: %v", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".lua" {
			continue
		}
		name := file.Name()
		LoadDyEventModule(baseDir, name)
	}

	// Start global loop only once
	dynamicEventsMutex.RLock()
	defer dynamicEventsMutex.RUnlock()
	if len(dynamicEvents) > 0 {
		globalLoopOnce.Do(func() {
			stopGlobalLoop = make(chan struct{})
			go globalEventLoop()
		})
	}
}

func LoadDyEventModule(folder, fileName string) {
	fullPath := filepath.Join(folder, fileName)
	helpers.Logf(helpers.INFO, "LoadDyEventModule %s %s [%s]", folder, fileName, fullPath)
	L := lua.NewState()
	L.DoString(fmt.Sprintf(`package.path = package.path .. ";%s/modules/?.lua;"`, folder))
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
		helpers.Logf(helpers.ERROR, "[DYNAMIC] Error loading %s: %v", fileName, err)
		return
	}

	if err := L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}); err != nil {
		helpers.Logf(helpers.ERROR, "[DYNAMIC] Error executing %s: %v", fileName, err)
		return
	}

	ev := &DynamicEvent{
		Name: fileName,
		Path: fullPath,
		Lua: DynamicEventLua{
			LState:    L,
			OnStart:   getGlobalFunction(L, "on_start"),
			OnTick:    getGlobalFunction(L, "on_tick"),
			OnEvent:   getGlobalFunction(L, "on_event"),
			OnMessage: getGlobalFunction(L, "on_message"),
			OnCommand: getGlobalFunction(L, "on_command"),
			OnRequest: getGlobalFunction(L, "on_request"),
		},
		State: DynamicEventState{
			NextTick:                time.Now().Add(time.Second),
			Interval:                time.Second, // default
			Paused:                  true,        // default
			ComputeTwitchSharedChat: false,       // default
			db:                      nil,
		},
	}

	setFunctionOnTable(ev, eventTable)

	dynamicEventsMutex.RLock()
	oldEvent := dynamicEvents[fileName]
	dynamicEventsMutex.RUnlock()

	if oldEvent != nil {
		oldEvent.Transfer(ev)
		oldEvent.Close()
	}

	dynamicEventsMutex.Lock()
	dynamicEvents[fileName] = ev
	dynamicEventsMutex.Unlock()

	helpers.Logf(helpers.DEBUG, "[DYNAMIC] Event loaded: %s", fileName)

	ev.ProcessStart()
}

func setFunctionOnTable(ev *DynamicEvent, tbl *lua.LTable) {
	ev.Lua.mu.Lock()
	defer ev.Lua.mu.Unlock()

	if ev.Lua.LState == nil {
		return
	}

	ev.Lua.LState.SetField(tbl, "socket_send", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		if L.Get(1) == lua.LNil || L.Get(2) == lua.LNil {
			return -1
		}

		t := L.CheckString(1)
		d := L.CheckTable(2)

		globals.SafeSend(globals.WsBroadcast, globals.SocketMessage{
			Filter: ev.Name,
			Type:   t,
			Data:   TableToMap(d),
		}, "WsBroadcast")

		return 0
	}))

	ev.Lua.LState.SetField(tbl, "set_interval", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		val := L.CheckNumber(1)
		ev.State.mu.Lock()
		defer ev.State.mu.Unlock()
		ev.State.Interval = time.Duration(float64(val) * float64(time.Second))
		return 0
	}))

	ev.Lua.LState.SetField(tbl, "set_paused", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		val := L.CheckBool(1)
		ev.State.mu.Lock()
		defer ev.State.mu.Unlock()
		ev.State.Paused = val
		return 0
	}))

	ev.Lua.LState.SetField(tbl, "set_compute_twitch_shared_chat", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		val := L.CheckBool(1)
		ev.State.mu.Lock()
		defer ev.State.mu.Unlock()
		ev.State.ComputeTwitchSharedChat = val
		return 0
	}))

	ev.Lua.LState.SetField(tbl, "get_interval", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		ev.State.mu.RLock()
		defer ev.State.mu.RUnlock()
		L.Push(lua.LNumber(ev.State.Interval.Seconds()))
		return 1
	}))

	ev.Lua.LState.SetField(tbl, "is_paused", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		ev.State.mu.RLock()
		defer ev.State.mu.RUnlock()
		L.Push(lua.LBool(ev.State.Paused))
		return 1
	}))

	ev.Lua.LState.SetField(tbl, "use_db", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		createdDb := false
		ev.State.mu.Lock()
		defer ev.State.mu.Unlock()
		if ev.State.db == nil {
			db, err := msql.OpenModuleDB(ev.Name)
			if err != nil {
				helpers.Logf(helpers.ERROR, "[DYEVENTS USE_DB] %s", err.Error())
				return 0
			}
			ev.State.db = db
			createdDb = true
		}
		L.Push(lua.LBool(createdDb))
		return 1
	}))

	ev.Lua.LState.SetField(tbl, "db_exec", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		if L.Get(1) == lua.LNil {
			return 0
		}
		query := L.CheckString(1)
		ev.State.mu.RLock()
		defer ev.State.mu.RUnlock()
		if ev.State.db == nil {
			return 0
		}
		r, err := ev.State.db.Exec(query)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[DYEVENTS DB_EXEC] %s", err.Error())
			return 0
		}
		rows, err := r.RowsAffected()
		if err != nil {
			helpers.Logf(helpers.ERROR, "[DYEVENTS DB_EXEC] %s", err.Error())
			return 0
		}
		L.Push(lua.LNumber(rows))
		return 1
	}))

	ev.Lua.LState.SetField(tbl, "db_query", ev.Lua.LState.NewFunction(func(L *lua.LState) int {
		if L.Get(1) == lua.LNil {
			helpers.Logf(helpers.ERROR, "[QUERY] QUERY IS NIL")
			return 0
		}
		query := L.CheckString(1)
		ev.State.mu.RLock()
		if ev.State.db == nil {
			helpers.Logf(helpers.ERROR, "[QUERY] DB IS NIL")
			ev.State.mu.RUnlock()
			return 0
		}
		helpers.Logf(helpers.DEBUG, "[QUERY] %s", query)
		r, err := ev.State.db.Query(query)
		ev.State.mu.RUnlock()
		defer r.Close()

		if err != nil {
			helpers.Logf(helpers.ERROR, "[QUERY] ERROR: %s", err.Error())
			return 0
		}

		cols, _ := r.Columns()
		helpers.Logf(helpers.DEBUG, "COLS %v", cols)
		result := []map[string]any{}

		for r.Next() {
			d := map[string]any{}
			items := make([]any, len(cols))
			for i := range items {
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
		L.Push(ToLValue(ev.Lua.LState, result))
		return 1
	}))

}

// Allow internal websocket events
func HandleDyEventWebsocket(msg any) {
	dynamicEventsMutex.RLock()
	defer dynamicEventsMutex.RUnlock()

	for _, ev := range dynamicEvents {
		ev.ProcessWebsocketEvent(msg)
	}
}

// Single loop for all events
func globalEventLoop() {
	helpers.Log(helpers.INFO, "Started dyevent global event loop!")
	ticker := time.NewTicker((1 * time.Second) / 60)
	defer ticker.Stop()

	for {
		select {
		case <-stopGlobalLoop:
			helpers.Logf(helpers.DEBUG, "[DYNAMIC] Loop global parado")
			return

		case now := <-ticker.C:
			dynamicEventsMutex.RLock()
			for _, ev := range dynamicEvents {
				if !ev.CanProcess() {
					continue
				}

				ev.State.mu.RLock()
				if ev.State.Interval == 0 {
					ev.State.mu.RUnlock()
					continue
				}
				nextTick := ev.State.NextTick
				ev.State.mu.RUnlock()

				if !now.After(nextTick) {
					continue
				}

				ev.ProcessOnTick()

				ev.State.mu.Lock()
				ev.State.LastTick = now
				ev.State.NextTick = now.Add(ev.State.Interval)
				ev.State.mu.Unlock()
			}
			dynamicEventsMutex.RUnlock()
		}
	}
}

func getGlobalFunction(L *lua.LState, name string) *lua.LFunction {
	f := L.GetGlobal(name)
	if fn, ok := f.(*lua.LFunction); ok {
		return fn
	}
	return nil
}
