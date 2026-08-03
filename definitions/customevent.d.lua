---@meta

---@class CommandData
---@field Name string Command name
---@field Args string[] Array of string arguments
---@field User string Username who triggered the command
---@field Text string Full message text
---@field Data table Additional key-value data
---@field Source string Platform ("twitch", "kick", "youtube")
---@field Channel string Channel name
---@field Message MessageData Original message table

---@class MessageData
---@field Source string Platform
---@field Channel string Channel name
---@field User string Username
---@field UserId string User ID
---@field MessageId string Message ID
---@field Message string Text content
---@field Metadata table Extra metadata map

---@class RequestData
---@field respond fun(response: table) Call to send a response back to the frontend caller

---Called when the module is loaded or hot-reloaded
function _G.on_start() end

---Called periodically if ev.set_interval(n) was set
function _G.on_tick() end

---Called when the server dispatches a named event (e.g. "stream.online", "channel.follow")
---@param name string Event type name
---@param data table Event payload (mirrors EventSub JSON structure)
function _G.on_event(name, data) end

---Called when a chat message is received from any platform
---@param msg MessageData
function _G.on_message(msg) end

---Called when a command !<name> is triggered
---@param name string
---@param data CommandData
function _G.on_command(name, data) end

---Called when the frontend or a plugin sends a request
---@param type string Request type
---@param data RequestData Request payload, includes data.respond() for sending a response
function _G.on_request(type, data) end
