package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

type Message struct {
	Channel        string
	Text           string
	MessageToReply string
}

// variaveis globais do streamer logado
var LoginDone = make(chan bool)

var Conn net.Conn
var Channels []string
var MsgQueue = make(chan Message, 100)

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

var ircHandlers = map[string]func(parts []string, afterMetadataIndex int, metadata ...map[string]any){
	"RECONNECT": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		// fazer o reconnect
		helpers.Logf(helpers.Twitch, "[TWITCH RECONNECT] Server requested reconnect")
		Disconnect()
		/*for _, channel := range Channels {
			JoinChannel(channel)
		}*/
	},
	"JOIN": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		user := strings.Split(parts[0], "!")[0][1:]
		channel := strings.TrimPrefix(parts[2], "#")
		helpers.Logf(helpers.Twitch, "[TWITCH JOIN] User %s joined channel %s", user, channel)
		if user == channel {
			globals.WsBroadcast <- globals.SocketMessage{
				Type: "twitch-chat-connection",
				Data: fmt.Sprintf("{\"name\":\"%s\",\"id\":\"#%s\"}", channel, channel),
			}
		}
	},
	"CLEARMSG": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		channel := strings.TrimPrefix(parts[afterMetadataIndex+2], "#")
		reason := ""
		if len(parts) > afterMetadataIndex+3 {
			reason = strings.Join(parts[(afterMetadataIndex+3):], " ")[1:]
		}
		helpers.Logf(helpers.Twitch, "[TWITCH CLEARMSG] Chat %s cleared: %s", channel, reason)
		globals.WsBroadcast <- globals.SocketMessage{Type: "user-message-delete", Data: fmt.Sprintf("\"%s\"", metadata[0]["target-msg-id"].(string))}
	},
	"CLEARCHAT": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		channel := strings.TrimPrefix(parts[afterMetadataIndex+2], "#")
		helpers.Logf(helpers.Twitch, "[TWITCH CLEARCHAT] Chat %s cleared", channel)
		globals.WsBroadcast <- globals.SocketMessage{Type: "clear-chat", Data: channel}
	},
	"NOTICE": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		channel := strings.TrimPrefix(parts[afterMetadataIndex+2], "#")
		reason := strings.Join(parts[(afterMetadataIndex+3):], " ")[1:]
		helpers.Logf(helpers.Twitch, "[TWITCH NOTICE] Notice in %s: %s", channel, reason)
	},
	"PRIVMSG": func(parts []string, afterMetadataIndex int, metadata ...map[string]any) {
		user := strings.Split(parts[afterMetadataIndex], "!")[0][1:]
		channel := strings.TrimPrefix(parts[afterMetadataIndex+2], "#")
		message := strings.Join(parts[(afterMetadataIndex+3):], " ")[1:]
		helpers.Logf(helpers.Twitch, "[TWITCH MESSAGE] %s in %s: %s", user, channel, message)
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

		state := globals.GetState()
		info := state.GetData("twitch-badges-info")
		//infoChannel := state.GetData(fmt.Sprintf("twitch-badges-info-%s", channel))
		if info == nil {
			info, _ = GetBadges()
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

		bi := make(map[string]any)
		for _, v := range strings.Split(socketdata.Metadata["badges"].(string), ",") {
			n := strings.Split(v, "/")[0]
			bi[n] = info.(map[string]any)[n]
		}

		socketdata.Metadata["badges-info"] = bi

		dataJSON, _ := json.Marshal(socketdata)
		globals.WsBroadcast <- globals.SocketMessage{Type: "user-message", Data: string(dataJSON)}
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
			helpers.Logf(helpers.Purple, "[TWITCH COMMAND] %+v", cmd)
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
		helpers.Logf(helpers.Twitch, "[TWITCH RECONNECT] Server requested reconnect")
		Disconnect()
		Connect()
	}

	for _, channel := range Channels {
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
	helpers.Logf(helpers.Blue, "[TWITCH] Joining channel: %s", channel)
	fmt.Fprintf(Conn, "JOIN #%s\r\n", channel)
	Channels = append(Channels, channel)
}

func reader() {
	scanner := bufio.NewScanner(Conn)
	for scanner.Scan() {
		msg := scanner.Text()
		helpers.Logf(helpers.Blue, "[Twitch] IRC Message: %s", msg)
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
		helpers.Logf(helpers.Twitch, "[Twitch] Handler key: %s", handlersKey)
		if handler, exists := ircHandlers[handlersKey]; exists {
			handler(parts, afterMetadataIndex, helpers.Ternary(parts[0][0] == '@', partseTags(parts[0]), nil))
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		helpers.Logf(helpers.Red, "[Twitch ERROR] Erro na leitura: %v", err)
		ircHandlers["RECONNECT"](nil, 0, nil)
	} else {
		helpers.Logf(helpers.Red, "[Twitch ERROR] Scanner finalizado")
		ircHandlers["RECONNECT"](nil, 0, nil)
	}
}

func writer() {
	for msg := range MsgQueue {
		if msg.MessageToReply != "" {
			text := fmt.Sprintf("@reply-parent-msg-id=%s PRIVMSG #%s : %s", msg.MessageToReply, msg.Channel, msg.Text)
			helpers.Logf(helpers.Yellow, "[TWITCH REPLY] %s", text)
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

	helpers.Logf(helpers.Red, "[TWITCH ERROR] Channel not found! %s", channel)
}
