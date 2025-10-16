package globals

import (
	"MyStreamBot/helpers"
	"sync"
)

type TwitchUser struct {
	Token                  string
	UserID                 string `json:"userId"`
	UserLogin              string `json:"userLogin"`
	Connected              bool   `json:"connected"`
	DisplayName            string `json:"display_name"`
	Type                   string `json:"type"`
	BroadcasterType        string `json:"broadcaster_type"`
	Description            string `json:"description"`
	ProfileImageURL        string `json:"profile_image_url"`
	ProfileOfflineImageURL string `json:"offline_image_url"`
	ViewCount              int    `json:"view_count"`
	Email                  string `json:"email"`
}

type State struct {
	sync.RWMutex
	ViewersTwitch []string
	Data          map[string]any
	TwitchUser    TwitchUser
}

type Config struct {
	TwitchClientID     string
	TwitchClientSecret string
	KickClientID       string
	KickClientSecret   string
	BotPrefix          string
}

var (
	state     *State
	config    *Config
	once      sync.Once
	onceState sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		config = &Config{
			BotPrefix: "!",
		}
	})
	return config
}

func GetState() *State {
	onceState.Do(func() {
		state = &State{
			Data: make(map[string]any),
		}
		helpers.Log(helpers.Blue, "State iniciado...")
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
	s.TwitchUser = user
	s.Unlock()
}
