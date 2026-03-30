# WebSocket Communication

## Overview

MyStreamBot uses WebSockets for real-time bidirectional communication between the frontend and backend. The server runs on `/ws` endpoint and uses the gorilla/websocket library.

## Connection Flow

### Server Side (`goweb/server.go`)

1. HTTP request to `/ws` is upgraded to WebSocket connection
2. Server assigns a unique tag index to each client
3. Client connection is stored in `wsClients` map
4. Server listens for messages in a loop and processes them

```go
http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    // ... assign tag and store client
    for {
        _, msg, err := conn.ReadMessage()
        // ... process message
    }
})
```

### Frontend Side (`web/wrapper.js`)

1. Create WebSocket connection to `ws://${location.host}/ws`
2. Send `"init"` string on open to receive initial state
3. Handle incoming messages via `onmessage` handler
4. Auto-reconnect on disconnect with exponential backoff

```js
let ws = new WebSocket(`ws://${location.host}/ws`);
ws.onopen = function () {
    ws.send("init");
}
```

## Message Format

### SocketMessage Structure

```go
type SocketMessage struct {
    Type    string      // Message type (e.g., "init", "response-upgrade")
    Data    interface{} // Payload data
    Filter  string      // Filter for targeted broadcasts (module name)
    Respond int         // Client tag for direct responses
}
```

### Client to Server Messages

**Init (raw string):**
```
init
```

**JSON Messages:**
```json
{ "type": "upgrade-conn", "data": { "conn": "module_name" } }
{ "type": "event_type", "data": { ... }, "filter": "module_name" }
```

## Server Handlers

### `init` Handler

Returns initial application state to the client:

```go
"init": func(c *websocket.Conn, m map[string]any, mytag int) {
    globals.WsBroadcast <- globals.SocketMessage{
        Type: "init",
        Data: map[string]any{
            "twitch":                   globals.GetState().GetTwitchUser(),
            "youtube":                  globals.GetState().GetYouTubeUser(),
            "kick":                     map[string]any{"connected_as": kick.UserLogin},
            "twitch_connected_chat":    twitch.Channels,
            "kick_connected_chat":      kick.Channels,
            "youtube_connected_chat":   globals.GetState().GetData("youtube-lives"),
            "custom_events_modules":    mlua.ListDynamicEvents(),
            "twitch_eventsubs":         globals.GetState().GetData("TwitchSubEventsConnectedEvents"),
        },
        Respond: mytag,
    }
}
```

### `upgrade-conn` Handler

Upgrades connection to receive filtered broadcasts for specific CustomEvents:

```go
"upgrade-conn": func(c *websocket.Conn, m map[string]any, mytag int) {
    if m["conn"] == "ignore-broadcast" {
        wsClients[c] = -1  // Ignore all broadcasts
        return
    }
    // Add to filtered clients list
    wsClientsUpgraded[m["conn"].(string)] = append(wsClientsUpgraded[m["conn"].(string)], c)
}
```

## Broadcast System

### WsBroadcast Channel

The server uses a global channel `globals.WsBroadcast` to send messages from backend to all connected clients.

### Broadcast Logic (`goweb/server.go:195-233`)

1. **Filtered Broadcasts:** If `msg.Filter` is set, only send to clients subscribed to that filter
2. **Direct Response:** If `msg.Respond` is set (non-zero, non-negative), send only to specific client
3. **Broadcast:** Otherwise, send to all clients (except those with tag -1)

```go
go func() {
    for msg := range globals.WsBroadcast {
        if msg.Filter != "" {
            // Filtered broadcast
            wsList := wsClientsUpgraded[msg.Filter]
            for _, client := range wsList {
                client.WriteMessage(websocket.TextMessage, jsonData)
            }
            continue
        }

        // Direct response or broadcast
        for client, tag := range wsClients {
            if tag == -1 continue  // Skip ignore-broadcast clients
            if msg.Respond != 0 && msg.Respond != -1 && msg.Respond != tag {
                continue  // Skip other clients for direct response
            }
            client.WriteMessage(websocket.TextMessage, jsonData)
        }
    }
}()
```

## Frontend API (`web/wrapper.js`)

### `connect()`

Initializes WebSocket connection. Must be called once.

```js
w.connect();
```

### `subscribe(event)`

Subscribes to a CustomEvent module. Returns an EventSubscriber object with `on()` and `send()` methods.

```js
const vote = w.subscribe("vote.lua");
if (vote) {
    vote.on("user_vote_update", (data) => { /* handle */ });
    vote.send("setup", { options: ["Yes", "No"] });
}
```

### `send(type, data)`

Sends a message to the server.

```js
w.send("my_event", { key: "value" });
```

### `on(eventType, func)`

Registers a handler for a specific event type.

```js
w.on("init", (data) => {
    console.log("Initialized with:", data);
});
```

### `onTwitch(eventType, func)`

Registers a handler for Twitch EventSub notifications.

```js
w.onTwitch("channel.follow", (data) => {
    console.log("New follower:", data);
});
```

### `ignoreBroadcast()`

Makes the connection ignore all broadcast messages.

```js
w.ignoreBroadcast();
```

### `exec(eventType, data)`

Manually triggers a handler (for testing/internal use).

```js
w.exec("my_event", { data: "test" });
```

## CustomEvents Communication

### Flow

1. Server loads CustomEvent Lua modules and exposes them via `init` message
2. Frontend receives `custom_events_modules` array in init data
3. Frontend calls `subscribe("module_name.lua")` to upgrade connection
4. Server adds client to `wsClientsUpgraded[module_name]` map
5. Lua CustomEvent uses `ev.socket_send(type, data)` to send messages
6. Server broadcasts to filtered clients based on Filter field
7. Frontend handler receives message via registered `on()` callback

### Filter Matching

The `Filter` field in messages must match the module filename (including extension) that was used in `subscribe()`.

```lua
-- Lua side (CustomEvent)
ev.socket_send("update", { score = 100 })
-- Server sends with Filter = "vote.lua"
-- Frontend subscribed to "vote.lua" receives it
```

## Message Flow Diagram

```
Frontend                      Server                    Lua/CustomEvent
   |                            |                            |
   |--- connect() ------------->|                            |
   |                            |                            |
   |--- "init" ---------------->|                            |
   |<-- {init data} ------------|                            |
   |                            |                            |
   |--- subscribe("vote.lua")-->|                            |
   |                            |                            |
   |                            |<-- ev.socket_send() ------|
   |                            |                            |
   |<-- {type, filter, data} --|                            |
   |                            |                            |
   |--- vote.on("update", fn)->|                            |
   |                            |                            |
```

## Administrative Endpoints

The server also exposes admin endpoints (`/admin/*`) for moderation actions:

- `POST /admin/delete/twitch` - Delete a message
- `POST /admin/ban/twitch` - Ban a user
- `POST /admin/automod/twitch` - Approve/deny AutoMod message
