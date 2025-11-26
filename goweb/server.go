package goweb

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/kick"
	"MyStreamBot/mlua"
	"MyStreamBot/twitch"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var mu sync.RWMutex
var lastTagIndex = 0
var wsClients = make(map[*websocket.Conn]int)
var wsClientsUpgraded = make(map[string][]*websocket.Conn)

var SocketHandlers = map[string]func(*websocket.Conn, map[string]any, int){
	"init": func(c *websocket.Conn, m map[string]any, mytag int) {
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "init",
			Data: map[string]any{
				"twitch":  globals.GetState().GetTwitchUser(),
				"youtube": globals.GetState().GetYouTubeUser(),
				"kick": map[string]any{
					"connected_as": kick.UserLogin,
				},
				"twitch_connected_chat": twitch.Channels,
				"kick_connected_chat":   kick.Channels,
				"custom_events_modules": mlua.ListDynamicEvents(),
				"twitch_eventsubs":      globals.GetState().GetData("TwitchSubEventsConnectedEvents"),
			},
			Respond: mytag,
		}
		//c.WriteMessage(websocket.TextMessage, []byte(d))
	},
	"upgrade-conn": func(c *websocket.Conn, m map[string]any, mytag int) {
		data := globals.SocketMessage{
			Type:    "response-upgrade",
			Data:    "",
			Respond: mytag,
		}
		if m["conn"] != nil {
			if m["conn"].(string) == "ignore-broadcast" {
				mu.Lock()
				wsClients[c] = -1
				mu.Unlock()
				return
			}
			mu.Lock()
			if wsClientsUpgraded[m["conn"].(string)] == nil {
				wsClientsUpgraded[m["conn"].(string)] = make([]*websocket.Conn, 0)
			}

			wsClientsUpgraded[m["conn"].(string)] = append(wsClientsUpgraded[m["conn"].(string)], c)
			mu.Unlock()
			data.Data = "conexão atualizada!"
		} else {
			data.Data = "conexão não especificada!"
		}

		globals.WsBroadcast <- data
		//c.WriteMessage(websocket.TextMessage, []byte(d))
	},
}

func StartHTTPServer() {
	// Servir frontend
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// WebSocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			helpers.Logf(helpers.Red, "[WebSocket] Erro: %s", err.Error())
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
				helpers.Logf(helpers.Red, "[Socket] Error: %v", err)
				delete(wsClients, conn)
				break
			}

			//helpers.Logf(helpers.Cyan, "[Socket] Message: %s", string(msg))
			m := string(msg)
			if m == "init" {
				SocketHandlers["init"](conn, nil, mytag)
				continue
			}
			var data globals.SocketMessage
			json.Unmarshal([]byte(m), &data)

			if data.Filter != "" {
				globals.LuaRequest <- data
				continue
			}

			if handler, exists := SocketHandlers[string(data.Type)]; exists {
				handler(conn, data.Data.(map[string]any), mytag)
				continue
			}
		}
	})

	// Endpoints administrativos
	http.HandleFunc("/admin/delete/twitch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Message string `json:"message"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		twitch.DeleteMessage(req.Message)
		//if IrcHandler != nil {
		//	IrcHandler.SendMessage("/delete " + req.Message)
		//}
		w.WriteHeader(200)
	})

	http.HandleFunc("/admin/ban/twitch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			UserId   string `json:"userId"`
			Duration int32  `json:"duration"`
			Reason   string `json:"reason"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		d, err := twitch.BanUser(req.UserId, req.Duration, req.Reason)
		//if IrcHandler != nil {
		//	IrcHandler.SendMessage("/ban " + req.User)
		//}
		if err != nil {
			w.WriteHeader(200)
			w.Write([]byte(fmt.Sprintf("Error %v", err)))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(d))
	})

	http.HandleFunc("/admin/automod/twitch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
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
		//if IrcHandler != nil {
		//	IrcHandler.SendMessage("/ban " + req.User)
		//}
		if err != nil {
			w.WriteHeader(200)
			w.Write([]byte(fmt.Sprintf("Error %v", err)))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(d))
	})

	// Goroutine para enviar mensagens do backend para todos os clients
	go func() {
		for msg := range globals.WsBroadcast {

			if msg.Filter != "" {
				//helpers.Logf(helpers.Cyan, "[WebSocket] Message filter %s: %s - %s", msg.Filter, msg.Type, msg.Data)
				mu.RLock()
				wsList := append([]*websocket.Conn(nil), wsClientsUpgraded[msg.Filter]...)
				mu.RUnlock()
				if len(wsList) == 0 {
					continue
				}
				jsonData, _ := json.Marshal(msg)
				for _, client := range wsList {
					_ = client.WriteMessage(websocket.TextMessage, []byte(jsonData))
				}
				continue
			}

			//helpers.Logf(helpers.Cyan, "[WebSocket] Broadcast: %s - %s", msg.Type, msg.Data)
			jsonData, _ := json.Marshal(msg)
			mu.RLock()
			clientsList := make([]*websocket.Conn, 0, len(wsClients))
			for client, tag := range wsClients {
				if tag == -1 {
					continue
				}
				if msg.Respond != 0 && msg.Respond != -1 {
					if msg.Respond != tag {
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
		panic(err)
	}

	port := globals.GetConfig().HTTPPort
	helpers.Log(helpers.Reset, "[MyStreamBot] Possíveis IP's (para os logins TEM que ser pelo localhost):")
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				helpers.Logf(helpers.Reset, "http://%s:%s", ipnet.IP.String(), port)
			}
		}
	}
	helpers.Logf(helpers.Reset, "[MyStreamBot] Servidor HTTP iniciado em http://localhost:%s", port)
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
}
