package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"MyStreamBot/services/kick"
	"MyStreamBot/services/twitch"
	"MyStreamBot/services/youtube"

	"github.com/gorilla/websocket"
)

func RegisterSocketHandlers() {
	goweb.SocketHandlers["connect-chat-kick"] = func(c *websocket.Conn, data map[string]any, md *goweb.SocketRequestMetadata) {
		roomId, ok := data["roomId"].(string)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] connect-chat-kick: invalid roomId type")
			return
		}
		slug, ok := data["channel"].(string)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] connect-chat-kick: invalid channel type")
			return
		}
		helpers.Printf(helpers.Reset, "[Socket Handler] connect-chat-kick %s\r\n", roomId)
		kick.Channels = append(kick.Channels, kick.IrcChannel{
			ID:   roomId,
			Slug: slug,
		})
		kick.JoinChannel(roomId)
	}

	goweb.SocketHandlers["connect-chat-twitch"] = func(c *websocket.Conn, data map[string]any, md *goweb.SocketRequestMetadata) {
		channel, ok := data["channel"].(string)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] connect-chat-twitch: invalid channel type")
			return
		}
		helpers.Printf(helpers.Reset, "[Socket Handler] connect-chat-twitch %s\r\n", channel)
		twitch.JoinChannel(channel)
	}

	goweb.SocketHandlers["connect-chat-youtube"] = func(c *websocket.Conn, data map[string]any, md *goweb.SocketRequestMetadata) {
		channel, ok := data["channel"].(string)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] connect-chat-youtube: invalid channel type")
			return
		}
		helpers.Printf(helpers.Reset, "[Socket Handler] connect-chat-youtube %s\r\n", channel)
		lives, err := youtube.GetCurrentStreamings()
		if err != nil {
			helpers.Logf(helpers.ERROR, "Failed to fetch YouTubeStreamings: %v", err)
			return
		}

		connectedChat := globals.GetState().GetData("youtube-lives")
		if connectedChat == nil {
			connectedChat = []youtube.LiveBroadcast{}
		}
		connectedChatTyped, ok := connectedChat.([]youtube.LiveBroadcast)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] connect-chat-youtube: invalid youtube-lives data type")
			return
		}

		helpers.Printf(helpers.Cyan, "YT LIVES: %v", len(lives.Items))
		for _, v := range lives.Items {

			_, finded := helpers.Find(connectedChatTyped, func(l youtube.LiveBroadcast) bool {
				return l.Snippet.LiveChatID == v.Snippet.LiveChatID
			})
			helpers.Printf(helpers.Cyan, "YT Listen to (%v): %v", finded, v)
			if !finded {
				go youtube.ListenToChat(v.Snippet.LiveChatID)
				connectedChatTyped = append(connectedChatTyped, v)
			}
		}

		globals.GetState().SetData("youtube-lives", connectedChatTyped)

		globals.WsBroadcast <- globals.SocketMessage{
			SocketTag:         md.Tag,
			ResponseMessageID: md.ID,
			Type:              "result-connect-chat-youtube",
			Data:              connectedChat,
		}
	}

	goweb.SocketHandlers["get-next-streams-youtube"] = func(c *websocket.Conn, data map[string]any, md *goweb.SocketRequestMetadata) {
		channel, ok := data["channel"].(string)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] get-next-streams-youtube: invalid channel type")
			return
		}
		helpers.Printf(helpers.Reset, "[Socket Handler] get-next-streams-youtube %s\r\n", channel)
		lives, err := youtube.GetNextStreamings()
		if err != nil {
			helpers.Logf(helpers.ERROR, "Failed to fetch YouTubeStreamings: %v", err)
			return
		}

		previews := globals.GetState().GetData("youtube-preview-lives")
		if previews == nil {
			previews = []youtube.LiveBroadcast{}
		}
		previewsTyped, ok := previews.([]youtube.LiveBroadcast)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] get-next-streams-youtube: invalid youtube-preview-lives data type")
			return
		}

		for _, v := range lives.Items {

			_, finded := helpers.Find(previewsTyped, func(l youtube.LiveBroadcast) bool {
				return l.Snippet.LiveChatID == v.Snippet.LiveChatID
			})
			if !finded {
				previewsTyped = append(previewsTyped, v)
			}
		}

		globals.GetState().SetData("youtube-preview-lives", previewsTyped)

		globals.WsBroadcast <- globals.SocketMessage{
			SocketTag:         md.Tag,
			ResponseMessageID: md.ID,
			Type:              "result-get-next-streams-youtube",
			Data:              previews,
		}
	}

	goweb.SocketHandlers["connect-to-preview-youtube"] = func(c *websocket.Conn, data map[string]any, md *goweb.SocketRequestMetadata) {
		liveChatId, ok := data["liveChatId"].(string)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] connect-to-preview-youtube: invalid liveChatId type")
			return
		}
		helpers.Printf(helpers.Reset, "[Socket Handler] connect-to-preview-youtube %s\r\n", liveChatId)

		if liveChatId == "" {
			return
		}

		previews := globals.GetState().GetData("youtube-preview-lives")
		if previews == nil {
			return
		}
		previewsTyped, ok := previews.([]youtube.LiveBroadcast)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] connect-to-preview-youtube: invalid youtube-preview-lives data type")
			return
		}

		connectedChat := globals.GetState().GetData("youtube-lives")
		if connectedChat == nil {
			connectedChat = []youtube.LiveBroadcast{}
		}
		connectedChatTyped, ok := connectedChat.([]youtube.LiveBroadcast)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] connect-to-preview-youtube: invalid youtube-lives data type")
			return
		}

		_, findedConnected := helpers.Find(connectedChatTyped, func(l youtube.LiveBroadcast) bool {
			return l.Snippet.LiveChatID == liveChatId
		})

		if findedConnected {
			return
		}

		current, finded := helpers.Find(previewsTyped, func(l youtube.LiveBroadcast) bool {
			return l.Snippet.LiveChatID == liveChatId
		})

		if finded {
			connectedChatTyped = append(connectedChatTyped, current)
			go youtube.ListenToChat(liveChatId)
			globals.GetState().SetData("youtube-lives", connectedChatTyped)
		}

		globals.WsBroadcast <- globals.SocketMessage{
			SocketTag:         md.Tag,
			ResponseMessageID: md.ID,
			Type:              "result-connect-chat-youtube",
			Data:              connectedChatTyped,
		}
	}

	goweb.SocketHandlers["send-chat-message"] = func(c *websocket.Conn, data map[string]any, md *goweb.SocketRequestMetadata) {
		text, ok := data["text"].(string)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] send-chat-message: invalid text type")
			return
		}
		if len(twitch.Channels) > 0 {
			for _, channel := range twitch.Channels {
				twitch.SendMessage(text, channel)
			}
		}
		if len(kick.Channels) > 0 {
			for _, channel := range kick.Channels {
				kick.SendMessage(text, channel)
			}
		}
		if ytLives := globals.GetState().GetData("youtube-lives"); ytLives != nil {
			ytLivesTyped, ok := ytLives.([]youtube.LiveBroadcast)
			if ok && len(ytLivesTyped) > 0 {
				for _, live := range ytLivesTyped {
					helpers.Logf(helpers.DEBUG, "Youtube chat should send to %s: %s", live.Snippet.LiveChatID, text)
				}
			}
		}
	}

	goweb.SocketHandlers["query-stream-game"] = func(c *websocket.Conn, m map[string]any, md *goweb.SocketRequestMetadata) {
		q, ok := m["q"].(string)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Socket Handler] query-stream-game: invalid q type")
			return
		}
		games, _ := twitch.GetListOfGames(q)
		globals.WsBroadcast <- globals.SocketMessage{
			SocketTag:         md.Tag,
			ResponseMessageID: md.ID,
			Type:              "result-query-stream-games",
			Data: map[string]any{
				"list": games,
			},
		}
	}

	goweb.SocketHandlers["get-streamer-data"] = func(c *websocket.Conn, m map[string]any, md *goweb.SocketRequestMetadata) {
		twitchData, _ := twitch.GetStreamData(globals.GetState().GetTwitchUser().UserID)

		ytData := []any{}
		if ytLives := globals.GetState().GetData("youtube-lives"); ytLives != nil {
			ytLivesTyped, ok := ytLives.([]youtube.LiveBroadcast)
			if ok && len(ytLivesTyped) > 0 {
				for _, live := range ytLivesTyped {
					data, err := youtube.GetStreamData(live.ID)
					if err != nil {
						helpers.Logf(helpers.ERROR, "Failed to get data from YouTubeStreaming %v", err)
						continue
					}
					ytData = append(ytData, data)
				}
			}
		}

		globals.WsBroadcast <- globals.SocketMessage{
			SocketTag:         md.Tag,
			ResponseMessageID: md.ID,
			Type:              "result-get-streamer-data",
			Data: map[string]any{
				"twitch":  twitchData,
				"youtube": ytData,
			},
		}
	}

	goweb.SocketHandlers["get-dy-statistics"] = func(c *websocket.Conn, m map[string]any, md *goweb.SocketRequestMetadata) {
		events := mlua.ListDynamicEvents()

		globals.WsBroadcast <- globals.SocketMessage{
			SocketTag:         md.Tag,
			ResponseMessageID: md.ID,
			Type:              "result-get-dy-statistics",
			Data:              events,
		}
	}

}
