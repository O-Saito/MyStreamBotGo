# Plugin system foundation (`/plugin`)

## Context

This branch (`temporary-addons-system`) is laying the groundwork for a custom
plugin system. Eventually plugins will be loaded as external `.bin`/`.so`
artifacts, but for now we need the **interface and runtime contract** that
such plugins (and any compiled-in plugins) will implement: an event emitter,
a receiver for chat/event data coming from the platform APIs, and an actions
registry that exposes Go functions to Lua CustomEvents (`ev`/dynamic events)
the same way `services.ExposeToLua` does for `twitch`/`kick`/`youtube`.

The design must slot into the existing architecture described in
`docs/project.md` and `docs/Module.md`:
- Channels in `globals/globals.go` (`ChatQueue`, `CommandQueue`, `EventQueue`,
  `WsBroadcast`) are the backbone; `mlua.DyEventQueue` is how CustomEvents
  receive the same data.
- `services/api_to_lua.go` (`services.LuaFunction`, `services.ExposeToLua`)
  is the existing reflection-based bridge for exposing Go funcs to Lua.
- `mlua.Init(funcs ...func(*lua.LState))` (`mlua/mlua.go:71`) registers
  global-state setup functions across `LChat`, `LCommands`, `LEvents`, and
  every dynamic-event `LState` (via `globalRegister`, used in
  `LoadDyEventModule`, `mlua/dyevents.go:484`). This is the mechanism plugin
  actions will piggyback on so `plugin_<name>` tables are available in every
  Lua state, including CustomEvents.
- Processors (`processors/chat_queue_processor.go`, `event_queue_processor.go`)
  already fan out queue items to `mlua.DyEventQueue` and static Lua handlers
  — plugin dispatch will sit alongside those `DyEventQueue` pushes, seeing
  the same unfiltered stream.

Plugins receive chat messages and platform events. They do **not** receive
commands or WebSocket broadcasts directly.

## New package: `plugin/`

### `plugin/plugin.go`

Defines the core contract:

```go
type Plugin interface {
    Name() string
    Init(ctx *Context) error
    Start() error
    Stop() error
    Actions() []services.LuaFunction
    OnChat(msg *globals.MessageFromStream)
    OnEvent(event *globals.Event)
}
```

Plus `BasePlugin` — a struct with no-op implementations of every method so
concrete plugins can embed it and override only the hooks they need (mirrors
how `DynamicEventLua` hooks are individually optional).

### `plugin/context.go`

`Context` struct passed to `Init`, giving plugins a host-facing API without
needing to import `globals`/`mlua` directly everywhere:

```go
type Context struct {
    Name string // plugin name, set by the registry
}

func (c *Context) EmitEvent(eventType string, data map[string]any, target ...string) {
    if len(target) > 0 {
        globals.LuaRequest <- globals.SocketMessage{
            Filter: target[0] + ".lua",
            Type:   eventType,
            Data:   data,
        }
    } else {
        globals.EventQueue <- globals.Event{Type: eventType, Data: data}
    }
}

func (c *Context) Logf(level helpers.Level, format string, a ...any) {
    helpers.Logf(level, "[PLUGIN:"+c.Name+"] "+format, a...)
}
```

This is the "event emitter" piece — two modes in one method:

- **No target** — pushes to `globals.EventQueue`, which flows through
  `ProcessEventQueue` to `WsBroadcast`, `DyEventQueue`, `HandleEvent`, and all
  plugins' `OnEvent`. This is a broadcast event.

- **With target** (e.g. `EmitEvent("milestone", data, "leaderboard")`)
  — pushes to `globals.LuaRequest` with the Lua module filename as `Filter`.
  The existing `ProcessLuaRequest` → `DyEventQueue` → `ProcessRequest` pipeline
  at `mlua/dyevents.go:314` delivers it **only** to the matching Lua module's
  `on_request(type, data)` hook. No other modules or plugins receive it.

Lua module usage — the target module just implements `on_request`:
```lua
-- modules/customevents/leaderboard.lua
function on_request(type, data)
    if type == "milestone" then
        ev.socket_send("milestone_update", data)
    end
end
```

### `plugin/registry.go`

Holds the registered plugins and dispatch/exposure logic:

```go
var (
    mu         sync.RWMutex
    registered = map[string]Plugin{}
)

func Register(p Plugin)                                    // called from each plugin's init()
func InitAll()                                             // calls p.Init(ctx) for all registered plugins
func StartAll()                                            // calls p.Start() for all registered plugins
func StopAll()                                             // calls p.Stop() for all registered plugins

func DispatchChat(msg *globals.MessageFromStream)          // fan-out to OnChat
func DispatchEvent(event *globals.Event)                    // fan-out to OnEvent

// RegisterLuaActions is passed into mlua.Init(...) so plugin_<name> tables
// are exposed in LChat/LCommands/LEvents and every dynamic-event LState.
func RegisterLuaActions(L *lua.LState)
```

`RegisterLuaActions` iterates `registered`, and for each plugin with a
non-empty `Actions()`, calls `services.ExposeToLua(L, "plugin_"+p.Name(), actions)`
— following the exact pattern used for `twitch`/`kick`/`youtube` in
`lua_functions.go:164` (`RegisterServiceAPIs`).

Dispatch functions hold the read lock briefly, copy the slice of plugins, and
call hooks without holding the lock (avoids deadlock if a plugin emits an
event synchronously inside `OnChat`/`OnEvent` that loops back through
`DispatchEvent`).

## Wiring into existing flow

### `processors/chat_queue_processor.go`
Add `plugin.DispatchChat(&ev)` immediately after the existing
`mlua.DyEventQueue <- mlua.DyEventQueueData{Type: mlua.DyEventChat, ...}` push
(before the Twitch shared-chat filter), so plugins see every chat message
regardless of source/channel — same as CustomEvents.

### `processors/event_queue_processor.go`
Add `plugin.DispatchEvent(&ev)` immediately after the
`mlua.DyEventQueue <- mlua.DyEventQueueData{Type: mlua.DyEventEvent, ...}` push.
(Note: `ev.Type` has already been normalized from
`twitch-eventsub-notification` to the underlying subscription type by this
point, matching what `mlua.HandleEvent` receives.)

### `processors/command_queue_processor.go` / `lua_request_processor.go`
**No changes.** Plugins do not receive commands or Lua requests directly.

## Wiring into `main.go`

- No blank import for the `plugin` package itself is needed — only future
  concrete plugin packages will need one. The package is simply ready for
  future plugin packages to do `import "MyStreamBot/plugin"` +
  `func init() { plugin.Register(&MyPlugin{}) }`, with `main.go` adding
  `_ "path/to/plugin/package"`.
- After `globals.LoadInitFile()`/DB setup and before `mlua.Init(...)`, call
  `plugin.InitAll()`.
- Update `mlua.Init(RegisterLuaFunctions, RegisterServiceAPIs)` to
  `mlua.Init(RegisterLuaFunctions, RegisterServiceAPIs, plugin.RegisterLuaActions)`.
- After starting processor goroutines, call `plugin.StartAll()`.
- On shutdown (after `mlua.SaveDynamicEvents()`), call `plugin.StopAll()`.

## Verification

- `go build ./...` to confirm the new package compiles and wiring in
  `main.go`/processors has no import cycles (`plugin` imports `globals` and
  `services`; `services` imports `mlua`; neither imports `plugin`, so no
  cycle).
- `go vet ./...` for sanity.
- Since there's no concrete plugin yet, full runtime verification (actions
  appearing in Lua, dispatch firing) isn't possible in this change — note
  this explicitly to the user. A follow-up task can add a real plugin to
  exercise the system.
