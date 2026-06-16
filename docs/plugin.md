# Plugin system

## Overview

The plugin system provides a way to extend MyStreamBot with custom Go code.
Plugins can receive chat messages and platform events, register actions
callable from Lua, and emit events back into the bot pipeline. The system is
designed so that a future `.so`/`.bin` loader can be added without changing
the interface.

## Plugin interface

Every plugin must implement the `Plugin` interface:

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

### Method reference

| Method | When called | Purpose |
|--------|-------------|---------|
| `Name()` | During `Register()` | Unique plugin identifier |
| `Init(ctx)` | During `InitAll()` | Receive `Context`, set up data structures |
| `Start()` | During `StartAll()` | Begin work (e.g. spawn goroutines, connect) |
| `Stop()` | During `StopAll()` | Cleanup (e.g. close connections, stop goroutines) |
| `Actions()` | During `RegisterLuaActions()` | Return list of functions exposed to Lua as `plugin_<name>.*` |
| `OnChat(msg)` | On every chat message | Process chat messages from all platforms |
| `OnEvent(event)` | On every platform event | Process events (follows, subs, raids, etc.) |

## BasePlugin

To avoid implementing every method, embed `BasePlugin` in your plugin struct:

```go
type MyPlugin struct {
    plugin.BasePlugin
}

func (p *MyPlugin) Name() string { return "myplugin" }
func (p *MyPlugin) Init(ctx *plugin.Context) error { return nil }
```

`BasePlugin` provides no-op defaults for `Init`, `Start`, `Stop`, `Actions`,
`OnChat`, and `OnEvent`. Override only what you need.

## Context

`Context` is passed to `Init()` and provides interaction with the bot:

```go
type Context struct {
    Name string // set by the registry to match the plugin name
}

// EmitEvent broadcasts or sends a directed event.
//   No target → pushes to EventQueue (broadcast to all Lua modules, plugins, WS clients).
//   With target → pushes to LuaRequest with filter (delivered only to the named Lua module's on_request).
func (c *Context) EmitEvent(eventType string, data map[string]any, target ...string)

// Logf logs with a [PLUGIN:<name>] prefix.
func (c *Context) Logf(level helpers.Level, format string, a ...any)
```

### EmitEvent — two modes

```go
// Broadcast — all Lua modules see it via on_event, all plugins via OnEvent
ctx.EmitEvent("raid", map[string]any{"from": "channel", "viewers": 50})

// Directed — only the Lua module "chatbot" receives it via on_request
ctx.EmitEvent("ai_response", map[string]any{"text": "Hello!"}, "chatbot")
```

### Lua module receiving directed events

```lua
-- modules/customevents/chatbot.lua
function on_request(type, data)
    if type == "ai_response" then
        g.send_message("twitch", data.channel, data.text)
    end
end
```

## Actions (Lua bridge)

Actions are Go functions exposed to every Lua state (chat modules, command
modules, event modules, and custom events) as `plugin_<name>.<function>()`:

```go
func (p *MyPlugin) Actions() []services.LuaFunction {
    return []services.LuaFunction{
        {Name: "greet", Fn: func(name string) string {
            return "Hello, " + name + "!"
        }},
        {Name: "set_timer", Fn: func(seconds float64) {
            // start a timer
        }},
    }
}
```

Lua usage:

```lua
local msg = plugin_myplugin.greet("Viewer")
g.log(msg)  -- "Hello, Viewer!"

plugin_myplugin.set_timer(30)
```

The bridge uses reflection (same as `twitch.*`, `kick.*`, `youtube.*`) and
supports automatic conversion of strings, numbers, booleans, and slices.

## Lifecycle

The plugin manager calls lifecycle methods in order:

1. **`Register(p)`** — called from each plugin's `init()` (or manually)
2. **`InitAll()`** — calls `Init(ctx)` on all registered plugins
3. **`StartAll()`** — calls `Start()` on all registered plugins
4. Runtime — `OnChat`/`OnEvent` fire as events arrive
5. **`StopAll()`** — calls `Stop()` on all registered plugins (on shutdown)

## Registering a plugin

Create a Go package and register in an `init()` function:

```go
package myplugin

import "MyStreamBot/plugin"

type Plugin struct {
    plugin.BasePlugin
    ctx *plugin.Context
}

func init() {
    plugin.Register(&Plugin{})
}

func (p *Plugin) Name() string { return "myplugin" }

func (p *Plugin) Init(ctx *plugin.Context) error {
    p.ctx = ctx
    p.ctx.Logf(helpers.DEBUG, "initialized")
    return nil
}

func (p *Plugin) Start() error {
    p.ctx.Logf(helpers.DEBUG, "started")
    return nil
}

func (p *Plugin) Stop() error {
    p.ctx.Logf(helpers.DEBUG, "stopped")
    return nil
}
```

Then add a blank import in `main.go`:

```go
import _ "MyStreamBot/plugins/myplugin"
```

## What plugins do NOT receive

- Commands (from `CommandQueue`)
- WebSocket broadcasts (from `WsBroadcast`)
- Lua requests (from `LuaRequest`)

## Future: `.so` / `.bin` plugins

The `Plugin` interface is the contract. A future loader will:

- **`.so`**: call `plugin.Open()` + `Lookup("Plugin")` to get a `Plugin`
- **`.bin`**: spawn a subprocess, wrap IPC in a struct that implements `Plugin`

No changes to the interface, processors, Context, or Lua bridge are needed.
