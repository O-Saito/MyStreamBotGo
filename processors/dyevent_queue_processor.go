package processors

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"sync"
	"time"
)

func ProcessDyEventQueue() {
	helpers.Log(helpers.INFO, "Started dyevents processor!")
	for deq := range mlua.DyEventQueue {
		events := mlua.GetDyEvents()
		wait := sync.WaitGroup{}
		wait.Add(len(events))
		for _, dev := range events {
			go func() {
				start := time.Now()
				sharedChat := dev.GetComputeTwitchSharedChat()
				helpers.Logf(helpers.DEBUG, "Processing dy event %s[%d] with shared chat: %t", dev.Name, deq.Type, sharedChat)
				switch deq.Type {
				case mlua.DyEventChat:
					if !sharedChat &&
						deq.MessageFromStream.Source == "twitch" &&
						deq.MessageFromStream.Channel != globals.GetState().TwitchUser.UserLogin {
						break
					}
					dev.ProcessChat(&deq.MessageFromStream)
				case mlua.DyEventCommand:
					if !sharedChat &&
						deq.LuaCommand.Source == "twitch" &&
						deq.LuaCommand.Channel != globals.GetState().TwitchUser.UserLogin {
						break
					}
					dev.ProcessCommand(&deq.LuaCommand)
				case mlua.DyEventEvent:
					if !sharedChat {
						if payload, ok := deq.LuaEvent.Data["payload"].(map[string]any); ok {
							if eventData, ok := payload["event"].(map[string]any); ok {
								if broadcasterLogin, ok := eventData["broadcaster_user_login"].(string); ok {
									if broadcasterLogin != globals.GetState().TwitchUser.UserLogin {
										break
									}
								}
							}
						} else if channel, ok := deq.LuaEvent.Data["channel"].(string); ok {
							if channel != globals.GetState().TwitchUser.UserLogin {
								break
							}
						}
					}
					dev.ProcessEvent(&deq.LuaEvent)
				case mlua.DyEventRequest:
					dev.ProcessRequest(&deq.SocketMessage)
				default:
				}
				elapsed := time.Since(start)
				dev.Statistics.AddTiming(elapsed)
				wait.Done()
			}()
		}
		wait.Wait()
	}
}
