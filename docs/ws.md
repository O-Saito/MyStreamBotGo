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

### Broadcast Logic (`goweb/server.go:195-233` and `handlers.go`)

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

## Client-to-Server Request Types

These are message types sent from the frontend to the server via WebSocket. The server uses `SocketHandlers` map to route these messages.

### `init`

Requests initial application state. Send as raw string (not JSON):

```
init
```

**Response:** `init` message with full application state

### `upgrade-conn`

Subscribes the connection to a specific CustomEvent filter:

```json
{ "type": "upgrade-conn", "data": { "conn": "module_name" } }
```

Special value `"ignore-broadcast"` makes the client ignore all broadcasts:

```json
{ "type": "upgrade-conn", "data": { "conn": "ignore-broadcast" } }
```

**Response:** `response-upgrade` message

### `connect-chat-kick`

Connects to a Kick chat room:

```json
{ "type": "connect-chat-kick", "data": { "roomId": "12345", "channel": "channel_name" } }
```

### `connect-chat-twitch`

Connects to a Twitch chat channel:

```json
{ "type": "connect-chat-twitch", "data": { "channel": "channel_name" } }
```

### `connect-chat-youtube`

Connects to YouTube live chat. Returns current live broadcasts:

```json
{ "type": "connect-chat-youtube", "data": {} }
```

**Response:** `result-connect-chat-youtube` with live broadcast list

### `send-chat-message`

Sends a chat message to all connected platform channels:

```json
{ "type": "send-chat-message", "data": { "text": "Hello world!" } }
```

### `query-stream-game`

Queries Twitch for game information:

```json
{ "type": "query-stream-game", "data": { "q": "Minecraft" } }
```

**Response:** `result-query-stream-games` with game list

### `get-streamer-data`

Gets current streamer's data:

```json
{ "type": "get-streamer-data", "data": {} }
```

**Response:** `result-get-streamer-data` with stream information

### Custom Event Types

Custom events sent with a `filter` field are forwarded to the Lua CustomEvent module:

```json
{ "type": "event_type", "filter": "module_name.lua", "data": { ... } }
```

The Lua module receives these via `on_request(type, data)` handler.

## Server-to-Client Message Types

These are messages broadcast from the server to connected clients via `globals.WsBroadcast`.

### `init`

Sent in response to client `init` request. Contains full application state:

```json
{
  "type": "init",
  "data": {
    "twitch": { ... },
    "youtube": { ... },
    "kick": { "connected_as": "..." },
    "twitch_connected_chat": [...],
    "kick_connected_chat": [...],
    "youtube_connected_chat": [...],
    "custom_events_modules": [...],
    "twitch_eventsubs": [...]
  },
  "respond": 1
}
```

### `response-upgrade`

Response to `upgrade-conn` request:

```json
{ "type": "response-upgrade", "data": "conexão atualizada!" }
```

### `user-message`

Chat message from any streaming platform. Sent when a user sends a message in chat:

```json
{
  "type": "user-message",
  "data": {
    "source": "twitch|kick|youtube",
    "channel": "channel_name",
    "userId": "12345",
    "user": "username",
    "messageId": "msg_123",
    "message": "Hello!",
    "metadata": { ... }
  }
}
```

### `result-connect-chat-youtube`

Response to `connect-chat-youtube` request:

```json
{
  "type": "result-connect-chat-youtube",
  "data": [...],
  "respond": 1
}
```

### `result-query-stream-games`

Response to `query-stream-game` request:

```json
{
  "type": "result-query-stream-games",
  "data": { "list": [...] },
  "respond": 1
}
```

### `result-get-streamer-data`

Response to `get-streamer-data` request:

```json
{
  "type": "result-get-streamer-data",
  "data": { ... },
  "respond": 1
}
```

### `kick-connection`

Sent when Kick WebSocket connection is established:

```json
{ "type": "kick-connection", "data": "user_login" }
```

### `kick-chat-connection`

Sent when successfully subscribed to a Kick channel chat:

```json
{ "type": "kick-chat-connection", "data": { "name": "channel_slug", "id": "channel_id" } }
```

### `user-message-delete`

Sent when a message is deleted on Kick:

```json
{ "type": "user-message-delete", "data": { "messageId": "msg_123" } }
```

### `youtube-connection`

Sent when YouTube authentication completes:

```json
{ "type": "youtube-connection", "data": { "ChannelName": "...", "Token": "...", ... } }
```

### `youtube-live-offline`

Sent when a YouTube live goes offline:

```json
{ "type": "youtube-live-offline", "data": { "liveId": "...", "offlineAt": "..." } }
```

### `twitch-eventsub-session-welcome`

Sent when Twitch EventSub session is established:

```json
{
  "type": "twitch-eventsub-session-welcome",
  "data": {
    "payload": { "session": { "id": "...", ... } },
    "metadata": { ... }
  }
}
```

### `twitch-eventsub-keepalive`

Sent periodically as Twitch EventSub keepalive:

```json
{
  "type": "twitch-eventsub-keepalive",
  "data": {
    "payload": { ... },
    "metadata": { ... }
  }
}
```

### `twitch-eventsub-notification`

Sent when a Twitch EventSub event occurs. The `payload.subscription.type` contains the event type:

```json
{
  "type": "twitch-eventsub-notification",
  "data": {
    "payload": {
      "subscription": { "type": "channel.follow", ... },
      "event": { ... }
    },
    "metadata": { ... }
  }
}
```

Common event types (handled by `twitchNotificationHandlers` on frontend):
- `channel.follow` - New follower
- `channel.ban` - User banned
- `channel.unban` - User unbanned
- `channel.subscribe` - New subscription
- `channel.cheer` - User cheered
- `channel.raid` - Raid event
- `stream.online` - Stream went live
- `stream.offline` - Stream went offline

### Dynamic/Event-based Messages

Messages from the event queue processor. The `Type` field matches the event name:

```json
{ "type": "event_type", "data": { ... } }
```

### CustomEvent Messages

Messages sent from Lua CustomEvent modules via `ev.socket_send(type, data)`:

```json
{ "type": "update", "filter": "vote.lua", "data": { "score": 100 } }
```

The `filter` field contains the module filename, allowing frontend to subscribe to specific module events.

## HTTP Endpoints

### Web

- `GET /` - Serves static files from `./web` directory

### WebSocket

- `GET /ws` - WebSocket endpoint for real-time communication

### Authentication Callbacks

- `GET /twitch/login` - Redirects to Twitch OAuth
- `GET /twitch/callback` - Twitch OAuth callback
- `GET /youtube/login` - Redirects to YouTube OAuth
- `GET /youtube/callback` - YouTube OAuth callback
- `GET /kick/login` - Redirects to Kick OAuth
- `GET /kick/callback` - Kick OAuth callback

### Administration

- `POST /admin/delete/twitch` - Delete a Twitch message
  - Body: `{ "message": "message_id" }`
- `POST /admin/ban/twitch` - Ban a Twitch user
  - Body: `{ "userId": "123", "duration": 600, "reason": "violation" }`
- `POST /admin/automod/twitch` - Approve/deny AutoMod message
  - Body: `{ "userId": "123", "msgId": "msg_123", "action": "ALLOW|DENY" }`
