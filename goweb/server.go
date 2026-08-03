package goweb

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"MyStreamBot/services/kick"
	"MyStreamBot/services/twitch"
	tf "MyStreamBot/services/twitch/fetch"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type SocketRequestMetadata struct {
	Tag int    `json:"tag"`
	ID  string `json:"id"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (srm *SocketRequestMetadata) Respond(t string, data any) {
	globals.SafeSend(globals.WsBroadcast, globals.SocketMessage{
		SocketTag:         srm.Tag,
		ResponseMessageID: srm.ID,
		Type:              t,
		Data:              data,
	}, "WsBroadcast", false)
}

var mu sync.RWMutex
var lastTagIndex = 0
var wsClients = make(map[*websocket.Conn]int)
var wsClientsUpgraded = make(map[string][]*websocket.Conn)

var SocketHandlers = map[string]func(*websocket.Conn, map[string]any, *SocketRequestMetadata){
	"init": func(c *websocket.Conn, m map[string]any, md *SocketRequestMetadata) {
		globals.SafeSend(globals.WsBroadcast, globals.SocketMessage{
			Type: "init",
			Data: map[string]any{
				"twitch":  globals.GetState().GetTwitchUser(),
				"youtube": globals.GetState().GetYouTubeUser(),
				"kick": map[string]any{
					"connected_as": globals.GetState().GetKickUser().UserLogin,
				},
				"twitch_connected_chat":  twitch.GetChannels(),
				"kick_connected_chat":    kick.GetChannels(),
				"youtube_connected_chat": globals.GetState().GetData("youtube-lives"),
				"youtube_live_previews":  globals.GetState().GetData("youtube-preview-lives"),
				"custom_events_modules":  mlua.ListDynamicEvents(),
				"twitch_eventsubs":       globals.GetState().GetData("TwitchSubEventsConnectedEvents"),
				"interface_config":       globals.GetState().GetData("htmlinterface"),
			},
			SocketTag: md.Tag,
		}, "WsBroadcast", false)
	},
	"upgrade-conn": func(c *websocket.Conn, m map[string]any, md *SocketRequestMetadata) {
		data := globals.SocketMessage{
			Type:      "response-upgrade",
			Data:      "",
			SocketTag: md.Tag,
		}
		if m["conn"] != nil {
			connVal, ok := m["conn"].(string)
			if !ok {
				helpers.Logf(helpers.ERROR, "[WebSocket] upgrade-conn: invalid conn type")
				return
			}
			if connVal == "ignore-broadcast" {
				mu.Lock()
				wsClients[c] = -1
				mu.Unlock()
				return
			}
			mu.Lock()
			if wsClientsUpgraded[connVal] == nil {
				wsClientsUpgraded[connVal] = make([]*websocket.Conn, 0)
			}

			wsClientsUpgraded[connVal] = append(wsClientsUpgraded[connVal], c)
			mu.Unlock()
			data.Data = "Connection updated!"
		} else {
			data.Data = "Connection not specified!"
		}

		globals.SafeSend(globals.WsBroadcast, data, "WsBroadcast", false)
	},
}

func StartHTTPServer() {
	// Serve frontend
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// WebSocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[WebSocket] Error: %s", err.Error())
			return
		}
		defer conn.Close()
		mu.Lock()
		lastTagIndex += 1
		mytag := lastTagIndex
		wsClients[conn] = mytag
		mu.Unlock()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				helpers.Logf(helpers.ERROR, "[Socket] Error: %v", err)
				mu.Lock()
				delete(wsClients, conn)
				mu.Unlock()
				break
			}

			//helpers.Logf(helpers.Cyan, "[Socket] Message: %s", string(msg))
			m := string(msg)
			if m == "init" {
				SocketHandlers["init"](conn, nil, &SocketRequestMetadata{
					Tag: mytag,
				})
				continue
			}
			var data globals.SocketMessage
			err = json.Unmarshal([]byte(m), &data)

			if err != nil {
				helpers.Logf(helpers.ERROR, "[Socket] Invalid message format: %s", err.Error())
				continue
			}

			if data.Filter != "" {
				data.SocketTag = mytag
				globals.LuaRequest <- data
				continue
			}

			if handler, exists := SocketHandlers[string(data.Type)]; exists {
				dataMap, ok := data.Data.(map[string]any)
				if !ok {
					helpers.Logf(helpers.ERROR, "[WebSocket] Invalid message data type")
					continue
				}
				handler(conn, dataMap, &SocketRequestMetadata{
					Tag: mytag,
					ID:  data.ResponseMessageID,
				})
				continue
			}
		}
	})

	// Administrative endpoints
	http.HandleFunc("/admin/delete/twitch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Message string `json:"message"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		tf.DeleteMessage(req.Message)
		//if IrcHandler != nil {
		//	IrcHandler.SendMessage("/delete " + req.Message)
		//}
		w.WriteHeader(200)
	})

	http.HandleFunc("/admin/ban/twitch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			UserId   string `json:"userId"`
			Duration int32  `json:"duration"`
			Reason   string `json:"reason"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		d, err := tf.BanUser(req.UserId, req.Duration, req.Reason)
		//if IrcHandler != nil {
		//	IrcHandler.SendMessage("/ban " + req.User)
		//}
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("Error %v", err)))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(d))
	})

	http.HandleFunc("/admin/automod/twitch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			UserId string `json:"userId"`
			MsgId  string `json:"msgId"`
			Action string `json:"action"`
		}

		json.NewDecoder(r.Body).Decode(&req)
		if req.Action != "ALLOW" && req.Action != "DENY" {
			w.WriteHeader(400)
			w.Write([]byte("action should be ALLOW or DENY."))
			return
		}
		d, err := twitch.UpdateAutomod(req.UserId, req.MsgId, req.Action)
		if err != nil {
			w.WriteHeader(200)
			w.Write([]byte(fmt.Sprintf("Error %v", err)))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(d))
	})

	// Goroutine to send messages from backend to all clients
	go func() {
		helpers.Log(helpers.INFO, "Started server broadcast!")
		for msg := range globals.WsBroadcast {

			if msg.Filter != "" {
				mu.RLock()
				wsList := append([]*websocket.Conn(nil), wsClientsUpgraded[msg.Filter]...)
				mu.RUnlock()
				if len(wsList) == 0 {
					continue
				}
				jsonData, err := json.Marshal(msg)
				if err != nil {
					helpers.Logf(helpers.ERROR, "[WebSocket] Broadcast filtered json.Marshal failed: %v", err)
					continue
				}
				for _, client := range wsList {
					_ = client.WriteMessage(websocket.TextMessage, []byte(jsonData))
				}
				continue
			}

			//helpers.Logf(helpers.Cyan, "[WebSocket] Broadcast: %s - %s", msg.Type, msg.Data)
			jsonData, err := json.Marshal(msg)
			if err != nil {
				helpers.Logf(helpers.ERROR, "[WebSocket] Broadcast json.Marshal failed: %v", err)
				continue
			}
			mu.RLock()
			clientsList := make([]*websocket.Conn, 0, len(wsClients))
			for client, tag := range wsClients {
				if tag == -1 {
					continue
				}
				if msg.SocketTag != 0 && msg.SocketTag != -1 {
					if msg.SocketTag != tag {
						continue
					}
				}
				clientsList = append(clientsList, client)
			}
			mu.RUnlock()
			for _, client := range clientsList {
				client.WriteMessage(websocket.TextMessage, []byte(jsonData))
			}
		}
	}()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		helpers.Logf(helpers.ERROR, "[WebSocket] Failed to get interface address: %s", err.Error())
		helpers.Log(helpers.ERROR, "[WebSocket] Exiting due to unrecoverable error")
		os.Exit(1)
	}

	port := globals.GetConfig().HTTPPort
	helpers.Print(helpers.Green, "[MyStreamBot] Possible IPs (logins MUST be via localhost):")
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				helpers.Printf(helpers.Reset, "http://%s:%s", ipnet.IP.String(), port)
			}
		}
	}
	helpers.Printf(helpers.Green, "[MyStreamBot] HTTP server started at http://localhost:%s", port)
	helpers.Log(helpers.INFO, "Started listen and serve (http started)!")
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil); err != nil && err != http.ErrServerClosed {
			helpers.Logf(helpers.ERROR, "[HTTP] Server failed: %v", err)
			helpers.Log(helpers.ERROR, "[HTTP] Exiting due to unrecoverable error")
			os.Exit(1)
		}
	}()
}
