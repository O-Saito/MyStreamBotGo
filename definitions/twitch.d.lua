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

---@class StreamData
---@field BroadcasterId               string  
---@field BroadcasterLogin            string   
---@field BroadcasterName             string  
---@field BroadcasterLanguage         string  
---@field GameID                      string  
---@field GameName                    string   
---@field Title                       string   
---@field Delay                       int      
---@field Tags                        []string 
---@field ContentClassificationLabels []string 
---@field IsBrandedContent            bool     

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
function g.get_follower_data(userid)end

---@return TwitchViewerData[]|nil
function g.get_followers_data()end

---@return StreamData|nil 
function g.get_channel_stream_data()end

-- declare global
_G.twitch = g