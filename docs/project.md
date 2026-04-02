# Project
This project is made in go because you can tell "go build", "go run" and the goroutines is "go DoThisThing".
The second reasons is because is good for API consumption.


## Repository Structure

```
mystreambot/
├── main.go                     # Initializer
├── handlers.go                 # General socket/request state configuration (for missing a better place)
├── lua_functions.go            # General lua state configuration (for missing a better place)
├── globals/                    # general config and state
├── goweb/                      # for webapi access
├── helpers/                    # shared utility functions
├── services/                   # streaming services
│   ├── kick/                   # Implementation of kick API
│   ├── twitch/                 # Implementation of twitch API
│   └── youtube/                # Implementation of youtube API
├── mlua/                    	# lua package handler
│   ├── dyevents.go             # DyEvent implementation
│   ├── parser.go               # Parser lua table to struct and vice-versa
│   └── mlua.go                 # Implementation
├── sql/                        # SQLite manager
│   ├── core.go                 # Main db usage
│   └── modules.go              # Modules focus db
├── modules/                    # lua modules folder
└── web/                        # web related folder
```


## Code Style Guidelines

### General Principles
- Write clean, readable, and maintainable code
- Follow the single responsibility principle
- Keep functions small and focused (max 30-40 lines when possible)
- Use meaningful variable and function names

### Naming Conventions (Go)
- **Files**: snake_case (e.g., `do_something.go`, `something.go`)
- **Structs**: PascalCase (e.g., `TwitchUser`, `TwitchHandler`)
- **Exported functions/types**: PascalCase (e.g., `MessageQueue`)
- **Unexported functions/types**: camelCase (e.g., `loadConfig`)
- **Constants**: SCREAMING_SNAKE_CASE (e.g., `MAX_QUEUE_SIZE`)
- **Packages**: File/Folder (e.g., `mystreambot/newImplementation`); The import (e.g.,`mystreambot/newImplementation`)
- **Avoid stutter**: `audio.Player` not `audio.AudioPlayer`

## Concurrency Rules
- Communication between goroutines must use channels
- Shared state must be protected (mutex or avoided)
- Single global mutable state

## Dependency Rules
- domain must not import adapters
- services may import domain and ports
- adapters implement ports
- main wires dependencies
- No circular dependencies allowed

## Environment Variables Required

### Config file
Create a `init.txt` file with:

The config should keep in "[]" the name of the config and the parameters as follow:
```
# this is a comment
[Config]
TwitchClientID=
TwitchClientSecret=
KickClientID=
KickClientSecret=
YouTubeClientID=
YouTubeClientSecret=
TwitchScopes=
```

Also it can add custom state data like
```
[State]
State.Data.Something
```

### Twitch Subtypes 
Create a `twitchsubtypes.json` file with all used twitch sub; for example:
```json
{
    "channel.follow": {"version": 2, "requires": "moderator:read:followers"},
	"channel.ban":   {"requires": "channel:moderate"},
	"channel.unban": {"requires": "channel:moderate"},
    "stream.online":  {},
	"stream.offline": {},
	...
}
```

## Stream's data

### Flow

Any Stream API
	  ↓
	IF CHAT → `ChatQueue`→ 		├ `WsBroadcast`
	  │							├ IF COMMAND → `CommandQueue` 	→	├ Send to DyEvent Queue
	  │							├ Send to DyEvent Queue				└ Lua Handler
	  │							└ Lua Handler				
	  ↓
	IF Event → `EventQueue`→	├ `WsBroadcast`
								├ Send to DyEvent Queue
	  							└ Lua Handler	

#### ChatQueu

#### CommandQueue

#### EventQueue
The queue receives the follow struct

```go
type Event struct {
	Type   string 					// Event name/identifyer
	Data   map[string]interface{}
}
```

## SQLite
There are two types of sqlite:

### Key/Value (KV)
- Shared across all modules
- Used for general state and configuration

### Dedicated per module (for customevent only)
- Each module can have its own .db file
- Fully customizable, it can create any table for specific use case

## Lua Modules

### How to implement
For modules implementation see [Module.md](./Module.md)

### Program workflow
There are two types of global variables System and API

#### System (g and ev)
"g" has general functions like: key/value db, consume global state, print, log, etc.
[g](/definitions/g.d.lua)

"ev" has event specific functions like: socket communication, timeout, event state, etc.
[ev](/definitions/ev.d.lua)

#### API
Access of services api like [twitch](/definitions/twitch.d.lua)
