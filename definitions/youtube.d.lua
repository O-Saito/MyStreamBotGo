---@class YouTubeChannel
---@field ID string
---@field Title string
---@field Description string
---@field CustomURL string
---@field Thumbnail string

---@class youtube.YouTubeChannelListResponse
---@field Kind string
---@field Etag string
---@field Items YouTubeChannel[]

---@class youtube.LiveBroadcastListResponse
---@field Kind string
---@field Etag string
---@field Items youtube.LiveBroadcast[]

---@class youtube.LiveBroadcast
---@field ID string
---@field Title string
---@field ConcurrentViewers string
---@field LiveChatID string
---@field ActualStartTime string
---@field ActualEndTime string
---@field ScheduledStartTime string
---@field ScheduledEndTime string

---@meta
---@class youtube
local g = {}

---Get cached YouTube channel list from state
---@return YouTubeChannel[]
function g.get_state() end

---Get current YouTube channel info
---@return youtube.YouTubeChannelListResponse
function g.get_current_youtube_channel() end

---Get current live streams
---@return youtube.LiveBroadcastListResponse
function g.get_current_streamings() end

-- declare global
_G.youtube = g