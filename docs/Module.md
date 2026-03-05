# Module authoring (Lua) and front-end dyevents

This document explains how to write and install Lua modules for MyStreamBot and how to provide optional front-end (JS) dyevent wrappers. It references loader and bridge functions so examples use the exact runtime APIs.

## Overview

- Supported module types: commands, chat modules, event modules, dynamic custom events (DyEvents).
- Place modules under `modules/` (preferred) or `build/modules/` for packaged builds.
- The Go loader (package `mlua`) initializes separate Lua states for commands, chat and dynamic events and hot-reloads modules when files change.

Key loader functions to refer to: [mlua/mlua.go](mlua/mlua.go) (`Init`, `LoadAllModules`, `loadModule`, `StartWatcher`) and [mlua/dyevents.go](mlua/dyevents.go) (`LoadDyEvents`, `setFunctionOnTable`).

## Where to put modules

- Command modules: `modules/commands/`
- Chat modules: `modules/chat/`
- Static event modules: `modules/events/<eventname>/`
- Dynamic/custom events (DyEvents): `modules/customevents/`

See examples in the repository:
- [modules/commands/first.lua](modules/commands/first.lua)
- [modules/chat/first.lua](modules/chat/first.lua)
- [modules/customevents/vote.lua](modules/customevents/vote.lua)

## Module types & lifecycle hooks

The loader detects functions exported as globals in a module and registers them according to module type. Common lifecycle hook names the loader recognizes:

- `on_start` — called when a DyEvent module is loaded or (re)started.
- `on_tick` — called periodically if an interval is set.
- `on_event` — called when the server dispatches a named event to this module.
- `on_message` — receives chat/message payloads (chat modules / dynamic modules).
- `on_command` — receives command invocations.
- `on_request` — receives arbitrary requests from the frontend or other parts.

The loader registers handlers by checking for these globals; see the detection logic in `mlua.loadModule` and `LoadDyEvents` ([mlua/mlua.go](mlua/mlua.go), [mlua/dyevents.go](mlua/dyevents.go)).

Example detection (for reference):

```go
// pseudocode from repo
f := L.GetGlobal("on_command")
if fn, ok := f.(*lua.LFunction); ok {
    commandFunctions[moduleName] = fn
}
```

## DyEvent (`ev`) helper API

DyEvents are dynamic modules loaded into their own Lua `LState` and receive an `ev` table with helper functions. The API surface can be found on `definitons`

Minimal DyEvent `on_start` example:

```lua
function on_start()
  ev.setInterval(1)        -- run on_tick every second
  ev.setPaused(false)
  ev.socket_send("config", { votos = ev.data and ev.data.votos or {} })
end
```

Notes:
- `ev.data` is preserved across reloads for DyEvents; use it to store module state.
- The `Filter` used in websocket messages is the module filename (including extension) — front-end subscribe calls must match it exactly.

## Command modules

Command modules typically export `on_command`. The Go side dispatches commands via `HandleCommand` (see [mlua/mlua.go](mlua/mlua.go)). A minimal command handler:

```lua
function on_command(name, payload)
  -- name: command invoked, payload: table with metadata
  -- respond using available helpers (e.g. g.send_message or via global APIs)
end
```

The loader stores command handlers by module filename; see `HandleCommand` and the `commandFunctions` registration in `mlua`.

## Chat modules

Chat modules implement `on_message` to inspect incoming chat messages. Example:

```lua
function on_message(msg)
  -- msg is a table representing the incoming message
  if msg.text == "!hello" then
    -- send response (via exposed API such as `g.send_message` if available)
  end
end
```

The chat dispatcher uses `HandleChat` in the `mlua` package.

## Front-end dyevent wrappers (JS)

The front-end uses `web/wrapper.js` to open a websocket and receive an `init` payload containing `custom_events_modules`. The wrapper exposes `subscribe(moduleName)` which performs an `upgrade-conn` handshake and returns an object with `on(type, fn)` and `send(type, data)`.

Typical client usage (from `web/dyevents/*.js`):

```js
const vote = w.subscribe("vote.lua");
if (vote) {
  vote.on("user_vote_update", onVoteUpdate);
  vote.send("setup", { opcoes: ["Sim","Não"] });
}
```

Notes:
- `subscribe` expects the module filename that matches the server `Filter` (usually `vote.lua`).
- The frontend and DyEvent communicate using messages of the shape `{ type: string, filter: string, data: any }`.

See `web/wrapper.js` for the exact client API and `web/dyevents/vote.lua.js` for real examples.

## Exposed services and helpers

The Go code exposes common services to Lua modules via `ExposeServiceToLua` and `luaFunctions.go`. Commonly exposed namespaces include:

- `g` — general helpers and globals (see `luaFunctions.go` or `definitions/g.d.lua`).
- `twitch` — helpers to query Twitch data (`twitch.get_channel_stream_data`, `twitch.get_user_data`, etc.).

Check `luaFunctions.go` for the full list of helper functions exported to Lua.

## Hot-reload behavior

- The loader starts a filesystem watcher (`StartWatcher`) and will reload modified modules. The watcher uses a debounce; rapid file changes can race.
- DyEvents preserve `ev.data` across reloads; other state is reinitialized.
- When reloading, module `on_start` will be called again for DyEvents.

## Testing & verification

Quick manual checks:

1. Start the bot:

```bash
go run .
```

2. Check stdout/logs for module load messages emitted by the loader (`Init`, `LoadAllModules`).
3. Trigger a command or chat message and verify the module responded (watch logs or chat output).
4. For DyEvents + frontend: open `web/index.html` in a browser, ensure the `init` payload contains `custom_events_modules`, call `w.subscribe("<module.lua>")` in the console and observe messages.

## Caveats & recommendations

- Module identity: DyEvent `Filter` is the module filename (including `.lua`). Match it exactly from the frontend.
- `package.path` is adjusted for DyEvents; place helper modules under `modules/.../modules/` if they will be `require()`d by dynamic events.
- Use `ev.setInterval` and `ev.setPaused` for periodic work instead of external goroutines.
- Check `luaFunctions.go` for available `twitch.*` helpers before duplicating functionality.

## Editor definitions (IntelliSense)

The repository includes a `definitions/` folder with Lua declaration (stub) files intended for editor tooling and IntelliSense. These files document the APIs the Go runtime exposes to Lua modules (names, parameters and return shapes) so your editor can provide accurate completions and hover signatures.

Files present in `definitions/` (examples):

- `ev.d.lua` — declarations for the `ev` table used by DyEvents (e.g. `ev.socket_send(type, table)`, `ev.setInterval(number)`, `ev.setPaused(bool)`, `ev.db_query(sql)` and `ev.data`).
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

Recommendation: consult the matching `.d.lua` before calling an API (especially `ev.*` and `twitch.*`) to avoid mismatches. Do not treat any single example module as the definitive API; prefer the `.d.lua` declarations and the actual Go bindings (`luaFunctions.go`, `mlua/dyevents.go`) when in doubt.

## References (source files)

- `mlua/mlua.go` — loader and module registration
- `mlua/dyevents.go` — DyEvent loader and `ev` helper bindings
- `luaFunctions.go` — functions exposed to Lua (`twitch` helpers, etc.)
- `globals/globals.go` — `SocketMessage` and `WsBroadcast` channel
- `web/wrapper.js` — client websocket wrapper and `subscribe` API
- `web/dyevents/vote.lua.js` — example UI wrappers
- `modules/customevents/*.lua`, `modules/commands/*.lua`, `modules/chat/*.lua` — example modules
