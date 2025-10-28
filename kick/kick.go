package kick

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// variaveis globais do streamer logado
var Token string
var UserID int
var UserLogin string
var LoginDone = make(chan bool)
var RefreshToken string
var TokenMutex sync.RWMutex
var CodeVerifier string
var OAuthState string

type ChatHandler struct {
	Conn     *websocket.Conn
	Channels []string
	MsgQueue Message
}

type Message struct {
	Channel IrcChannel
	Text    string
}

type KickMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Data    string `json:"data"`
}

type IrcChannel struct {
	ID        string
	Slug      string
	Connected bool
}

var Conn *websocket.Conn
var Channels []IrcChannel
var MsgQueue = make(chan Message, 100)

var ircHandlers = map[string]func(km KickMessage, data map[string]any){
	"pusher:connection_established": func(km KickMessage, data map[string]any) {
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "kick-connection",
			Data: UserLogin,
		}
	},
	"pusher_internal:subscription_succeeded": func(km KickMessage, data map[string]any) {
		channel := strings.Trim(strings.Split(km.Channel, ".")[1], " ")
		helpers.Logf(helpers.Kick, "[Kick IRC Handler] Subscribed to channel: %s", channel)
		for _, value := range Channels {
			if value.ID == channel {
				value.Connected = true
				globals.WsBroadcast <- globals.SocketMessage{
					Type: "kick-chat-connection",
					Data: map[string]any{"name": value.Slug, "id": value.ID}, //fmt.Sprintf("{\"name\":\"%s\",\"id\":\"%s\"}", value.Slug, value.ID),
				}
				break
			}
		}
	},
	"App\\Events\\ChatMessageEvent": func(km KickMessage, data map[string]any) {
		sender := data["sender"].(map[string]any)
		socketdata := globals.MessageFromStream{
			Source:    "kick",
			Channel:   km.Channel,
			UserId:    strconv.FormatFloat(sender["id"].(float64), 'f', 0, 64),
			User:      sender["username"].(string),
			MessageId: data["id"].(string),
			Message:   data["content"].(string),
			Metadata:  sender["identity"].(map[string]any),
		}
		globals.WsBroadcast <- globals.SocketMessage{Type: "user-message", Data: socketdata}
	},
	"App\\Events\\MessageDeletedEvent": func(km KickMessage, data map[string]any) {
		globals.WsBroadcast <- globals.SocketMessage{Type: "user-message-delete", Data: data["message"].(map[string]any)["id"]}
	},
}

func FindChannelByID(id string) *IrcChannel {
	for i, c := range Channels {
		if c.ID == id {
			return &Channels[i]
		}
	}
	return nil
}

func Connect() error {
	token := GetKickToken()
	if token == "" {
		return fmt.Errorf("kick Token não encontrado")
	}

	url := "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679?protocol=7&client=js&version=8.4.0-rc2&flash=false" // endpoint do chat Kick
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	//defer conn.Close()
	Conn = conn
	log.Printf("[Kick IRC] Connectado ao IRC")

	go reader()
	go writer()
	go func() {
		ticker := time.NewTicker(4 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			Conn.WriteJSON(map[string]string{"event": "pusher:ping"})
		}
	}()
	return nil
}

func JoinChannel(channel string) {
	helpers.Logf(helpers.Kick, "[Kick IRC] Conectando ao canal: %s", channel)
	token := GetKickToken()
	if token == "" {
		helpers.Logf(helpers.Red, "[Kick IRC] Token não encontrado")
		return
	}
	// Inscrição no chat autenticado
	subscribe := map[string]interface{}{
		"event": "pusher:subscribe",
		"data": map[string]interface{}{
			"channel": fmt.Sprintf("chatrooms.%s.v2", channel),
			"auth":    token,
		},
	}
	if err := Conn.WriteJSON(subscribe); err != nil {
		log.Println("[Kick IRC] Erro ao enviar subscribe:", err)
		return
	}
}

func reader() {
	for {
		_, msg, err := Conn.ReadMessage()
		if err != nil {
			helpers.Logf(helpers.Red, "[Kick IRC] ReadMessage erro: %s", err.Error())
			return
		}
		helpers.Logf(helpers.Kick, "[Kick IRC] Message: %s", msg)

		// Parse da mensagem JSON
		var km KickMessage
		if err := json.Unmarshal(msg, &km); err != nil {
			helpers.Logf(helpers.Red, "[Kick IRC] Erro ao parsear JSON: %v", err)
			continue
		}

		if handler, exists := ircHandlers[km.Event]; exists {
			data := map[string]any{}
			json.Unmarshal([]byte(km.Data), &data)
			handler(km, data)
			continue
		}
	}
}

func writer() {
	for msg := range MsgQueue {
		PostMessage(msg)
	}
}

func SendMessageIfChannelExist(msg string, channel string) {
	c := FindChannelByID(channel)
	if c == nil {
		helpers.Logf(helpers.Red, "[Kick] Canal não encontrado: %s", channel)
		return
	}
	SendMessage(msg, *c)
}

func SendMessage(msg string, channel IrcChannel) {
	MsgQueue <- Message{Channel: channel, Text: msg}
}
