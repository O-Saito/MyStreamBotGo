# Plugin system

## Overview

The plugin system provides two ways to extend MyStreamBot:

1. **Compile-time Go plugins** — any package in the `MyStreamBot` module that
   implements `contract.Plugin` and calls `plugin.Register()`.
2. **Runtime C ABI native plugins** — `.dll` / `.so` / `.dylib` shared
   libraries placed in `./plugins/`, loaded at startup via CGo.

Both kinds receive chat messages and platform events, register Lua-callable
actions, and emit events back into the bot pipeline. The native plugin ABI
uses JSON strings for all data crossing the C/Go boundary and follows the
same lifecycle as a compile-time plugin.

## Package layout

Compile-time Go plugins depend only on `MyStreamBot/plugin`, which exports
the `Plugin` interface, `Context`, and `BasePlugin`. (`plugin/loader/` and
`plugin/sdk/` are internal to the host and native-plugin ABI, not something
a plugin author imports directly.)

## Plugin interface

Every plugin must implement the `contract.Plugin` interface:

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

`BasePlugin` provides no-op defaults for every method. Override only what you
need.

## Context

`Context` is passed to `Init()` and provides interaction with the bot:

```go
type Context struct {
    Name string // set by the registry to match the plugin name
}

// EmitEvent broadcasts or sends a directed event.
//   No target → pushes to EventQueue (broadcast to all Lua modules, plugins, WS clients).
//   With target → pushes to LuaRequest (delivered only to the named Lua module's on_request).
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

### Lua safety — plugin not loaded

The `plugin_<name>` global table only exists if the plugin was loaded _and_
returned a non-empty `Actions()`. If a plugin isn't loaded, the table is
`nil`. Safely guard call sites:

```lua
-- Option A: nil guard (simplest)
if plugin_keylistener then
    plugin_keylistener.listen("my_module")
end

-- Option B: pcall with anonymous function (catches nil table error too)
local ok, err = pcall(function()
    return plugin_keylistener.listen("my_module")
end)
```

`pcall(plugin_keylistener.listen, ...)` does **not** work — the nil-index
error happens during argument evaluation, before `pcall` runs.

## Lifecycle

The plugin manager calls lifecycle methods in order:

1. **`LoadPlugins("./plugins")`** — scans `./plugins/` for `.dll`/`.so`/`.dylib`
   and loads each as a native plugin
2. **`Register(p)`** — called for each loaded native plugin (or from compile-time
   plugin `init()` functions)
3. **`InitAll()`** — calls `Init(ctx)` on all registered plugins
4. **`StartAll()`** — calls `Start()` on all registered plugins
5. **Runtime** — `OnChat`/`OnEvent` fire as events arrive
6. **`StopProcessors()`** — signals processor goroutines to drain their
   channels before shutdown (prevents concurrent dispatch during unload)
7. **`StopAll()`** — calls `Stop()`, then unloads native plugin libraries
   (see Thread safety below for the in-flight-dispatch guarantee this relies on)

### Thread safety

The host guarantees no `OnChat`/`OnEvent`/action call is in flight for a
plugin while `Stop()` and native library unload run for that same plugin —
you don't need your own locking to protect against concurrent dispatch
during shutdown.

### Panic recovery

If a plugin's `Init`, `Start`, `Stop`, `OnChat`, or `OnEvent` panics, the
host recovers it, logs the panic and stack trace, and (except during `Stop`)
permanently stops dispatching to that plugin. A misbehaving plugin can't
crash the host process, but a panic in `Init`/`Start`/`OnChat`/`OnEvent`
disables the plugin for the rest of the run — don't rely on panics for
control flow.

---

## Native C ABI plugins

Native plugins are standalone shared libraries compiled with
`-buildmode=c-shared`. They are loaded at runtime by `plugin/loader/`.

### ABI contract

Each plugin must export the following C functions. The plugin's C code
receives a `bot_host_api_t*` struct (defined in `plugin/loader/hostapi.h`)
that provides host callbacks.

#### Mandatory exports

| C export | Signature | Purpose |
|---|---|---|
| `plugin_name` | `const char* plugin_name(void)` | Returns malloc'd C string with plugin name; freed by host via `plugin_free` |
| `plugin_api_version` | `uint32_t plugin_api_version(void)` | Must return `0x00020000`; mismatched versions are rejected |
| `plugin_set_host` | `void plugin_set_host(const bot_host_api_t* api)` | Stores the host API struct for use by `host_log`/`host_emit_event` helpers |
| `plugin_free` | `void plugin_free(void* ptr)` | Frees memory allocated by the plugin (used by host for return values) |

#### Optional exports

All optional exports use **JSON strings** for input/output. The host calls
`plugin_free` on any non-nil return value.

| C export | Signature | Purpose |
|---|---|---|
| `plugin_init` | `const char* plugin_init(const char* configJSON)` | Called during `InitAll()`; return error JSON or NULL |
| `plugin_start` | `void plugin_start(void)` | Called during `StartAll()` |
| `plugin_stop` | `void plugin_stop(void)` | Called during `StopAll()` before library unload |
| `plugin_on_chat` | `void plugin_on_chat(const char* msgJSON)` | Receives serialized `MessageFromStream` |
| `plugin_on_event` | `void plugin_on_event(const char* evtJSON)` | Receives serialized `Event` |
| `plugin_get_actions` | `const char* plugin_get_actions(void)` | Returns JSON array `[{"name":"..."}, ...]` |
| `plugin_call_action` | `const char* plugin_call_action(const char* name, const char* argsJSON, const char* metaJSON)` | Calls a named action with JSON args, returns JSON result. `metaJSON` carries the calling Lua module's identity (from `__plugin_meta`) |

### Host API struct (`bot_host_api_t`)

Defined in `plugin/loader/hostapi.h`:

```c
#define MYSTREAM_BOT_API_VERSION 0x00020000

typedef struct bot_host_api_s {
    uint32_t api_version;
    void     (*log)(int level, char *msg);
    char *   (*emit_event)(char *type, char *data, char *target);
    void     (*free)(void *ptr);
} bot_host_api_t;
```

The plugin typically stores this in a global and wraps it with helpers:

```go
var hostAPI *C.bot_host_api_t

//export plugin_set_host
func plugin_set_host(api *C.bot_host_api_t) {
    hostAPI = api
}

func hostLog(level int, format string, args ...any) {
    msg := C.CString(fmt.Sprintf(format, args...))
    defer C.free(unsafe.Pointer(msg))
    C.host_log(hostAPI, C.int(level), msg)
}
```

### JSON protocol

All data crossing the C/Go boundary for optional exports is JSON-encoded.
The schema for each function mirrors the Go types in `plugin/sdk/types.go`:

- `plugin_on_chat`: serialized `globals.MessageFromStream` — includes
  `IsCommand` (true if the message starts with the bot's command prefix) and
  `IsAtOwnerChannel` (true if the message originates from the bot owner's
  own channel/session)
- `plugin_on_event`: serialized `globals.Event`
- `plugin_init`: receives `{"name":"<plugin name>"}`, returns error string or null
- `plugin_get_actions`: returns `[{"name":"listen"},{"name":"stop_listen"}]`
- `plugin_call_action`: receives action name + JSON args + JSON meta (calling module identity), returns JSON result

### Null pointer safety

Optional C exports are looked up by name at load time. If your plugin
doesn't implement one, the host simply skips calling it — omitting an
optional export is safe and never crashes the host.

### Building a native plugin

Create a standalone Go module with `-buildmode=c-shared`. The repo ships a
minimal starting point at `plugins/template_go/`:

```
plugins/template_go/
├── go.mod      (standalone module, no dependency on MyStreamBot)
├── main.go     (all C exports + Lua action dispatch in one file)
├── build.bat   (Windows build script)
└── build.sh    (Unix build script)
```

Build (see `build.bat` / `build.sh`):

```
cd plugins/template_go
go build -buildmode=c-shared -o ../template_plugin.dll .
```

The output `.dll`/`.so` goes into `./plugins/`, which is scanned at startup
by `plugin.LoadPlugins("./plugins")` in `main.go`.

### Reference plugin: template_go

`plugins/template_go/main.go` is a minimal, single-file example plugin
(`template_plugin`) that demonstrates:

- All mandatory C exports (`plugin_name`, `plugin_api_version`,
  `plugin_set_host`, `plugin_free`)
- Optional lifecycle exports (`plugin_init`, `plugin_start`, `plugin_stop`)
- `plugin_on_chat` / `plugin_on_event` JSON unmarshaling
- `plugin_get_actions` / `plugin_call_action` with the 3-argument
  `(name, argsJSON, metaJSON)` signature, including reading the caller's
  identity out of `metaJSON`
- Two example actions: `hello` and `echo`

---

## What plugins do NOT receive

- Commands (from `CommandQueue`)
- WebSocket broadcasts (from `WsBroadcast`)
- Lua requests (from `LuaRequest`)

## Registering a compile-time Go plugin

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
