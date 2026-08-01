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

// Global variables for the logged-in streamer
var LoginDone = make(chan bool)
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
var channelsMu sync.RWMutex

// AddChannel registers a new channel to join/track.
func AddChannel(id, slug string) {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	Channels = append(Channels, IrcChannel{ID: id, Slug: slug})
}

// GetChannels returns a snapshot copy of the tracked channel list.
func GetChannels() []IrcChannel {
	channelsMu.RLock()
	defer channelsMu.RUnlock()
	out := make([]IrcChannel, len(Channels))
	copy(out, Channels)
	return out
}

// ResetChannels clears the tracked channel list.
func ResetChannels() {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	Channels = []IrcChannel{}
}

// MarkChannelConnected flags the channel with the given ID as connected and
// returns its slug. ok is false if no channel with that ID is tracked.
func MarkChannelConnected(id string) (slug string, ok bool) {
	channelsMu.Lock()
	defer channelsMu.Unlock()
	for i := range Channels {
		if Channels[i].ID == id {
			Channels[i].Connected = true
			return Channels[i].Slug, true
		}
	}
	return "", false
}

var ircHandlers = map[string]func(km KickMessage, data map[string]any){
	"pusher:connection_established": func(km KickMessage, data map[string]any) {
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "kick-connection",
			Data: globals.GetState().GetKickUser().UserLogin,
		}
	},
	"pusher_internal:subscription_succeeded": func(km KickMessage, data map[string]any) {
		channel := strings.Trim(strings.Split(km.Channel, ".")[1], " ")
		helpers.Logf(helpers.DEBUG, "[Kick IRC Handler] Subscribed to channel: %s", channel)
		if slug, ok := MarkChannelConnected(channel); ok {
			globals.WsBroadcast <- globals.SocketMessage{
				Type: "kick-chat-connection",
				Data: map[string]any{"name": slug, "id": channel},
			}
		}
	},
	"App\\Events\\ChatMessageEvent": func(km KickMessage, data map[string]any) {
		sender, ok := data["sender"].(map[string]any)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Kick] ChatMessageEvent: invalid sender type")
			return
		}
		senderId, _ := sender["id"].(float64)
		username, _ := sender["username"].(string)
		msgId, _ := data["id"].(string)
		content, _ := data["content"].(string)
		identity, _ := sender["identity"].(map[string]any)

		socketdata := globals.MessageFromStream{
			Source:    "kick",
			Channel:   km.Channel,
			UserId:    strconv.FormatFloat(senderId, 'f', 0, 64),
			User:      username,
			MessageId: msgId,
			Message:   content,
			Metadata:  identity,
		}
		globals.ChatQueue <- socketdata
	},
	"App\\Events\\MessageDeletedEvent": func(km KickMessage, data map[string]any) {
		msgData, ok := data["message"].(map[string]any)
		if !ok {
			return
		}
		globals.EventQueue <- globals.Event{Type: "user-message-delete", Data: map[string]any{"messageId": msgData["id"]}}
	},
}

func FindChannelByID(id string) *IrcChannel {
	channelsMu.RLock()
	defer channelsMu.RUnlock()
	for _, c := range Channels {
		if c.ID == id {
			c := c
			return &c
		}
	}
	return nil
}

func Connect() error {
	token := globals.GetState().GetKickUser().Token
	if token == "" {
		return fmt.Errorf("kick Token not found")
	}

	url := "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679?protocol=7&client=js&version=8.4.0-rc2&flash=false" // Kick chat endpoint
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	//defer conn.Close()
	if Conn != nil {
		Conn.Close()
	}
	Conn = conn
	log.Printf("[Kick IRC] Connected to IRC")

	go reader()
	go writer()
	go func() {
		helpers.Log(helpers.INFO, "Started kick ping!")
		ticker := time.NewTicker(4 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			Conn.WriteJSON(map[string]string{"event": "pusher:ping"})
		}
	}()
	return nil
}

func JoinChannel(channel string) {
	helpers.Logf(helpers.DEBUG, "[Kick IRC] Connecting to channel: %s", channel)
	token := globals.GetState().GetKickUser().Token
	if token == "" {
		helpers.Logf(helpers.ERROR, "[Kick IRC] Token not found")
		return
	}
	// Subscribe to authenticated chat
	subscribe := map[string]interface{}{
		"event": "pusher:subscribe",
		"data": map[string]interface{}{
			"channel": fmt.Sprintf("chatrooms.%s.v2", channel),
			"auth":    token,
		},
	}
	if err := Conn.WriteJSON(subscribe); err != nil {
		log.Println("[Kick IRC] Error sending subscribe:", err)
		return
	}
}

func reader() {
	helpers.Log(helpers.INFO, "Started kick reader!")
	for {
		_, msg, err := Conn.ReadMessage()
		if err != nil {
			helpers.Logf(helpers.ERROR, "[Kick IRC] ReadMessage erro: %s", err.Error())
			return
		}
		helpers.Printf(helpers.Kick, "[Kick IRC] Message: %s", msg)

		// Parse JSON message
		var km KickMessage
		if err := json.Unmarshal(msg, &km); err != nil {
			helpers.Logf(helpers.ERROR, "[Kick IRC] Error parsing JSON: %v", err)
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
	helpers.Log(helpers.INFO, "Started kick writer!")
	for msg := range MsgQueue {
		PostMessage(msg)
	}
}

func SendMessageIfChannelExist(msg string, channel string) {
	c := FindChannelByID(channel)
	if c == nil {
		helpers.Logf(helpers.ERROR, "[Kick] Channel not found: %s", channel)
		return
	}
	SendMessage(msg, *c)
}

func SendMessage(msg string, channel IrcChannel) {
	MsgQueue <- Message{Channel: channel, Text: msg}
}
