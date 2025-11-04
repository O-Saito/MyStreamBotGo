---@class UserData
---@field ID                     string
---@field Login                  string
---@field DisplayName            string 
---@field Type                   string 
---@field BroadcasterType        string 
---@field Description            string
---@field ProfileImageURL        string 
---@field ProfileOfflineImageURL string 
---@field ViewCount              int    
---@field Email                  string 
---@field CreatedAt              string 

---@class TwitchViewerData
---@field UserId     string
---@field UserName   string
---@field UserLogin  string
---@field FollowedAt string

---@meta
---@class twitch
local g = {}

---Get color of user's chat messages
---@param username string
---@param color string|nil
function g.get_cache_user_chat_color(username) end

---Get user data by username
---@param username string
---@return UserData|nil
function g.get_user_data(username) end

---Get user data by user ID
---@param userid string
---@return UserData|nil
function g.get_user_data_by_id(userid)end

---Get followers data of logged streamer by user ID
---@param userid string
---@return TwitchViewerData[]|nil
function g.get_followers_data(userid)end

-- declare global
_G.twitch = g