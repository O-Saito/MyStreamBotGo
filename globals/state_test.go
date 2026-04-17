package globals

import (
	"testing"
	"sync"
)

func TestState_GetData_SetData(t *testing.T) {
	s := &State{
		Data: make(map[string]any),
	}

	t.Run("set_and_get", func(t *testing.T) {
		s.SetData("key1", "value1")
		got := s.GetData("key1")
		if got != "value1" {
			t.Errorf("GetData(%q) = %q, want %q", "key1", got, "value1")
		}
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		got := s.GetData("nonexistent")
		if got != nil {
			t.Errorf("GetData(%q) = %v, want nil", "nonexistent", got)
		}
	})

	t.Run("overwrite", func(t *testing.T) {
		s.SetData("key1", "value1")
		s.SetData("key1", "value2")
		got := s.GetData("key1")
		if got != "value2" {
			t.Errorf("GetData(%q) = %q, want %q", "key1", got, "value2")
		}
	})
}

func TestState_GetTwitchUser_SetTwitchUser(t *testing.T) {
	s := &State{
		Data: make(map[string]any),
	}

	t.Run("set_and_get", func(t *testing.T) {
		user := TwitchUser{
			UserID:    "12345",
			UserLogin: "testuser",
			Token:    "oauth:test",
		}
		s.SetTwitchUser(user)
		got := s.GetTwitchUser()
		if got.UserID != user.UserID || got.UserLogin != user.UserLogin {
			t.Errorf("GetTwitchUser() = %+v, want %+v", got, user)
		}
	})
}

func TestState_AddTwitchViewer_GetViewerList(t *testing.T) {
	s := &State{
		ViewersTwitch: []string{},
		Data:          make(map[string]any),
	}

	t.Run("add_single", func(t *testing.T) {
		s.AddTwitchViewer("viewer1")
		list := s.GetViewerList()
		if len(list) != 1 || list[0] != "viewer1" {
			t.Errorf("GetViewerList() = %v, want [viewer1]", list)
		}
	})

	t.Run("add_multiple", func(t *testing.T) {
		s.AddTwitchViewer("viewer2")
		list := s.GetViewerList()
		if len(list) != 2 {
			t.Errorf("GetViewerList() len = %d, want 2", len(list))
		}
	})
}

func TestState_GetYouTubeUser_SetYouTubeUser(t *testing.T) {
	s := &State{
		Data: make(map[string]any),
	}

	t.Run("set_and_get", func(t *testing.T) {
		user := YouTubeUser{
			Token:        "yt_token",
			RefreshToken: "yt_refresh",
		}
		s.SetYouTubeUser(user)
		got := s.GetYouTubeUser()
		if got.Token != user.Token {
			t.Errorf("GetYouTubeUser().Token = %q, want %q", got.Token, user.Token)
		}
	})
}

func TestState_AddYouTubeChannel(t *testing.T) {
	s := &State{
		Data:        make(map[string]any),
		YouTubeUser: YouTubeUser{},
	}

	t.Run("add_channel", func(t *testing.T) {
		channel := YouTubeChannel{
			ID:    "channel_id",
			Title: "Channel Title",
		}
		s.AddYouTubeChannel(channel)
		got := s.GetYouTubeUser()
		if len(got.Channels) != 1 || got.Channels[0].ID != "channel_id" {
			t.Errorf("GetYouTubeUser().Channels = %+v", got.Channels)
		}
	})
}

func TestState_GetTwitchEventSubId_SetTwitchEventSubId(t *testing.T) {
	s := &State{
		Data: make(map[string]any),
	}

	t.Run("set_and_get", func(t *testing.T) {
		s.SetTwitchEventSubId("session_123")
		got := s.GetTwitchEventSubId()
		if got != "session_123" {
			t.Errorf("GetTwitchEventSubId() = %q, want %q", got, "session_123")
		}
	})
}

func TestState_Concurrency(t *testing.T) {
	s := &State{
		Data: make(map[string]any),
	}

	t.Run("concurrent_rw", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(2)
			go func(i int) {
				defer wg.Done()
				s.SetData("key", i)
			}(i)
			go func() {
				defer wg.Done()
				s.GetData("key")
			}()
		}
		wg.Wait()
	})
}