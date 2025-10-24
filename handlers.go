package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/helpers"
	"MyStreamBot/kick"
	"MyStreamBot/twitch"

	"github.com/gorilla/websocket"
)

func RegisterSocketHandlers() {
	goweb.SocketHandlers["connect-chat-kick"] = func(c *websocket.Conn, data map[string]any) {
		helpers.Logf(helpers.Reset, "[Socket Handler] connect-chat-kick %s\r\n", data["roomId"].(string))
		kick.Channels = append(kick.Channels, kick.IrcChannel{
			ID:   data["roomId"].(string),
			Slug: data["channel"].(string),
			//Connected: false,
		})
		kick.JoinChannel(data["roomId"].(string))
	}

	goweb.SocketHandlers["connect-chat-twitch"] = func(c *websocket.Conn, data map[string]any) {
		helpers.Logf(helpers.Reset, "[Socket Handler] connect-chat-twitch %s\r\n", data["channel"].(string))
		twitch.JoinChannel(data["channel"].(string))
	}

	goweb.SocketHandlers["send-chat-message"] = func(c *websocket.Conn, data map[string]any) {
		if len(twitch.Channels) > 0 {
			for _, c := range twitch.Channels {
				twitch.SendMessage(data["text"].(string), c)
			}
		}
		if len(kick.Channels) > 0 {
			for _, c := range kick.Channels {
				kick.SendMessage(data["text"].(string), c)
			}
		}
	}

	goweb.SocketHandlers["query-stream-game"] = func(c *websocket.Conn, m map[string]any) {
		games, _ := twitch.GetListOfGames(m["q"].(string))
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "result-query-stream-games",
			Data: map[string]any{
				"list": games,
			},
		}
	}

}
