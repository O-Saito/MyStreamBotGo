package globals

import (
	"MyStreamBot/helpers"
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
