package globals

import (
	"MyStreamBot/helpers"
	"MyStreamBot/sql"
	"sync"
	"sync/atomic"
)

type TwitchStreamData struct {
	ID           string   `json:"id"`
	UserId       string   `json:"user_id"`
	UserLogin    string   `json:"user_login"`
	UserName     string   `json:"user_name"`
	GameId       string   `json:"game_id"`
	GameName     string   `json:"game_name"`
	Type         string   `json:"type"`
	Title        string   `json:"title"`
	Tags         []string `json:"tags"`
	ViewerCount  int32    `json:"viewer_count"`
	StartedAt    string   `json:"started_at"`
	Language     string   `json:"language"`
	ThumbnailURL string   `json:"thumbnail_url"`
	IsMature     bool     `json:"is_mature"`
}

type TwitchUser struct {
	Token                  string `json:"-"`
	UserID                 string `json:"userId"`
	UserLogin              string `json:"userLogin"`
	Connected              bool   `json:"connected"`
	DisplayName            string `json:"display_name"`
	Type                   string `json:"type"`
	BroadcasterType        string `json:"broadcaster_type"`
	Description            string `json:"description"`
	ProfileImageURL        string `json:"profile_image_url"`
	ProfileOfflineImageURL string `json:"offline_image_url"`
	//ViewCount              int               `json:"view_count"`
	Email         string            `json:"email"`
	StreamDetails *TwitchStreamData `json:"stream_details"`
}

type YouTubeChannel struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CustomURL   string   `json:"customUrl,omitempty"`
	Thumbnail   string   `json:"thumbnail,omitempty"`
	ChatIDs     []string `json:"chat-ids"`
}

type YouTubeUser struct {
	Token          string `json:"-"`
	RefreshToken   string `json:"-"`
	TokenExpiresIn int
	Channels       []YouTubeChannel `json:"channels"`
}

type State struct {
	sync.RWMutex
	ViewersTwitch           []string
	Data                    map[string]any
	TwitchUser              TwitchUser
	TwitchEventSubSessionId string
	YouTubeUser             YouTubeUser
}

type Config struct {
	sync.RWMutex
	TwitchClientID      string
	TwitchClientSecret  string
	KickClientID        string
	KickClientSecret    string
	YouTubeClientID     string
	YouTubeClientSecret string
	BotPrefix           string
	TwitchScopes        string
	HTTPPort            string
	DBName              string
	TwitchSubTypes      map[string]map[string]any

	YouTubeRefresh string
}

var (
	state     *State
	config    *Config
	once      sync.Once
	onceState sync.Once
	db        atomic.Pointer[sql.CoreDB]
)

func GetConfig() *Config {
	once.Do(func() {
		config = &Config{
			BotPrefix:    "!",
			HTTPPort:     "1699",
			TwitchScopes: "",
			DBName:       "main",
		}
	})
	return config
}

func GetGlobalDB() *sql.CoreDB {
	return db.Load()
}

func SetGlobalDB(ndb *sql.CoreDB) {
	db.Store(ndb)
}

func (c *Config) GetTwitchSubTypes() map[string]map[string]any {
	c.RLock()
	defer c.RUnlock()
	return c.TwitchSubTypes
}

func (c *Config) SetTwitchSubTypes(data map[string]map[string]any) {
	c.Lock()
	defer c.Unlock()
	c.TwitchSubTypes = data
}

func GetState() *State {
	onceState.Do(func() {
		state = &State{
			Data: make(map[string]any),
		}
		helpers.Log(helpers.INFO, "State iniciado...")
	})

	return state
}

func (s *State) AddTwitchViewer(viewer string) {
	s.Lock()
	defer s.Unlock()
	s.ViewersTwitch = append(s.ViewersTwitch, viewer)
}

func (s *State) GetViewerList() []string {
	s.RLock()
	defer s.RUnlock()
	return s.ViewersTwitch
}

func (s *State) GetTwitchEventSubId() string {
	s.RLock()
	defer s.RUnlock()
	return s.TwitchEventSubSessionId
}

func (s *State) SetTwitchEventSubId(id string) {
	s.Lock()
	defer s.Unlock()
	s.TwitchEventSubSessionId = id
}

func (s *State) GetData(key string) any {
	s.RLock()
	defer s.RUnlock()
	return s.Data[key]
}

func (s *State) SetData(key string, value any) {
	s.Lock()
	defer s.Unlock()
	s.Data[key] = value
}

func (s *State) GetTwitchUser() TwitchUser {
	s.Lock()
	defer s.Unlock()
	return s.TwitchUser
}

func (s *State) SetTwitchUser(user TwitchUser) {
	s.Lock()
	defer s.Unlock()
	s.TwitchUser = user
	WsBroadcast <- SocketMessage{
		Type: "twitch-connection",
		Data: s.TwitchUser,
	}
}

func (s *State) SetTwitchStreamDetails(sd TwitchStreamData) {
	s.Lock()
	defer s.Unlock()
	s.TwitchUser.StreamDetails = &sd
	WsBroadcast <- SocketMessage{
		Type: "twitch-connection",
		Data: s.TwitchUser,
	}
}

func (s *State) GetYouTubeUser() YouTubeUser {
	s.Lock()
	defer s.Unlock()
	return s.YouTubeUser
}

func (s *State) SetYouTubeUser(user YouTubeUser) {
	s.Lock()
	defer s.Unlock()
	s.YouTubeUser = user
	WsBroadcast <- SocketMessage{
		Type: "youtube-connection",
		Data: s.YouTubeUser,
	}
}

func (s *State) AddYouTubeChannel(channel YouTubeChannel) {
	s.Lock()
	defer s.Unlock()
	s.YouTubeUser.Channels = append(s.YouTubeUser.Channels, channel)
	WsBroadcast <- SocketMessage{
		Type: "youtube-connection",
		Data: s.YouTubeUser,
	}
}

func (s *State) AddYouTubeChat(channelId, chatId string) {
	s.Lock()
	defer s.Unlock()
	for i := range s.YouTubeUser.Channels {
		c := &s.YouTubeUser.Channels[i]
		if c.ID == channelId {
			helpers.Log(helpers.DEBUG, "FOUND!")
			c.ChatIDs = append(c.ChatIDs, chatId)
			break
		}
	}
	WsBroadcast <- SocketMessage{
		Type: "youtube-connection",
		Data: s.YouTubeUser,
	}
}
