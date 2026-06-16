---@meta
---@class g
local g = {}

---Log text/data into console and file
---@param text string
---@param data table|nil
function g.log(text, data) end

---Print text/data into console 
---@param text string
---@param data table|nil
function g.print(text, data) end

---Send data to the web socket server
---@param type string
---@param data table
function g.socket_send(type, data) end

---Send a message to a channel
---@param source string
---@param channel string
---@param message string
---@param reply string|nil
function g.send_message(source, channel, message, reply) end

---Get a value from the bot state
---@param key string
---@return any
function g.get(key)end

---Set a value in the bot state
---@param key string
---@param value any
function g.set(key, value)end

---Get a value from the bot global db
---@param key string
---@return any
function g.kv_get(key)end

---Set a value in the bot global db
---@param key string
---@param value any
function g.kv_set(key, value)end

-- declare global
_G.g = g