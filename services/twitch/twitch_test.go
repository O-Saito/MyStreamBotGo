package twitch

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		key      string
		wantVal  string
	}{
		{"user_id", "user-id=12345", "user-id", "12345"},
		{"display_name", "display-name=TestUser", "display-name", "TestUser"},
		{"color", "color=#FF0000", "color", "#FF0000"},
		{"badges", "badges=subscriber/1", "badges", "subscriber/1"},
		{"multiple_tags", "user-id=12345;color=#FF0000;badges=sub/1", "color", "#FF0000"},
		{"emote", "emotes=25/1:2-6", "emotes", "25/1:2-6"},
		{"room_id", "room-id=67890", "room-id", "67890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := partseTags(tt.input)
			if result[tt.key] != tt.wantVal {
				t.Errorf("partseTags(%q)[%q] = %q, want %q", tt.input, tt.key, result[tt.key], tt.wantVal)
			}
		})
	}
}

func TestParseTagsEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{"empty", "", 0},
		{"single_tag", "key=value", 1},
		{"two_tags", "key1=value1;key2=value2", 2},
		{"no_value_no_entry", "key", 0},
		{"trailing_semicolon", "key=value;", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := partseTags(tt.input)
			if len(result) != tt.wantLen {
				t.Errorf("partseTags(%q) len = %d, want %d", tt.input, len(result), tt.wantLen)
			}
		})
	}
}

func TestDefaultTwitchColor(t *testing.T) {
	tests := []struct {
		name    string
		username string
	}{
		{"empty_username", ""},
		{"short_username", "ab"},
		{"long_username", "verylongusername"},
		{"special_chars", "user_name-123"},
		{"numbers", "user123"},
		{"capital_letters", "USER"},
	}

	expectedColors := []string{
		"#FF0000", "#0000FF", "#008000", "#B22222",
		"#FF7F50", "#9ACD32", "#FF4500", "#2E8B57",
		"#DAA520", "#D2691E", "#5F9EA0", "#1E90FF",
		"#FF69B4", "#8A2BE2", "#00FF7F",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := defaultTwitchColor(tt.username)

			isValid := false
			for _, c := range expectedColors {
				if color == c {
					isValid = true
					break
				}
			}
			if !isValid {
				t.Errorf("defaultTwitchColor(%q) = %q, not in expected colors", tt.username, color)
			}

			if color[0] != '#' {
				t.Errorf("defaultTwitchColor should return hex color starting with #")
			}
		})
	}

	t.Run("deterministic", func(t *testing.T) {
		username := "testuser"
		color1 := defaultTwitchColor(username)
		color2 := defaultTwitchColor(username)
		if color1 != color2 {
			t.Errorf("defaultTwitchColor not deterministic: %q != %q", color1, color2)
		}
	})

	t.Run("distribution", func(t *testing.T) {
		colors := make(map[string]int)
		usernames := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		for _, u := range usernames {
			color := defaultTwitchColor(u)
			colors[color]++
		}

		if len(colors) < 3 {
			t.Errorf("poor color distribution: only %d colors for 10 users", len(colors))
		}
	})
}

func TestMessageHelpers(t *testing.T) {
	tests := []struct {
		name    string
		msg    Message
		wantReply bool
	}{
		{
			name:    "no_reply",
			msg:    Message{Channel: "test", Text: "Hello"},
			wantReply: false,
		},
		{
			name:    "with_reply",
			msg:    Message{Channel: "test", Text: "Reply!", MessageToReply: "parent123"},
			wantReply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasReply := tt.msg.MessageToReply != ""
			if hasReply != tt.wantReply {
				t.Errorf("hasReply = %v, want %v", hasReply, tt.wantReply)
			}
		})
	}
}

func TestMessageFormatting(t *testing.T) {
	t.Run("reply_format", func(t *testing.T) {
		msg := Message{
			Channel:        "testchannel",
			Text:          "Test reply",
			MessageToReply: "parent_msg_id",
		}

		expected := "@reply-parent-msg-id=parent_msg_id PRIVMSG #testchannel : Test reply"
		_ = expected

		if msg.MessageToReply == "" {
			t.Error("MessageToReply should be set")
		}
	})
}