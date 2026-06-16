# Module (Lua) and front-end CustomEvents

This document explains how to write and install Lua modules for MyStreamBot and how to provide optional front-end (JS) CustomEvents wrappers. It references loader and bridge functions so examples use the exact runtime APIs.

## Overview

- Supported module types: commands, chat modules, event modules, dynamic custom events (CustomEvents).
- Place modules under `modules/` for a new module and `modules/customevents/modules` for imports.
- The Go loader (package `mlua`) initializes separate Lua states for commands, chat and dynamic events and hot-reloads modules when files change.

Key loader functions to refer to: [mlua/mlua.go](./../mlua/mlua.go) (`Init`, `LoadAllModules`, `loadModule`, `StartWatcher`) and [mlua/dyevents.go](./../mlua/dyevents.go) (`LoadDyEvents`, `setFunctionOnTable`).

## Where to put modules

- Chat modules: `modules/chat/`
- Command modules: `modules/commands/`
- Static event modules: `modules/events/<eventname>/`
- Dynamic/custom events (CustomEvents): `modules/customevents/`

See examples in the repository:
- [modules/commands/first.lua](./../modules/commands/first.lua)
- [modules/chat/first.lua](./../modules/chat/first.lua)
- [modules/customevents/customevent_example.lua](./../modules/customevents/customevent_example.lua)
- [modules/events/channel.follow/follow_example.lua](./../modules/events/channel.follow/follow_example.lua)

## Module types & lifecycle hooks

The loader detects functions exported as globals in a module and registers them according to module type. Common lifecycle hook names the loader recognizes:

### Chat modules `modules/chat`

Receives any chat message that is not sent by the bot.

It should have an `on_message` function

```lua
function on_message(ev) end
```

### Commands modules `modules/commands`

Receives the command `!<command_name>`, the `<command_name>` is the file name without `.lua` 

It should have an `on_command` function

```lua
function on_command(ev) end
```

### Static event modules `modules/<event_name>`
Receives events based on event name folder ([modules/events/channel.follow/follow_example.lua](./../modules/events/channel.follow/follow_example.lua)) 

It should have an `on_event` function

```lua
function on_event(data) end
```

### CustomEvent

It can be more complex, receiving all of the above options

It can have:
- `on_start` — called when a CustomEvent module is loaded or (re)started.
- `on_tick` — called periodically if an interval is set `ev.set_interval`.
- `on_event` — called when the server dispatches a named event to this module.
- `on_message` — receives chat/message payloads.
- `on_command` — receives command invocations.
- `on_request` — receives arbitrary requests from the frontend or other parts.

It can also import modules `local string_helper = require("string_helper")` that is placed on `/modules/customevents/modules`

```lua
local string_helper = require("string_helper")

function on_request(type, data) end

function on_start() end

function on_tick(data) end

function on_message(msg) end

function on_event(name, data) end

function on_command(name, data) end
```

## Keeping the data

### Persistent in memory data (for hot reload)

To keep the data between hot reloads it's needed `g.get/g.set` if is a CustomEvent it can use `ev.data`

### Real Persistent data

For all types it have `g.kv_get(key)/g.kv_set(key, value)` it saves into a SQLite db into a table of key/value 
For CustomEvents it can create it's own database with `ev.use_db()`

Example:
```lua
function on_start() 
  ev.use_db()
  ev.db_exec("CREATE TABLE IF NOT EXISTS my_table (id INTEGER PRIMARY KEY, value TEXT)")
end
```

## CustomEvent (`ev`) helper API

CustomEvents are dynamic modules loaded into their own Lua `LState` and receive an `ev` table with helper functions. The API surface can be found in `definitons/`

Minimal CustomEvents `on_start` example:

```lua
function on_start()
  ev.set_interval(1)        -- run on_tick every second
  ev.set_paused(false)
end
```

Notes:
- `ev.data` is preserved across reloads for CustomEvents; use it to store module state.
- The `Filter` used in websocket messages is the module filename (including extension) — front-end subscribe calls must match it exactly.

## Front-end CustomEvents wrappers (JS)

The front-end uses `web/wrapper.js` to open a websocket and receive an `init` payload containing `custom_events_modules`. The wrapper exposes `subscribe(moduleName)` which performs an `upgrade-conn` handshake and returns an object with `on(type, fn)` and `send(type, data)`.

Typical client usage (from `web/customevents/*.js`):

```js
const example = w.subscribe("customevent_example.lua");
if (example) {
  example.on("tick", (data) => console.log("Tick:", data));
  example.send("ping", {}, (data) => console.log("Pong:", data));
}
```

Notes:
- `subscribe` expects the module filename that matches the server `Filter` (usually `vote.lua`).
- The frontend and CustomEvent communicate using messages of the shape `{ type: string, filter: string, data: any }`.

See `web/wrapper.js` for the exact client API and `web/customevents/customevent_example.lua.js` for real examples.

## Hot-reload behavior

- The loader starts a filesystem watcher (`StartWatcher`) and will reload modified modules. The watcher uses a debounce; rapid file changes can race.
- CustomEvents preserve `ev.data` across reloads; other state is reinitialized.
- When reloading, module `on_start` will be called again for CustomEvents.

## Editor definitions (IntelliSense)

The repository includes a `definitions/` folder with Lua declaration (stub) files intended for editor tooling and IntelliSense. These files document the APIs the Go runtime exposes to Lua modules (names, parameters and return shapes) so your editor can provide accurate completions and hover signatures.

Files present in `definitions/` (examples):

- `ev.d.lua` — declarations for the `ev` table used by CustomEvents (e.g. `ev.socket_send(type, table)`, `ev.set_interval(number)`, `ev.set_paused(bool)`, `ev.db_query(sql)` and `ev.data`).
- `g.d.lua` — declarations for the `g` helpers/globals exposed to modules.
- `twitch.d.lua` — declarations for `twitch.*` helpers (e.g. `twitch.get_channel_stream_data`, `twitch.get_user_data`).

How to use them:

- Open these `.d.lua` files when authoring modules to verify exact function names and expected return values.
- Configure your Lua language server (e.g. Lua Language Server / Sumneko) to include the `definitions` directory in the workspace library so you get completions and hover info. Example VS Code setting:

```json
{
  "Lua.workspace.library": [
    "./definitions"
  ]
}
```

- Remember these files are documentation stubs — they are not executed by the runtime. Use them to check "what is called with what and the return" while writing modules.

Recommendation: consult the matching `.d.lua` before calling an API (especially `ev.*` and `twitch.*`) to avoid mismatches.

## References (source files)

- `mlua/mlua.go` — loader and module registration
- `mlua/dyevents.go` — CustomEvent loader and `ev` helper bindings
- `lua_functions.go` — functions exposed to Lua (`twitch` helpers, etc.)
- `globals/globals.go` — `SocketMessage` and `WsBroadcast` channel
- `web/wrapper.js` — client websocket wrapper and `subscribe` API
- `web/customevents/vote.lua.js` — example UI wrappers
- `modules/customevents/*.lua`, `modules/commands/*.lua`, `modules/chat/*.lua` — example modules
