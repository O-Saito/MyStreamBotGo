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

	goweb.SocketHandlers["get-next-streams-youtube"] = func(c *websocket.Conn, data map[string]any, tag int) {
		helpers.Printf(helpers.Reset, "[Socket Handler] get-next-streams-youtube %s\r\n", data["channel"].(string))
		lives, err := youtube.GetNextStreamings()
		if err != nil {
			helpers.Logf(helpers.ERROR, "Falha ao buscar YouTubeStreamings %v", err)
			return
		}

		previews := globals.GetState().GetData("youtube-preview-lives")
		if previews == nil {
			previews = []youtube.LiveBroadcast{}
		}

		for _, v := range lives.Items {

			_, finded := helpers.Find(previews.([]youtube.LiveBroadcast), func(l youtube.LiveBroadcast) bool {
				return l.Snippet.LiveChatID == v.Snippet.LiveChatID
			})
			if !finded {
				previews = append(previews.([]youtube.LiveBroadcast), v)
			}
		}

		globals.GetState().SetData("youtube-preview-lives", previews)

		globals.WsBroadcast <- globals.SocketMessage{
			Respond: tag,
			Type:    "result-get-next-streams-youtube",
			Data:    previews,
		}
	}

	goweb.SocketHandlers["connect-to-preview-youtube"] = func(c *websocket.Conn, data map[string]any, tag int) {
		helpers.Printf(helpers.Reset, "[Socket Handler] connect-to-preview-youtube %s\r\n", data["liveChatId"].(string))
		liveChatId := data["liveChatId"].(string)

		if liveChatId == "" {
			return
		}

		previews := globals.GetState().GetData("youtube-preview-lives")
		if previews == nil {
			return
		}

		connectedChat := globals.GetState().GetData("youtube-lives")
		if connectedChat == nil {
			connectedChat = []youtube.LiveBroadcast{}
		}

		_, findedConnected := helpers.Find(connectedChat.([]youtube.LiveBroadcast), func(l youtube.LiveBroadcast) bool {
			return l.Snippet.LiveChatID == liveChatId
		})

		if findedConnected {
			return
		}

		current, finded := helpers.Find(previews.([]youtube.LiveBroadcast), func(l youtube.LiveBroadcast) bool {
			return l.Snippet.LiveChatID == liveChatId
		})

		if finded {
			connectedChat = append(connectedChat.([]youtube.LiveBroadcast), current)
			go youtube.ListenToChat(liveChatId)
			globals.GetState().SetData("youtube-lives", connectedChat)
		}

		globals.WsBroadcast <- globals.SocketMessage{
			Respond: tag,
			Type:    "result-connect-chat-youtube",
			Data:    connectedChat,
		}
	}

	goweb.SocketHandlers["send-chat-message"] = func(c *websocket.Conn, data map[string]any, tag int) {
		if len(twitch.Channels) > 0 {
			for _, channel := range twitch.Channels {
				twitch.SendMessage(data["text"].(string), channel)
			}
		}
		if len(kick.Channels) > 0 {
			for _, channel := range kick.Channels {
				kick.SendMessage(data["text"].(string), channel)
			}
		}
		if ytLives := globals.GetState().GetData("youtube-lives"); ytLives != nil && len(ytLives.([]youtube.LiveBroadcast)) > 0 {
			for _, live := range ytLives.([]youtube.LiveBroadcast) {
				helpers.Logf(helpers.DEBUG, "Youtube chat should send to %s: %s", live.Snippet.LiveChatID, data["text"].(string))
				//youtube.SendMessage(data["text"].(string), live.Snippet.LiveChatID)
				//helpers.Printf(helpers.Yellow, "YT Send Message to %s: %s", live.Snippet.Title, data["text"].(string))
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
		twitchData, _ := twitch.GetStreamData(globals.GetState().GetTwitchUser().UserID)

		ytData := []any{}
		if ytLives := globals.GetState().GetData("youtube-lives"); ytLives != nil && len(ytLives.([]youtube.LiveBroadcast)) > 0 {
			for _, live := range ytLives.([]youtube.LiveBroadcast) {
				data, err := youtube.GetStreamData(live.ID)
				if err != nil {
					helpers.Logf(helpers.ERROR, "Failed to get data from YouTubeStreaming %v", err)
					continue
				}
				ytData = append(ytData, data)
			}
		}

		globals.WsBroadcast <- globals.SocketMessage{
			Respond: tag,
			Type:    "result-get-streamer-data",
			Data: map[string]any{
				"twitch":  twitchData,
				"youtube": ytData,
			},
		}
	}

}
