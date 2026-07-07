---@class TwitchState
---@field UserID string
---@field UserLogin string
---@field DisplayName string
---@field Type string
---@field BroadcasterType string
---@field Description string
---@field ProfileImageURL string
---@field ProfileOfflineImageURL string

---@class TwitchStream
---@field ID string
---@field GameId string
---@field GameName string
---@field Type string
---@field Title string
---@field Tags string[]
---@field ViewerCount number
---@field StartedAt string
---@field Language string
---@field ThumbnailURL string
---@field IsMature boolean

---@class UserData
---@field ID string
---@field Login string
---@field DisplayName string
---@field Type string
---@field BroadcasterType string
---@field Description string
---@field ProfileImageURL string
---@field ProfileOfflineImageURL string
---@field ViewCount number
---@field Email string
---@field CreatedAt string

---@class TwitchViewerData
---@field UserId string
---@field UserName string
---@field UserLogin string
---@field FollowedAt string

---@class StreamData
---@field BroadcasterId string
---@field BroadcasterLogin string
---@field BroadcasterName string
---@field BroadcasterLanguage string
---@field GameID string
---@field GameName string
---@field Title string
---@field Delay number
---@field Tags string[]
---@field ContentClassificationLabels string[]
---@field IsBrandedContent boolean

---@class TwitchUserAuthorization
---@field ClientID string
---@field UserID string
---@field Scopes string[]
---@field ExpiresAt string

---@class TwitchUserBlock
---@field UserId string
---@field UserLogin string
---@field UserName string

---@class TwitchUserExtension
---@field CanActivate boolean
---@field ID string
---@field Name string
---@field Version string
---@field Type string[]

---@class TwitchActiveExtensions
---@field panel table
---@field overlay table
---@field component table
---@field video_overlay table
---@field chat table

---@class TwitchClip
---@field ID string
---@field URL string
---@field EmbedURL string
---@field BroadcasterID string
---@field BroadcasterName string
---@field BroadcasterLogin string
---@field CreatorID string
---@field CreatorName string
---@field CreatorLogin string
---@field VideoID string
---@field GameID string
---@field Language string
---@field Title string
---@field ViewCount number
---@field CreatedAt string
---@field Duration number
---@field ThumbnailURL string

---@meta
---@class twitch
local twitch = {}

---Get cached Twitch user state
---@return TwitchState
function twitch.get_state() end

---Get cached stream data
---@return TwitchStream
function twitch.get_cache_stream() end

---Get color of user's chat messages
---@param username string
function twitch.get_cache_user_chat_color(username) end

---Get user data by username
---@param username string
---@return UserData|nil
function twitch.get_user_data(username) end

---Get user data by user ID
---@param userid string
---@return UserData|nil
function twitch.get_user_data_by_id(userid) end

---Get followers data of logged streamer by user ID
---@param userid string
---@return TwitchViewerData[]|nil
function twitch.get_follower_data(userid) end

---Get followers data of logged streamer
---@return TwitchViewerData[]|nil
function twitch.get_followers_data() end

---Get channel stream data by user ID
---@param userId string
---@return StreamData
function twitch.get_channel_stream_data(userId) end

---Get user authorization by user ID
---@param userId string
---@return TwitchUserAuthorization|nil
function twitch.get_authorization_by_user(userId) end

---Get user block list
---@param targetUserID string
---@return table|nil
function twitch.get_user_block_list(targetUserID) end

---Get user extensions
---@return TwitchUserExtension[]
function twitch.get_user_extensions() end

---Get user active extensions
---@return TwitchActiveExtensions|nil
function twitch.get_user_active_extensions() end

---Create a clip
---@param broadcasterID string
---@param title string
---@param duration number
---@return TwitchClip[]
function twitch.create_clip(broadcasterID, title, duration) end

---Get clips
---@param broadcasterID string
---@param gameID string
---@param ... string
---@return TwitchClip[]
function twitch.get_clips(broadcasterID, gameID, ...) end

---Delete a chat message
---@param msgID string
function twitch.delete_message(msgID) end

---Ban a user
---@param userId string
---@param duration number
---@param reason string
---@return string|nil
function twitch.ban_user(userId, duration, reason) end

-- declare global
_G.twitch = twitch