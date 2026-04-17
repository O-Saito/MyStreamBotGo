package processors

import (
	"testing"
	"time"

	"MyStreamBot/globals"
	"MyStreamBot/mlua"
)

func TestProcessChatQueue_Integration(t *testing.T) {
	oldChatQueue := globals.ChatQueue
	oldWsBroadcast := globals.WsBroadcast
	oldCommandQueue := globals.CommandQueue

	globals.ChatQueue = make(chan globals.MessageFromStream, 10)
	globals.WsBroadcast = make(chan globals.SocketMessage, 10)
	globals.CommandQueue = make(chan globals.Command, 10)

	cfg := globals.GetConfig()
	cfg.BotPrefix = "!"

	go ProcessChatQueue()
	time.Sleep(50 * time.Millisecond)

	tests := []struct {
		name          string
		msg           globals.MessageFromStream
		wantWsType    string
		wantCmdName   string
	}{
		{
			name: "regular_message_broadcasts",
			msg: globals.MessageFromStream{
				Source:    "twitch",
				Channel:  "testchannel",
				User:     "testuser",
				UserId:   "123",
				MessageId: "msg_1",
				Message:  "Hello world",
			},
			wantWsType: "user-message",
		},
		{
			name: "self_message_broadcasts_as_self",
			msg: globals.MessageFromStream{
				Source:    "twitch",
				Channel:  "testchannel",
				User:     "testuser",
				UserId:   "123",
				MessageId: "self",
				Message:  "Hello world",
			},
			wantWsType: "self-message",
		},
		{
			name: "command_extracts_to_queue",
			msg: globals.MessageFromStream{
				Source:    "twitch",
				Channel:  "testchannel",
				User:     "testuser",
				UserId:   "123",
				MessageId: "msg_2",
				Message:  "!testcmd arg1 arg2",
			},
			wantWsType: "user-message",
			wantCmdName: "testcmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			select {
			case globals.ChatQueue <- tt.msg:
			case <-time.After(200 * time.Millisecond):
				t.Fatal("timeout sending to ChatQueue")
			}

			select {
			case wsMsg := <-globals.WsBroadcast:
				if wsMsg.Type != tt.wantWsType {
					t.Errorf("WsBroadcast type = %q, want %q", wsMsg.Type, tt.wantWsType)
				}
			case <-time.After(200 * time.Millisecond):
				t.Error("timeout waiting for WsBroadcast")
			}

			if tt.wantCmdName != "" {
				select {
				case cmd := <-globals.CommandQueue:
					if cmd.Name != tt.wantCmdName {
						t.Errorf("Command.Name = %q, want %q", cmd.Name, tt.wantCmdName)
					}
				case <-time.After(200 * time.Millisecond):
					t.Error("timeout waiting for CommandQueue")
				}
			}
		})
	}

	globals.ChatQueue = oldChatQueue
	globals.WsBroadcast = oldWsBroadcast
	globals.CommandQueue = oldCommandQueue
}

func TestProcessCommandQueue_Integration(t *testing.T) {
	oldCommandQueue := globals.CommandQueue
	oldDyEventQueue := mlua.DyEventQueue

	globals.CommandQueue = make(chan globals.Command, 10)
	mlua.DyEventQueue = make(chan mlua.DyEventQueueData, 10)

	go ProcessCommandQueue()
	time.Sleep(50 * time.Millisecond)

	cmd := globals.Command{
		Source:  "twitch",
		Channel: "testchannel",
		Name:    "testcmd",
		Args:    []string{"arg1", "arg2"},
		User:    "testuser",
		Text:    "!testcmd arg1 arg2",
	}

	select {
	case globals.CommandQueue <- cmd:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout sending to CommandQueue")
	}

	select {
	case dyEv := <-mlua.DyEventQueue:
		if dyEv.Type != mlua.DyEventCommand {
			t.Errorf("DyEventQueue type = %d, want %d", dyEv.Type, mlua.DyEventCommand)
		}
		if dyEv.LuaCommand.Name != "testcmd" {
			t.Errorf("LuaCommand.Name = %q, want %q", dyEv.LuaCommand.Name, "testcmd")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("timeout waiting for DyEventQueue")
	}

	globals.CommandQueue = oldCommandQueue
	mlua.DyEventQueue = oldDyEventQueue
}

func TestProcessEventQueue_Integration(t *testing.T) {
	oldEventQueue := globals.EventQueue
	oldWsBroadcast := globals.WsBroadcast

	globals.EventQueue = make(chan globals.Event, 10)
	globals.WsBroadcast = make(chan globals.SocketMessage, 10)

	go ProcessEventQueue()
	time.Sleep(50 * time.Millisecond)

	tests := []struct {
		name       string
		ev        globals.Event
		wantWsType string
	}{
		{
			name: "stream_online",
			ev: globals.Event{
				Type: "stream.online",
				Data: map[string]interface{}{
					"channel": "test",
				},
			},
			wantWsType: "stream.online",
		},
		{
			name: "channel_follow",
			ev: globals.Event{
				Type: "channel.follow",
				Data: map[string]interface{}{
					"user": "follower",
				},
			},
			wantWsType: "channel.follow",
		},
		{
			name: "clear_chat",
			ev: globals.Event{
				Type: "clear-chat",
				Data: map[string]interface{}{
					"channel": "test",
				},
			},
			wantWsType: "clear-chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			select {
			case globals.EventQueue <- tt.ev:
			case <-time.After(200 * time.Millisecond):
				t.Fatal("timeout sending to EventQueue")
			}

			select {
			case wsMsg := <-globals.WsBroadcast:
				if wsMsg.Type != tt.wantWsType {
					t.Errorf("WsBroadcast type = %q, want %q", wsMsg.Type, tt.wantWsType)
				}
			case <-time.After(200 * time.Millisecond):
				t.Error("timeout waiting for WsBroadcast")
			}
		})
	}

	globals.EventQueue = oldEventQueue
	globals.WsBroadcast = oldWsBroadcast
}

func TestProcessDyEventQueue_Integration(t *testing.T) {
	oldDyEventQueue := mlua.DyEventQueue
	oldWsBroadcast := globals.WsBroadcast

	mlua.DyEventQueue = make(chan mlua.DyEventQueueData, 10)
	globals.WsBroadcast = make(chan globals.SocketMessage, 10)

	go ProcessDyEventQueue()
	time.Sleep(50 * time.Millisecond)

	ev := mlua.DyEventQueueData{
		Type: mlua.DyEventChat,
		MessageFromStream: globals.MessageFromStream{
			Source:    "twitch",
			Channel:  "testchannel",
			User:     "testuser",
			UserId:   "123",
			MessageId: "msg_1",
			Message:  "!test arg",
		},
	}

	select {
	case mlua.DyEventQueue <- ev:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout sending to DyEventQueue")
	}

	select {
	case <-globals.WsBroadcast:
	case <-time.After(200 * time.Millisecond):
	}

	mlua.DyEventQueue = oldDyEventQueue
	globals.WsBroadcast = oldWsBroadcast
}