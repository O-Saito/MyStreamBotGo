package processors

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"MyStreamBot/plugin"
	"MyStreamBot/services/twitch"
	"strings"
)

func ProcessChatQueue() {
	helpers.Log(helpers.INFO, "Started chat queue processor!")
	for {
		select {
		case <-stopCh:
			return
		case ev, ok := <-globals.ChatQueue:
			if !ok {
				return
			}
			processChatMessage(ev)
		}
	}
}

func processChatMessage(ev globals.MessageFromStream) {
	config := globals.GetConfig()

	if ev.MessageId == "self" {
		globals.SafeSend(globals.WsBroadcast, globals.SocketMessage{Type: "self-message", Data: ev}, "WsBroadcast", false)
		return
	}

	ev.IsCommand = strings.HasPrefix(ev.Message, config.BotPrefix)
	ev.IsAtOwnerChannel = true

	if ev.Source == "twitch" {
		ev.IsAtOwnerChannel = twitch.IsOwnChannelMessage(ev.Metadata)
	}

	globals.SafeSend(globals.WsBroadcast, globals.SocketMessage{Type: "user-message", Data: ev}, "WsBroadcast", false)

	if ev.IsCommand {
		parts := strings.SplitN(ev.Message[1:], " ", 2)
		cmd := globals.Command{
			Source:  ev.Source,
			Name:    parts[0],
			Channel: ev.Channel,
			Args:    parts[1:],
			User:    ev.User,
			Text:    ev.Message,
			Message: ev,
			Data:    map[string]any{},
		}
		if len(parts) > 1 {
			cmd.Args = strings.Split(parts[1], " ")
		}
		helpers.Printf(helpers.Yellow, "[CHAT QUEUE COMMAND] %+v", cmd)
		globals.CommandQueue <- cmd
	}

	mlua.DyEventQueue <- mlua.DyEventQueueData{
		Type:              mlua.DyEventChat,
		MessageFromStream: ev,
	}
	plugin.DispatchChat(&ev)
	if ev.IsAtOwnerChannel {
		mlua.HandleChat(&ev)
	}
}
