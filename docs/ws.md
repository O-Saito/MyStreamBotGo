# WebSocket Communication

## Overview

MyStreamBot uses WebSockets for real-time bidirectional communication between the frontend and backend. The server runs on `/ws` endpoint and uses the gorilla/websocket library.

## Connection Flow

1. Client opens a WebSocket connection to `ws://${location.host}/ws`.
2. On open, the client sends the raw string `"init"` to request initial state.
3. The server assigns the connection a tag and replies with an `init` message containing full application state.
4. Client handles further messages via its `onmessage` handler; `web/wrapper.js` also auto-reconnects on disconnect with exponential backoff.

```js
let ws = new WebSocket(`ws://${location.host}/ws`);
ws.onopen = function () {
    ws.send("init");
}
```

## Message Format

### Message Envelope

Every message (client→server or server→client) is JSON with these fields:

| JSON field | Type | Purpose |
|---|---|---|
| `type` | string | Message type (e.g. `"init"`, `"response-upgrade"`) |
| `data` | any | Payload data |
| `filter` | string | Targets the message to connections subscribed via `upgrade-conn` with a matching module name |
| `respond` | number | Client tag; used for direct (single-client) responses |
| `responseClientID` | string | Correlation ID tying a response back to the request that triggered it |

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

- **`init`** — returns full application state (see the `init` message under Server-to-Client Message Types below).
- **`upgrade-conn`** — subscribes the connection to a filter so it receives targeted broadcasts (see Client-to-Server Request Types below). The special filter value `"ignore-broadcast"` opts the connection out of all broadcasts instead.

## Broadcast Rules

The server routes every outgoing message according to its fields:

1. **Filtered:** if `filter` is set, the message is sent only to connections that called `upgrade-conn` with that same filter value.
2. **Direct response:** if `respond` (the client tag) is set to a specific client's tag, the message is sent only to that client.
3. **Broadcast:** otherwise, the message is sent to every connected client (except ones that opted out via `"ignore-broadcast"`).

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

### `send(type, data, func)`

Sends a message to the server. The optional `func` callback receives a response from the server for request-response patterns.

```js
w.send("my_event", { key: "value" });
w.send("my_event", { key: "value" }, (response) => {
    console.log("Response:", response);
});
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

### `getEmotes()`

Returns the current emote map (BTTV/FFZ/7TV/YouTube emojis) as `{ "code": "url", ... }`.

```js
const emotes = w.getEmotes();
```

### `loadEmote(twitchId, login)`

Loads emotes from BTTV, FFZ, 7TV, and YouTube for a given channel. Returns a promise resolving to the updated emote map.

```js
await w.loadEmote("12345", "channel_login");
```

## CustomEvents Communication

### Flow

1. Server loads CustomEvent Lua modules and exposes them via `init` message
2. Frontend receives `custom_events_modules` array in init data
3. Frontend calls `subscribe("module_name.lua")` to upgrade the connection to that filter
4. Lua CustomEvent uses `ev.socket_send(type, data)` to send messages
5. Server routes the message to connections subscribed to that filter
6. Frontend handler receives message via registered `on()` callback

### Filter Matching

The `Filter` field in messages must match the module filename (including extension) that was used in `subscribe()`.

```lua
-- Lua side (CustomEvent)
ev.socket_send("update", { score = 100 })
-- Server sends with Filter = "vote.lua"
-- Frontend subscribed to "vote.lua" receives it
```

### Request/response convention (`respond()`)

When the frontend calls `subscribe(mod).send(type, data, callback)`, the outgoing message carries `respond`/`responseClientID` for correlation. On the Lua side, `on_request(type, data)` handlers can call the injected `respond(data)` function to reply:

```lua
function on_request(type, data)
    if type == "get_score" then
        respond({ score = 42 })
    end
end
```

The server echoes the original `respond`/`responseClientID`, keeps `filter` as the module name, and sends back `type: "return-" + <original type>`:

```json
{ "type": "return-get_score", "filter": "vote.lua", "data": { "score": 42 }, "respond": 1, "responseClientID": "..." }
```

The `send(type, data, func)` callback on the JS side is invoked when a message with `responseClientID` matching the original request arrives — this is what ties `return-<type>` back to the original `send()` call.

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

These are message types sent from the frontend to the server via WebSocket.

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

Connects to YouTube live chat. Returns current live broadcasts. `channel` is required (the handler silently no-ops if it's missing or not a string):

```json
{ "type": "connect-chat-youtube", "data": { "channel": "channel_name" } }
```

**Response:** `result-connect-chat-youtube` with live broadcast list

### `send-chat-message`

Sends a chat message to all connected platform channels:

```json
{ "type": "send-chat-message", "data": { "text": "Hello world!" } }
```

### `get-streamer-data`

Gets current streamer's data:

```json
{ "type": "get-streamer-data", "data": {} }
```

**Response:** `result-get-stream-data` with stream information

### `get-next-streams-youtube`

Gets next scheduled YouTube streams from polling and merges with preview lives:

```json
{ "type": "get-next-streams-youtube", "data": { "channel": "channel_name" } }
```

**Response:** `result-get-next-streams-youtube` with merged preview list

### `connect-to-preview-youtube`

Connects to a YouTube live preview chat:

```json
{ "type": "connect-to-preview-youtube", "data": { "liveChatId": "..." } }
```

**Response:** `result-connect-chat-youtube` with updated connected chats

### `get-dy-statistics`

Gets dynamic event (CustomEvent module) statistics:

```json
{ "type": "get-dy-statistics", "data": {} }
```

**Response:** `result-get-dy-statistics` with event list

### `get-state`

Gets the full bot state object:

```json
{ "type": "get-state", "data": {} }
```

**Response:** `result-get-state` with state data

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

### `result-get-stream-data`

Response to `get-streamer-data` request:

```json
{
  "type": "result-get-stream-data",
  "data": { "twitch": { ... }, "youtube": [ ... ] },
  "respond": 1
}
```

### `twitch-connection`

Sent when the Twitch user/session changes or stream details update:

```json
{ "type": "twitch-connection", "data": { ... TwitchUser ... } }
```

### `twitch-chat-connection`

Sent when the bot's own login joins a Twitch chat channel (i.e. it just connected to that channel):

```json
{ "type": "twitch-chat-connection", "data": { "name": "channel", "id": "channel" } }
```

### `twitch-user-join` / `twitch-user-part`

Sent when a user joins/parts a Twitch chat channel (IRC JOIN/PART):

```json
{ "type": "twitch-user-join", "data": { "user": "username", "channel": "channel", "metadata": { ... }, "color": "#RRGGBB" } }
```

### `clear-chat`

Sent on a Twitch CLEARCHAT IRC event (chat cleared or a user's messages purged):

```json
{ "type": "clear-chat", "data": { "channel": "channel", "metadata": { ... } } }
```

### `self-message`

Sent instead of `user-message` when the message originated from the bot itself (`MessageId == "self"`):

```json
{ "type": "self-message", "data": { ... MessageFromStream ... } }
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
