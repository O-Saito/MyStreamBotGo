package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"bufio"
	"fmt"
	"hash/fnv"
	"net"
	"slices"
	"strings"
)

type Message struct {
	Channel        string
	Text           string
	MessageToReply string
}

// variaveis globais do streamer logado
var (
	LoginDone = make(chan bool)
	Conn      net.Conn
	Channels  []string
	MsgQueue  = make(chan Message, 100)
)

func partseTags(tagsStr string) map[string]any {
	metadata := map[string]any{}
	tags := strings.SplitSeq(strings.TrimLeft(tagsStr, "@"), ";")
	for tag := range tags {
		kv := strings.SplitN(tag, "=", 2)
		if len(kv) == 2 {
			metadata[kv[0]] = kv[1]
		}
	}
	return metadata
}

func defaultTwitchColor(username string) string {
	colors := []string{
		"#FF0000", "#0000FF", "#008000", "#B22222", "#FF7F50",
		"#9ACD32", "#FF4500", "#2E8B57", "#DAA520", "#D2691E",
		"#5F9EA0", "#1E90FF", "#FF69B4", "#8A2BE2", "#00FF7F",
	}

	h := fnv.New32a()
	h.Write([]byte(strings.ToLower(username)))
	index := int(h.Sum32()) % len(colors)

	return colors[index]
}

func GetCacheUserChatColor(user string) string {
	state := globals.GetState()
	userColor := state.GetData("twitch-user-color")
	color := ""
	if userColor == nil {
		userColor = make(map[string]any)
	}

	if userColor.(map[string]any)[user] == nil {
		d, err := GetUserData(user)
		if err == nil {
			c, err := GetUserChatColor(d.ID)
			if err == nil {
				color = c.Color
			}
		}

		if color == "" {
			color = defaultTwitchColor(user)
		}

		userColor.(map[string]any)[user] = color

		state.SetData("twitch-user-color", userColor)
	}

	return userColor.(map[string]any)[user].(string)
}

var ircHandlers = map[string]func(parts []string, afterMetadataIndex int, metadata ...map[string]any){
	"RECONNECT": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		// fazer o reconnect
		helpers.Logf(helpers.INFO, "[TWITCH RECONNECT] Server requested reconnect")
		Disconnect()
	},
	"JOIN": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		user := strings.Split(parts[0], "!")[0][1:]
		channel := strings.TrimPrefix(parts[2], "#")
		helpers.Printf(helpers.Twitch, "[TWITCH JOIN] User %s joined channel %s", user, channel)
		if user == channel {
			globals.WsBroadcast <- globals.SocketMessage{
				Type: "twitch-chat-connection",
				Data: map[string]any{
					"name": channel,
					"id":   channel,
				},
			}
		}

		globals.EventQueue <- globals.LuaEvent{
			Type: "twitch-user-join",
			Data: map[string]any{
				"user":     user,
				"channel":  channel,
				"metadata": metadata,
				"color":    GetCacheUserChatColor(user),
			},
		}
	},
	"PART": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		user := strings.Split(parts[0], "!")[0][1:]
		channel := strings.TrimPrefix(parts[2], "#")
		helpers.Printf(helpers.Twitch, "[TWITCH PART] User %s joined channel %s", user, channel)
		globals.EventQueue <- globals.LuaEvent{
			Type: "twitch-user-part",
			Data: map[string]any{
				"user":     user,
				"channel":  channel,
				"metadata": metadata,
				"color":    GetCacheUserChatColor(user),
			},
		}
	},
	"CLEARMSG": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		channel := strings.TrimPrefix(parts[afterMetadataIndex+2], "#")
		reason := ""
		if len(parts) > afterMetadataIndex+3 {
			reason = strings.Join(parts[(afterMetadataIndex+3):], " ")[1:]
		}
		helpers.Printf(helpers.Twitch, "[TWITCH CLEARMSG] Chat %s cleared: %s", channel, reason)
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "user-message-delete",
			Data: metadata[0]["target-msg-id"].(string),
		}
		globals.EventQueue <- globals.LuaEvent{
			Type: "user-message-delete",
			Data: metadata[0],
		}
	},
	"CLEARCHAT": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		channel := strings.TrimPrefix(parts[afterMetadataIndex+2], "#")
		helpers.Printf(helpers.Twitch, "[TWITCH CLEARCHAT] Chat %s cleared", channel)
		globals.WsBroadcast <- globals.SocketMessage{Type: "clear-chat", Data: map[string]any{
			"channel":  channel,
			"metadata": metadata[0],
		}}
	},
	"NOTICE": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		channel := strings.TrimPrefix(parts[afterMetadataIndex+2], "#")
		reason := strings.Join(parts[(afterMetadataIndex+3):], " ")[1:]
		helpers.Printf(helpers.Twitch, "[TWITCH NOTICE] Notice in %s: %s", channel, reason)

		/*globals.CommandQueue <- globals.LuaCommand{
			Source:  "twitch",
			Name:    strings.TrimPrefix(parts[0], "#"),
			Channel: channel,
			Args:    parts[1:],
			User:    user,
			Text:    message,
			Message: socketdata,
			Data:    map[string]any{},
		}*/
	},
	"PRIVMSG": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		user := strings.Split(parts[afterMetadataIndex], "!")[0][1:]
		channel := strings.TrimPrefix(parts[afterMetadataIndex+2], "#")
		message := strings.Join(parts[(afterMetadataIndex+3):], " ")[1:]
		helpers.Printf(helpers.Twitch, "[TWITCH MESSAGE] %s in %s: %s", user, channel, message)
		// enviar para WebSocket
		socketdata := globals.MessageFromStream{
			Source:    "twitch",
			Channel:   channel,
			User:      user,
			UserId:    metadata[0]["user-id"].(string),
			MessageId: metadata[0]["id"].(string),
			Message:   message,
			Metadata:  metadata[0],
		}

		// adding tag to facilitate user-type validation
		if user == channel {
			socketdata.Metadata["user-type"] = "mod"
		}

		state := globals.GetState()
		info := state.GetData("twitch-badges-info")
		//infoChannel := state.GetData(fmt.Sprintf("twitch-badges-info-%s", channel))
		if info == nil {
			data, _ := GetBadges()
			if data != nil {
				info = *data
			}
			state.SetData("twitch-badges-info", info)
		}

		roomId := socketdata.Metadata["room-id"]
		if socketdata.Metadata["source-room-id"] != nil {
			roomId = socketdata.Metadata["source-room-id"]
		}

		if roomId != nil {
			current := globals.GetState().GetTwitchUser()
			socketdata.Metadata["room"] = current
			if current.UserID != roomId {
				streamerInfo := state.GetData("twitch-streamer-info")
				if streamerInfo == nil {
					streamerInfo = make(map[string]any)
				}

				id := roomId.(string)
				if streamerInfo.(map[string]any)[id] == nil {
					streamerInfo.(map[string]any)[id], _ = GetUserDataById(id)
				}

				state.SetData("twitch-streamer-info", streamerInfo)

				socketdata.Metadata["room"] = streamerInfo.(map[string]any)[id]
			}
		}

		if socketdata.Metadata["color"] == nil || socketdata.Metadata["color"] == "" {
			userColor := state.GetData("twitch-user-color")
			if userColor == nil {
				userColor = make(map[string]any)
			}
			if userColor.(map[string]any)[user] == nil {
				userColor.(map[string]any)[user] = defaultTwitchColor(user)
				state.SetData("twitch-user-color", userColor)
			}
			socketdata.Metadata["color"] = userColor.(map[string]any)[user]
		}

		bi := make(map[string]any)
		for _, v := range strings.Split(socketdata.Metadata["badges"].(string), ",") {
			n := strings.Split(v, "/")[0]
			bi[n] = info.(map[string]any)[n]
		}

		socketdata.Metadata["badges-info"] = bi

		globals.WsBroadcast <- globals.SocketMessage{Type: "user-message", Data: socketdata}
		globals.ChatQueue <- socketdata

		config := globals.GetConfig()

		if strings.HasPrefix(message, config.BotPrefix) {
			parts := strings.SplitN(message[1:], " ", 2)
			cmd := globals.LuaCommand{
				Source:  "twitch",
				Name:    strings.TrimPrefix(parts[0], "#"),
				Channel: channel,
				Args:    parts[1:],
				User:    user,
				Text:    message,
				Message: socketdata,
				Data:    map[string]any{},
			}
			if len(parts) > 1 {
				cmd.Args = strings.Split(parts[1], " ")
			}
			globals.CommandQueue <- cmd
			helpers.Printf(helpers.Purple, "[TWITCH COMMAND] %+v", cmd)
		}
	},
}

func Connect() error {
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		return err
	}
	Conn = conn
	user := globals.GetState().GetTwitchUser()
	//fmt.Printf("{TWITCH USERDATA} %v \r\n", user)
	fmt.Fprintf(Conn, "PASS oauth:%s\r\n", user.Token)
	fmt.Fprintf(Conn, "NICK %s\r\n", user.UserLogin)
	fmt.Fprintf(Conn, "CAP REQ :twitch.tv/membership\r\n")
	fmt.Fprintf(Conn, "CAP REQ :twitch.tv/tags\r\n")
	fmt.Fprintf(Conn, "CAP REQ :twitch.tv/commands\r\n")

	go reader()
	go writer()

	ircHandlers["RECONNECT"] = func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		// fazer o reconnect
		helpers.Logf(helpers.INFO, "[TWITCH RECONNECT] Server requested reconnect")
		Disconnect()
		Connect()
	}

	reconnectChannels := Channels

	Channels = Channels[:0]

	for _, channel := range reconnectChannels {
		JoinChannel(channel)
	}

	return nil
}

func Disconnect() {
	if Conn != nil {
		Conn.Close()
	}
	//close(MsgQueue)
}

func JoinChannel(channel string) {
	if slices.Contains(Channels, channel) {
		return
	}
	helpers.Printf(helpers.Twitch, "[TWITCH] Joining channel: %s", channel)
	fmt.Fprintf(Conn, "JOIN #%s\r\n", channel)
	Channels = append(Channels, channel)
}

func reader() {
	scanner := bufio.NewScanner(Conn)
	for scanner.Scan() {
		msg := scanner.Text()
		helpers.Printf(helpers.Twitch, "[Twitch] IRC Message: %s", msg)
		if strings.HasPrefix(msg, "PING") {
			fmt.Fprintf(Conn, "PONG :tmi.twitch.tv\r\n")
			continue
		}
		parts := strings.Split(msg, " ")
		if len(parts) < 2 {
			continue
		}

		afterMetadataIndex := helpers.Ternary(parts[0][0] == '@', 1, 0)

		handlersKey := parts[afterMetadataIndex+1]
		helpers.Printf(helpers.Twitch, "[Twitch] Handler key: %s", handlersKey)
		if handler, exists := ircHandlers[handlersKey]; exists {
			handler(parts, afterMetadataIndex, helpers.Ternary(parts[0][0] == '@', partseTags(parts[0]), nil))
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		helpers.Logf(helpers.ERROR, "[Twitch ERROR] Erro na leitura: %v", err)
		ircHandlers["RECONNECT"](nil, 0, nil)
	} else {
		helpers.Logf(helpers.INFO, "[Twitch ERROR] Scanner finalizado")
		ircHandlers["RECONNECT"](nil, 0, nil)
	}
}

func writer() {
	for msg := range MsgQueue {
		if msg.MessageToReply != "" {
			text := fmt.Sprintf("@reply-parent-msg-id=%s PRIVMSG #%s : %s", msg.MessageToReply, msg.Channel, msg.Text)
			helpers.Printf(helpers.Yellow, "[TWITCH REPLY] %s", text)
			fmt.Fprintf(Conn, "%s\r\n", text)
			continue
		}
		fmt.Fprintf(Conn, "PRIVMSG #%s :%s\r\n", msg.Channel, msg.Text)
	}
}

func SendMessage(msg, channel string, messageToReply ...string) {
	if helpers.Contains(Channels, channel) {
		if messageToReply == nil {
			messageToReply = []string{""}
		}
		MsgQueue <- Message{Channel: channel, Text: msg, MessageToReply: helpers.Ternary(len(messageToReply) > 0, messageToReply[0], "")}
		return
	}

	helpers.Logf(helpers.ERROR, "[TWITCH ERROR] Channel not found! %s", channel)
}
