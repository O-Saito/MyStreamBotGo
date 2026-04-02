---@meta
---@class ev
---@field data table|nil
local ev = {}

---Send socket message
---@param type string
---@param data table
function ev.socket_send(type, data) end

---Get current interval in seconds
---@return number
function ev.get_interval()end

---Set interval in seconds
---@param seconds number
function ev.set_interval(seconds) end

---Check if the module is paused
---@return boolean
function ev.is_paused() end

---Set the module paused or unpaused
---@param paused boolean
function ev.set_paused(paused) end

---Tell the module to use the database connection
function ev.use_db() end

---Execute a raw SQL query on the database connection, returning the number of affected rows
---@param query string
---@return number
function ev.db_exec(query) end

---Query the database
---@param query string
---@return table
function ev.db_query(query) end

_G.ev = ev