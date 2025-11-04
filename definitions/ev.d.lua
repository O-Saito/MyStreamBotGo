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
function ev.getInterval()end

---Set interval in seconds
---@param seconds number
function ev.setInterval(seconds) end

---Check if the module is paused
---@return boolean
function ev.isPaused() end

---Set the module paused or unpaused
---@param paused boolean
function ev.setPaused(paused) end

_G.ev = ev