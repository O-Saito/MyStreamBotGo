package kick

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"MyStreamBot/globals"
)

type kickMockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *kickMockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func setupKickMockTransport(resp *http.Response) func() {
	orig := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: &kickMockRoundTripper{response: resp}}
	return func() { http.DefaultClient = orig }
}

func TestGetUser_Success(t *testing.T) {
	cfg := globals.GetConfig()
	cfg.KickClientID = "test_client_id"

	restore := setupKickMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"data": [{
				"user_id": 12345,
				"email": "test@example.com",
				"name": "testuser",
				"profile_picture": "https://example.com/image.png"
			}]
		}`)),
	})
	defer restore()

	user, err := GetUser("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserId != 12345 {
		t.Errorf("user.UserId = %d, want 12345", user.UserId)
	}
	if user.Name != "testuser" {
		t.Errorf("user.Name = %q, want %q", user.Name, "testuser")
	}
	if user.Email != "test@example.com" {
		t.Errorf("user.Email = %q, want %q", user.Email, "test@example.com")
	}
}

func TestGetUser_NotFound(t *testing.T) {
	cfg := globals.GetConfig()
	cfg.KickClientID = "test_client_id"

	restore := setupKickMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{"data":[]}`)),
	})
	defer restore()

	_, err := GetUser("nonexistent")
	if err == nil {
		t.Error("expected error for user not found, got nil")
	}
}

func TestGetChannel_Success(t *testing.T) {
	cfg := globals.GetConfig()
	cfg.KickClientID = "test_client_id"

	restore := setupKickMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"data": [{
				"broadcaster_user_id": 12345,
				"slug": "testchannel",
				"channel_description": "Test channel",
				"stream": {
					"is_live": true,
					"key": "testchannel",
					"language": "en",
					"viewer_count": 500,
					"is_mature": false
				},
				"stream_title": "Test Stream",
				"category": {"id":1,"name":"Just Chatting"}
			}]
		}`)),
	})
	defer restore()

	channel, err := GetChannel(12345, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel.Slug != "testchannel" {
		t.Errorf("channel.Slug = %q, want %q", channel.Slug, "testchannel")
	}
	if !channel.Stream.IsLive {
		t.Error("channel.Stream.IsLive should be true")
	}
	if channel.Stream.ViewerCount != 500 {
		t.Errorf("channel.Stream.ViewerCount = %d, want 500", channel.Stream.ViewerCount)
	}
	if channel.Category.Name != "Just Chatting" {
		t.Errorf("category.Name = %q, want %q", channel.Category.Name, "Just Chatting")
	}
}

func TestGetChannel_NotFound(t *testing.T) {
	cfg := globals.GetConfig()
	cfg.KickClientID = "test_client_id"

	restore := setupKickMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{"data":[]}`)),
	})
	defer restore()

	_, err := GetChannel(99999, nil)
	if err == nil {
		t.Error("expected error for channel not found, got nil")
	}
}

func TestGetChannelStream_Offline(t *testing.T) {
	cfg := globals.GetConfig()
	cfg.KickClientID = "test_client_id"

	restore := setupKickMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"data": [{
				"broadcaster_user_id": 12345,
				"slug": "testchannel",
				"channel_description": "Test channel",
				"stream": {
					"is_live": false,
					"key": "testchannel",
					"language": "en",
					"viewer_count": 0,
					"is_mature": false
				},
				"stream_title": "",
				"category": {"id":1,"name":""}
			}]
		}`)),
	})
	defer restore()

	channel, err := GetChannel(12345, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel.Stream.IsLive {
		t.Error("channel.Stream.IsLive should be false for offline stream")
	}
	if channel.Stream.ViewerCount != 0 {
		t.Errorf("channel.Stream.ViewerCount = %d, want 0", channel.Stream.ViewerCount)
	}
}

func TestGetChatroom_Success(t *testing.T) {
	cfg := globals.GetConfig()
	cfg.KickClientID = "test_client_id"

	restore := setupKickMockTransport(&http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`{
			"id": 12345,
			"pinned_message": "Welcome!"
		}`)),
	})
	defer restore()

	chatroom, err := GetChatroom("testchannel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatroom.ID != 12345 {
		t.Errorf("chatroom.ID = %d, want 12345", chatroom.ID)
	}
}

func TestKickMessageStruct(t *testing.T) {
	msg := KickMessage{
		Event:   "App\\Events\\ChatMessageEvent",
		Channel: "testchannel",
		Data:    `{"content":"Hello"}`,
	}

	if msg.Event != "App\\Events\\ChatMessageEvent" {
		t.Errorf("KickMessage.Event = %q", msg.Event)
	}
}

func TestIrcChannelStruct(t *testing.T) {
	ch := IrcChannel{
		ID:        "12345",
		Slug:      "testchannel",
		Connected: false,
	}

	if ch.ID != "12345" {
		t.Errorf("IrcChannel.ID = %q", ch.ID)
	}
	if ch.Slug != "testchannel" {
		t.Errorf("IrcChannel.Slug = %q", ch.Slug)
	}
}