package processors

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"MyStreamBot/plugin"
)

func ProcessEventQueue() {
	helpers.Log(helpers.INFO, "Started event queue processor!")
	for ev := range globals.EventQueue {
		isTwitch := false
		shouldSkip := false
		if ev.Type == "twitch-eventsub-notification" {
			isTwitch = true
		}

		if isTwitch {
			if payload, ok := ev.Data["payload"].(map[string]any); ok {
				if eventData, ok := payload["event"].(map[string]any); ok {
					if broadcasterId, ok := eventData["broadcaster_user_id"].(string); ok {
						if broadcasterId != globals.GetState().TwitchUser.UserID {
							shouldSkip = true
						}
					}
				}
			}
		}

		if !shouldSkip {
			globals.WsBroadcast <- globals.SocketMessage{Type: ev.Type, Data: ev.Data}
		}

		if ev.Type == "twitch-eventsub-notification" {
			ev.Type = ev.Data["payload"].(map[string]any)["subscription"].(map[string]any)["type"].(string)
		}

		mlua.DyEventQueue <- mlua.DyEventQueueData{
			Type:     mlua.DyEventEvent,
			LuaEvent: ev,
		}
		plugin.DispatchEvent(&ev)

		if !shouldSkip {
			mlua.HandleEvent(ev.Type, &ev)
		}
	}
}
