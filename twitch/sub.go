package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type EventSubCondition struct {
	UserId              string `json:"user_id"`
	BroadcasterUserId   string `json:"broadcaster_user_id"`
	ModeratorUserId     string `json:"moderator_user_id"`
	ToBroadcasterUserId string `json:"to_broadcaster_user_id"`
}

type EventSubTransport struct {
	Method    string `json:"method"`
	SessionId string `json:"session_id"`
}

type EventSub struct {
	Type      string            `json:"type"`
	Version   int               `json:"version"`
	Condition EventSubCondition `json:"condition"`
	Transport EventSubTransport `json:"transport"`
}

type SessionWelcome struct {
	Metadata struct {
		MessageType string `json:"message_type"`
	} `json:"metadata"`
	Payload struct {
		Session struct {
			ID               string  `json:"id"`
			Status           string  `json:"status"`
			KeepaliveTimeout int     `json:"keepalive_timeout_seconds"`
			ReconnectURL     *string `json:"reconnect_url"`
		} `json:"session"`
	} `json:"payload"`
}

var subTypes = map[string]map[string]any{
	// //automod.message.hold,
	// //automod.message.update,
	// //automod.settings.update,
	// //automod.terms.update,
	// //channel.update,
	"channel.follow": {"version": 2, "requires": "moderator:read:followers"},
	// channel.ad_break.begin,
	// channel.chat.clear,
	// channel.chat.clear_user_messages,
	//'channel.chat.message',
	//"channel.chat.message_delete": {},
	// channel.chat.notification,
	// //channel.chat_settings.update,
	// //channel.chat.user_message_hold,
	// //channel.chat.user_message_update,
	"channel.shared_chat.begin":  {},
	"channel.shared_chat.update": {},
	"channel.shared_chat.end":    {},
	// channel.subscribe,
	// channel.subscription.end,
	// channel.subscription.gift,
	// channel.subscription.message,
	// channel.cheer,
	"channel.raid": {},
	"channel.ban":  {"requires": "channel:moderate"},
	// channel.unban,
	// channel.unban_request.create,
	// channel.unban_request.resolve,
	//"channel.moderate',
	//{ type: "channel.moderate", version: "2' },
	// channel.moderator.add,
	// channel.moderator.remove,
	// //channel.guest_star_session.begin,
	// //channel.guest_star_session.end,
	// //channel.guest_star_guest.update,
	// //channel.guest_star_settings.update,
	"channel.channel_points_automatic_reward_redemption.add": {"requires": "channel:manage:redemptions"},
	//"channel.channel_points_custom_reward.add":               {},
	//"channel.channel_points_custom_reward.update":            {},
	//"channel.channel_points_custom_reward.remove":            {},
	//"channel.channel_points_custom_reward_redemption.add":    {},
	//"channel.channel_points_custom_reward_redemption.update": {},
	//"channel.poll.begin":                                     {},
	//"channel.poll.progress":                                  {},
	//"channel.poll.end":                                       {},
	//"channel.prediction.begin":                               {},
	//"channel.prediction.progress":                            {},
	//"channel.prediction.lock":                                {},
	//"channel.prediction.end":                                 {},
	// channel.suspicious_user.message,
	// channel.suspicious_user.update,
	//"channel.vip.add":    {},
	//"channel.vip.remove": {},
	// channel.warning.acknowledge,
	// channel.warning.send,
	// channel.charity_campaign.donate,
	// channel.charity_campaign.start,
	// channel.charity_campaign.progress,
	// channel.charity_campaign.stop,
	// conduit.shard.disabled,
	// drop.entitlement.grant,
	// extension.bits_transaction.create,
	// channel.goal.begin,
	// channel.goal.progress,
	// channel.goal.end,
	// channel.hype_train.begin,
	// channel.hype_train.progress,
	// channel.hype_train.end,
	// channel.shield_mode.begin,
	// channel.shield_mode.end,
	// channel.shoutout.create,
	// channel.shoutout.receive,
	"stream.online":  {},
	"stream.offline": {},
	// user.authorization.grant,
	// user.authorization.revoke,
	// user.update,
	// user.whisper.message
}

var messageHandlers = map[string]func(map[string]any, map[string]any){
	"session_welcome": func(payload, metadata map[string]any) {
		globals.GetState().SetTwitchEventSubId(payload["session"].(map[string]any)["id"].(string))
		subscribeToEvents()
		//ts.execute("session_welcome", payload);
		j, _ := json.Marshal(map[string]any{
			"payload":  payload,
			"metadata": metadata,
		})
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "twitch-eventsub-session-welcome",
			Data: string(j),
		}
	},
	"session_keepalive": func(payload, metadata map[string]any) {
		//helpers.Logf(helpers.Twitch, "[TWITCH EventSub] Session Keepalive %v", metadata)
		//ts.execute("session_keepalive", metadata);
		j, _ := json.Marshal(map[string]any{
			"payload":  payload,
			"metadata": metadata,
		})
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "twitch-eventsub-keepalive",
			Data: string(j),
		}
	},
	"notification": func(payload, metadata map[string]any) {
		helpers.Logf(helpers.Twitch, "[TWITCH EventSub] notification %v", payload)
		//ts.execute(metadata.subscription_type, payload.event, payload.subscription);
		j, _ := json.Marshal(map[string]any{
			"payload":  payload,
			"metadata": metadata,
		})
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "twitch-eventsub-notification",
			Data: string(j),
		}
		globals.EventQueue <- globals.LuaEvent{
			Type: payload["subscription"].(map[string]any)["type"].(string),
			Data: map[string]any{
				"payload":  payload,
				"metadata": metadata,
			},
		}
	},
}

var (
	//EventSubConn *websocket.Conn
	//quitChan   = make(chan struct{})
	eventSubMu sync.RWMutex
)

func connectToEventSub() {
	u := url.URL{Scheme: "wss", Host: "eventsub.wss.twitch.tv", Path: "/ws"}
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("[Twitch] Falha no handshake (%d): %s", resp.StatusCode, string(body))
		}
		log.Printf("[Twitch] Erro ao conectar: %v", err)
		time.Sleep(10 * time.Second)
		//StartEventSub(clientID, token, broadcasterID)
		return
	}
	eventSubMu.Lock()
	//EventSubConn = conn

	messageHandlers["session_reconnect"] = func(payload, metadata map[string]any) {
		reconnectURL := payload["session"].(map[string]any)["reconnect_url"].(string)
		helpers.Logf(helpers.Yellow, "[Twitch EventSub] Reconnect solicitado: %s", reconnectURL)
		eventSubMu.Lock()
		defer eventSubMu.Unlock()
		conn.Close()
		//EventSubConn.Close()
		conn, _, err := websocket.DefaultDialer.Dial(reconnectURL, nil)
		if err != nil {
			log.Printf("[Twitch EventSub] Falha ao reconectar: %v", err)
			time.Sleep(5 * time.Second)
			connectToEventSub()
			return
		}

		go listenToEventSub(conn)
		//EventSubConn = conn
	}
	helpers.Logf(helpers.Twitch, "[Twitch EventSub] Conexão WebSocket aberta com sucesso!")
	eventSubMu.Unlock()

	go listenToEventSub(conn)
}

func listenToEventSub(conn *websocket.Conn) {
	defer func() {
		if conn != nil {
			conn.Close()
		}
		helpers.Logf(helpers.Twitch, "[Twitch EventSub] Leitura encerrada.")
		connectToEventSub()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			helpers.Logf(helpers.Red, "[Twitch EventSub] erro de leitura: %v", err)
			break // <- Sai naturalmente do loop
		}

		var base map[string]any
		if err := json.Unmarshal(msg, &base); err != nil {
			log.Println("[Twitch EventSub] Erro ao decodificar JSON:", err)
			continue
		}

		meta, ok := base["metadata"].(map[string]any)
		if !ok {
			continue
		}

		handler := messageHandlers[meta["message_type"].(string)]

		if handler == nil {
			helpers.Logf(helpers.Red, "[TWITCH EventSub] Handler not found %s", meta["message_type"])
			continue
		}

		handler(base["payload"].(map[string]any), base["metadata"].(map[string]any))
	}

	/*for {
		select {
		case <-quitChan:
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				helpers.Logf(helpers.Twitch, "[Twitch EventSub] Erro ao ler: %v", err)
				return
			}

			var base map[string]any
			if err := json.Unmarshal(msg, &base); err != nil {
				log.Println("[Twitch EventSub] Erro ao decodificar JSON:", err)
				continue
			}

			meta, ok := base["metadata"].(map[string]any)
			if !ok {
				continue
			}

			handler := messageHandlers[meta["message_type"].(string)]

			if handler == nil {
				helpers.Logf(helpers.Red, "[TWITCH EventSub] Handler not found %s", meta["message_type"])
				continue
			}

			handler(base["payload"].(map[string]any), base["metadata"].(map[string]any))
		}
	}*/
}

func subscribeToEvents() {
	var data = EventSub{
		Type:    "",
		Version: 1,
		Condition: EventSubCondition{
			UserId:              globals.GetState().GetTwitchUser().UserID,
			BroadcasterUserId:   globals.GetState().GetTwitchUser().UserID,
			ModeratorUserId:     globals.GetState().GetTwitchUser().UserID,
			ToBroadcasterUserId: globals.GetState().GetTwitchUser().UserID,
		},
		Transport: EventSubTransport{
			Method:    "websocket",
			SessionId: globals.GetState().GetTwitchEventSubId(),
		},
	}
	e := globals.GetState().GetData("TwitchSubEventsConnectedEvents")
	events := []string{}
	if e != nil {
		events = e.([]string)
	}
	for name, sub := range subTypes {
		data.Type = name
		data.Version = 1
		if sub["version"] != nil {
			data.Version = sub["version"].(int)
		}
		jsonData, _ := json.Marshal(data)
		req, _ := http.NewRequest("POST", urlAPIEventSub, bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
		req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 204 {
			body, _ := io.ReadAll(resp.Body)
			var d struct {
				Error   string `json:"error"`
				Status  int    `json:"status"`
				Message string `json:"message"`
				Data    []any  `json:"data"`
			}
			_ = json.Unmarshal(body, &d)

			if d.Status != 0 || len(d.Data) == 0 {
				helpers.Logf(helpers.Red, "[TWITCH EventSub] %s: %s", d.Error, d.Message)
				continue
			}

			cd := d.Data[0].(map[string]any)
			//helpers.Logf(helpers.Red, "[TWITCH EventSub] %v", cd)
			events = append(events, cd["type"].(string))

			if cd["max_total_cost"] != nil && cd["total_cost"] != nil && cd["max_total_cost"].(int) < cd["total_cost"].(int) {
				helpers.Logf(helpers.Red, "FODEU MANÉ LOTO OS COST TUDO!")
			}

			//return fmt.Errorf("erro ao excluir mensagem: %s", body)
		}
		globals.GetState().SetData("TwitchSubEventsConnectedEvents", events)
	}
}
