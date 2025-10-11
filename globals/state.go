package globals

import (
	"sync"
)

type State struct {
	sync.RWMutex
	ViewersTwitch []string
	Data          map[string]any
}

type Config struct {
	TwitchClientID     string
	TwitchClientSecret string
	KickClientID       string
	KickClientSecret   string
	BotPrefix          string
}

var (
	state  *State
	config *Config
	once   sync.Once
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
	once.Do(func() {
		state = &State{
			Data: make(map[string]any),
		}
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
