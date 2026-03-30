package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"slices"
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
	Method         string `json:"method"`
	SessionId      string `json:"session_id"`
	ConnectedAt    string `json:"connected_at"`
	DisconnectedAt string `json:"disconnected_at"`
}

type EventSubData struct {
	Data []struct {
		Id        string            `json:"id"`
		Status    string            `json:"status"`
		Type      string            `json:"type"`
		Version   string            `json:"version"`
		Condition EventSubCondition `json:"condition"`
		CreatedAt string            `json:"created_at"`
		Transport EventSubTransport `json:"transport"`
		Cost      int32             `json:"cost"`
	} `json:"data"`
	TotalCost    int32          `json:"total_cost"`
	MaxTotalCost int32          `json:"max_total_cost"`
	Pagination   map[string]any `json:"pagination"`
}

type EventSub struct {
	Type      string            `json:"type"`
	Version   float64           `json:"version"`
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

/*
var subTypes = map[string]map[string]any{
	"automod.message.hold": {"version": 2, "requires": "moderator:manage:automod"},
	//automod.message.update,
	//automod.settings.update,
	//automod.terms.update,
	//channel.update,
	"channel.follow": {"version": 2, "requires": "moderator:read:followers"},
	// channel.ad_break.begin,
	// channel.chat.clear,
	// channel.chat.clear_user_messages,
	//'channel.chat.message',
	//"channel.chat.message_delete": {},
	// channel.chat.notification,
	//channel.chat_settings.update,
	//channel.chat.user_message_hold,
	//channel.chat.user_message_update,
	"channel.shared_chat.begin":  {},
	"channel.shared_chat.update": {},
	"channel.shared_chat.end":    {},
	// channel.subscribe,
	// channel.subscription.end,
	// channel.subscription.gift,
	// channel.subscription.message,
	// channel.cheer,
	"channel.raid":  {},
	"channel.ban":   {"requires": "channel:moderate"},
	"channel.unban": {"requires": "channel:moderate"},
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
}*/

var messageHandlers = map[string]func(map[string]any, map[string]any){
	"session_welcome": func(payload, metadata map[string]any) {
		globals.GetState().SetTwitchEventSubId(payload["session"].(map[string]any)["id"].(string))
		subscribeToEvents()
		helpers.Printf(helpers.Twitch, "Session: %s", payload["session"].(map[string]any)["id"].(string))
		//ts.execute("session_welcome", payload);
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "twitch-eventsub-session-welcome",
			Data: map[string]any{
				"payload":  payload,
				"metadata": metadata,
			},
		}
	},
	"session_keepalive": func(payload, metadata map[string]any) {
		//helpers.Logf(helpers.Twitch, "[TWITCH EventSub] Session Keepalive %v", metadata)
		//ts.execute("session_keepalive", metadata);
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "twitch-eventsub-keepalive",
			Data: map[string]any{
				"payload":  payload,
				"metadata": metadata,
			},
		}
	},
	"notification": func(payload, metadata map[string]any) {
		helpers.Printf(helpers.Twitch, "[TWITCH EventSub] notification %v", payload)

		eventType := payload["subscription"].(map[string]any)["type"].(string)
		//ts.execute(metadata.subscription_type, payload.event, payload.subscription);
		globals.EventQueue <- globals.Event{
			Type: "twitch-eventsub-notification",
			Data: map[string]any{
				"payload":  payload,
				"metadata": metadata,
			},
		}

		if payload["event"] == nil {
			return
		}
		switch eventType {
		case "stream.online":
			user := globals.GetState().GetTwitchUser()
			if user.StreamDetails == nil {
				user.StreamDetails, _ = GetStreamData(user.UserID)
				// recreate copy to garantee sync data
				u := globals.GetState().GetTwitchUser()
				u.StreamDetails = user.StreamDetails
				globals.GetState().SetTwitchUser(u)
				return
			}

			user.StreamDetails.StartedAt = payload["event"].(map[string]any)["started_at"].(string)
			globals.GetState().SetTwitchUser(user)
		case "stream.offline":
			user := globals.GetState().GetTwitchUser()
			if user.StreamDetails == nil {
				return
			}
			user.StreamDetails.StartedAt = ""
			globals.GetState().SetTwitchUser(user)
		case "channel.update":
			user := globals.GetState().GetTwitchUser()
			if user.StreamDetails == nil {
				user.StreamDetails, _ = GetStreamData(user.UserID)
				// recreate copy to garantee sync data
				u := globals.GetState().GetTwitchUser()
				u.StreamDetails = user.StreamDetails
				globals.GetState().SetTwitchUser(u)
				return
			}
			event := payload["event"].(map[string]any)
			user.StreamDetails.Title = event["title"].(string)
			user.StreamDetails.Language = event["language"].(string)
			user.StreamDetails.GameId = event["category_id"].(string)
			user.StreamDetails.GameName = event["category_name"].(string)
			if event["content_classification_labels"] != nil && slices.Contains(event["content_classification_labels"].([]string), "MatureGame") {
				user.StreamDetails.IsMature = false
			}
			globals.GetState().SetTwitchUser(user)
		}
	},
}

var (
	eventSubMu sync.RWMutex
)

func connectToEventSub() {
	u := url.URL{Scheme: "wss", Host: "eventsub.wss.twitch.tv", Path: "/ws"}
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			helpers.Logf(helpers.ERROR, "[Twitch] Falha no handshake (%d): %s", resp.StatusCode, string(body))
		}
		helpers.Logf(helpers.ERROR, "[Twitch] Erro ao conectar: %v", err)
		time.Sleep(10 * time.Second)
		//StartEventSub(clientID, token, broadcasterID)
		return
	}
	eventSubMu.Lock()
	//EventSubConn = conn

	messageHandlers["session_reconnect"] = func(payload, metadata map[string]any) {
		reconnectURL := payload["session"].(map[string]any)["reconnect_url"].(string)
		helpers.Logf(helpers.WARN, "[Twitch EventSub] Reconnect solicitado: %s", reconnectURL)
		eventSubMu.Lock()
		defer eventSubMu.Unlock()
		conn.Close()
		//EventSubConn.Close()
		conn, _, err := websocket.DefaultDialer.Dial(reconnectURL, nil)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[Twitch EventSub] Falha ao reconectar: %v", err)
			time.Sleep(5 * time.Second)
			connectToEventSub()
			return
		}

		go listenToEventSub(conn)
		//EventSubConn = conn
	}
	helpers.Printf(helpers.Twitch, "[Twitch EventSub] Conexão WebSocket aberta com sucesso!")
	eventSubMu.Unlock()

	go listenToEventSub(conn)
}

func listenToEventSub(conn *websocket.Conn) {
	helpers.Log(helpers.INFO, "Started eventsub listener!")
	defer func() {
		if conn != nil {
			conn.Close()
		}
		helpers.Printf(helpers.Twitch, "[Twitch EventSub] Leitura encerrada.")
		connectToEventSub()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			helpers.Logf(helpers.ERROR, "[Twitch EventSub] erro de leitura: %v", err)
			break // <- Sai naturalmente do loop
		}

		var base map[string]any
		if err := json.Unmarshal(msg, &base); err != nil {
			helpers.Logf(helpers.ERROR, "[Twitch EventSub] Erro ao decodificar JSON: %v", err)
			continue
		}

		meta, ok := base["metadata"].(map[string]any)
		if !ok {
			helpers.Logf(helpers.ERROR, "[Twitch EventSub] Metadata !ok")
			continue
		}

		handler := messageHandlers[meta["message_type"].(string)]

		if handler == nil {
			helpers.Logf(helpers.ERROR, "[TWITCH EventSub] Handler not found %s", meta["message_type"])
			continue
		}

		handler(base["payload"].(map[string]any), base["metadata"].(map[string]any))
	}

	helpers.Logf(helpers.INFO, "[TWITCH EventSub] Is not in the loop!")
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

	oldSubs, _ := GetEventSubscriptions()

	for _, sub := range oldSubs.Data {
		if sub.Transport.Method == "websocket" && sub.Condition.BroadcasterUserId == data.Condition.BroadcasterUserId && sub.Transport.SessionId != data.Transport.SessionId {
			DeleteEventSubscriptions(sub.Id)
		}
	}
	subTypes := globals.GetConfig().GetTwitchSubTypes()
	for name, sub := range subTypes {
		data.Type = name
		data.Version = 1
		if sub["version"] != nil {
			data.Version = sub["version"].(float64)
		}
		jsonData, _ := json.Marshal(data)
		req, _ := http.NewRequest("POST", urlAPIEventSub, bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
		req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[Twitch Sub] err: %v", err)
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
				helpers.Logf(helpers.ERROR, "[TWITCH EventSub] %s: %s", d.Error, d.Message)
				continue
			}

			cd := d.Data[0].(map[string]any)
			events = append(events, cd["type"].(string))

			if cd["max_total_cost"] != nil && cd["total_cost"] != nil && cd["max_total_cost"].(float64) < cd["total_cost"].(float64) {
				helpers.Logf(helpers.ERROR, "FODEU MANÉ LOTO OS COST TUDO!")
			}
			//helpers.Logf(helpers.INFO, "TWITCH COST FOR %s: (Total Cost: %d / Max Total Cost: %d)", name, int(cd["total_cost"].(float64)), int(cd["max_total_cost"].(float64)))

			//return fmt.Errorf("erro ao excluir mensagem: %s", body)
		}
		globals.GetState().SetData("TwitchSubEventsConnectedEvents", events)
	}
}
