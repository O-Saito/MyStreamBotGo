---@class kick.UserData
---@field UserId number
---@field Email string
---@field Name string
---@field ProfilePicture string

---@class kick.ChannelData
---@field BroadcasterUserId number
---@field Slug string
---@field ChannelDescription string
---@field BannerPicture string
---@field Stream kick.ChannelStream
---@field StreamTitle string
---@field Category kick.ChannelCategory

---@class kick.ChannelStream
---@field IsLive boolean
---@field IsMature boolean
---@field Key string
---@field Language string
---@field StartTime string
---@field Thumbnail string
---@field Url string
---@field ViewerCount number

---@class kick.ChannelCategory
---@field Id number
---@field Name string
---@field Thumbnail string

---@class kick.ChatroomData
---@field ID number
---@field PinnedMessage string

---@meta
---@class kick
local kick = {}

---Get user data by user ID
---@param userId string
---@return kick.UserData
function kick.get_user(userId) end

---Get channel data by user ID
---@param userId number
---@return kick.ChannelData
function kick.get_channel(userId) end

---Get chatroom data by channel ID
---@param channelId string
---@return kick.ChatroomData
function kick.get_chatroom(channelId) end

-- declare global
_G.kick = kick