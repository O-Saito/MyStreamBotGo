package processors

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
)

func ProcessEventQueue() {
	helpers.Log(helpers.INFO, "Started event queue processor!")
	for ev := range globals.EventQueue {
		globals.WsBroadcast <- globals.SocketMessage{Type: ev.Type, Data: ev.Data}

		isTwitch := false
		if ev.Type == "twitch-eventsub-notification" {
			ev.Type = ev.Data["payload"].(map[string]any)["subscription"].(map[string]any)["type"].(string)
			isTwitch = true
		}

		mlua.DyEventQueue <- mlua.DyEventQueueData{
			Type:     mlua.DyEventEvent,
			LuaEvent: ev,
		}

		if isTwitch {
			helpers.Logf(helpers.DEBUG, "Skipping Twitch event for broadcaster %v and %V", ev.Data["payload"], globals.GetState().TwitchUser)
			if payload, ok := ev.Data["payload"].(map[string]any); ok {
				if eventData, ok := payload["event"].(map[string]any); ok {
					if broadcasterLogin, ok := eventData["broadcaster_user_login"].(string); ok {
						if broadcasterLogin != globals.GetState().TwitchUser.UserLogin {
							continue
						}
					}
				}
			}
		}
		mlua.HandleEvent(ev.Type, &ev)
	}
}
