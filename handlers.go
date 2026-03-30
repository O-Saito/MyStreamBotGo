package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/helpers"
	"MyStreamBot/services/kick"
	"MyStreamBot/services/twitch"
	"MyStreamBot/services/youtube"

	"github.com/gorilla/websocket"
)

func RegisterSocketHandlers() {
	goweb.SocketHandlers["connect-chat-kick"] = func(c *websocket.Conn, data map[string]any, tag int) {
		helpers.Printf(helpers.Reset, "[Socket Handler] connect-chat-kick %s\r\n", data["roomId"].(string))
		kick.Channels = append(kick.Channels, kick.IrcChannel{
			ID:   data["roomId"].(string),
			Slug: data["channel"].(string),
			//Connected: false,
		})
		kick.JoinChannel(data["roomId"].(string))
	}

	goweb.SocketHandlers["connect-chat-twitch"] = func(c *websocket.Conn, data map[string]any, tag int) {
		helpers.Printf(helpers.Reset, "[Socket Handler] connect-chat-twitch %s\r\n", data["channel"].(string))
		twitch.JoinChannel(data["channel"].(string))
	}

	goweb.SocketHandlers["connect-chat-youtube"] = func(c *websocket.Conn, data map[string]any, tag int) {
		helpers.Printf(helpers.Reset, "[Socket Handler] connect-chat-youtube %s\r\n", data["channel"].(string))
		lives, err := youtube.GetCurrentStreamings()
		if err != nil {
			helpers.Logf(helpers.ERROR, "Falha ao buscar YouTubeStreamings %v", err)
			return
		}

		connectedChat := globals.GetState().GetData("youtube-lives")
		if connectedChat == nil {
			connectedChat = []youtube.LiveBroadcast{}
		}

		helpers.Printf(helpers.Cyan, "YT LIVES: %v", len(lives.Items))
		for _, v := range lives.Items {

			_, finded := helpers.Find(connectedChat.([]youtube.LiveBroadcast), func(l youtube.LiveBroadcast) bool {
				return l.Snippet.LiveChatID == v.Snippet.LiveChatID
			})
			helpers.Printf(helpers.Cyan, "YT Listen to (%v): %v", finded, v)
			if !finded {
				go youtube.ListenToChat(v.Snippet.LiveChatID)
				connectedChat = append(connectedChat.([]youtube.LiveBroadcast), v)
			}
		}

		globals.GetState().SetData("youtube-lives", connectedChat)

		globals.WsBroadcast <- globals.SocketMessage{
			Respond: tag,
			Type:    "result-connect-chat-youtube",
			Data:    connectedChat,
		}
	}

	goweb.SocketHandlers["send-chat-message"] = func(c *websocket.Conn, data map[string]any, tag int) {
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

	goweb.SocketHandlers["query-stream-game"] = func(c *websocket.Conn, m map[string]any, tag int) {
		games, _ := twitch.GetListOfGames(m["q"].(string))
		globals.WsBroadcast <- globals.SocketMessage{
			Respond: tag,
			Type:    "result-query-stream-games",
			Data: map[string]any{
				"list": games,
			},
		}
	}

	goweb.SocketHandlers["get-streamer-data"] = func(c *websocket.Conn, m map[string]any, tag int) {
		data, _ := twitch.GetStreamData(globals.GetState().GetTwitchUser().UserID)
		globals.WsBroadcast <- globals.SocketMessage{
			Respond: tag,
			Type:    "result-get-streamer-data",
			Data:    data,
		}
	}

}
