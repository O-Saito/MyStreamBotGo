package processors

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"strings"
)

func ProcessChatQueue() {
	helpers.Log(helpers.INFO, "Started chat queue processor!")
	for ev := range globals.ChatQueue {
		config := globals.GetConfig()

		// Ignore messages sent by the bot itself to prevent loops
		if ev.MessageId == "self" {
			globals.WsBroadcast <- globals.SocketMessage{Type: "self-message", Data: ev}
			continue
		}

		globals.WsBroadcast <- globals.SocketMessage{Type: "user-message", Data: ev}

		if strings.HasPrefix(ev.Message, config.BotPrefix) {
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
		if ev.Source == "twitch" {
			if ev.Channel != globals.GetState().TwitchUser.UserLogin {
				continue
			}
		}
		mlua.HandleChat(&ev)
	}
}
