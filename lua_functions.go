package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"MyStreamBot/services"
	"MyStreamBot/services/kick"
	"MyStreamBot/services/twitch"
	"MyStreamBot/services/youtube"
	"encoding/json"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type TwitchStateLuaData struct {
	UserID                 string
	UserLogin              string
	DisplayName            string
	Type                   string
	BroadcasterType        string
	Description            string
	ProfileImageURL        string
	ProfileOfflineImageURL string
}

type TwitchStreamLuaData struct {
	ID           string
	GameId       string
	GameName     string
	Type         string
	Title        string
	Tags         []string
	ViewerCount  int32
	StartedAt    string
	Language     string
	ThumbnailURL string
	IsMature     bool
}

var KickFunctionList = []services.LuaFunction{
	{Name: "get_user", Fn: kick.GetUser},
	{Name: "get_channel", Fn: kick.GetChannel},
	{Name: "get_chatroom", Fn: kick.GetChatroom},
}

var YouTubeFunctionList = []services.LuaFunction{
	{Name: "get_state", Fn: func() []globals.YouTubeChannel {
		return globals.GetState().GetYouTubeUser().Channels
	}},
	{Name: "get_current_youtube_channel", Fn: youtube.GetCurrentYouTubeChannel},
	{Name: "get_current_streamings", Fn: youtube.GetCurrentStreamings},
}

func RegisterLuaFunctions(L *lua.LState) {
	mlua.ExposeServiceToLua(L, "g", map[string]func(*lua.LState) int{
		"log": func(L *lua.LState) int {
			if L.Get(2) == lua.LNil {
				helpers.Logf(helpers.ERROR, "[LUA g.log] %s", L.CheckString(1))
				return 0
			}

			table := L.CheckTable(2)
			jsonData, err := json.Marshal(mlua.TableToMap(table))
			if err != nil {
				helpers.Logf(helpers.ERROR, "[LUA g.log] json.Marshal failed: %v", err)
			} else {
				helpers.Logf(helpers.DEBUG, "[LUA g.log] %s: %s", L.CheckString(1), jsonData)
			}
			return 0
		},
		"print": func(L *lua.LState) int {
			if L.Get(2) == lua.LNil {
				helpers.Logf(helpers.ERROR, "[LUA g.print] %s", L.CheckString(1))
				return 0
			}
			table := L.CheckTable(2)
			jsonData, err := json.Marshal(mlua.TableToMap(table))
			if err != nil {
				helpers.Logf(helpers.ERROR, "[LUA g.print] json.Marshal failed: %v", err)
			} else {
				helpers.Printf(helpers.Lua, "[LUA g.print] %s: %s", L.CheckString(1), jsonData)
			}
			return 0
		},
		"socket_send": func(L *lua.LState) int {
			defer func() {
				if r := recover(); r != nil {
					helpers.Logf(helpers.ERROR, "[LUA SOCKET_SEND PANIC] %v", r)
				}
			}()
			t := L.CheckString(1)
			t2 := L.CheckTable(2)
			data := mlua.TableToMap(t2)
			helpers.Printf(helpers.Lua, "[LUA g.socket_send] %s; %v", t, data)
			globals.WsBroadcast <- globals.SocketMessage{
				Type: t,
				Data: data,
			}
			return 0
		},
		"send_message": func(L *lua.LState) int {
			source := L.CheckString(1)
			channel := L.CheckString(2)
			msg := L.CheckString(3)
			reply := ""
			if L.Get(4) != lua.LNil {
				reply = L.CheckString(4)
			}
			helpers.Printf(helpers.Lua, "[LUA g.send_message] {%s} %s: %s [%s]", source, channel, msg, reply)
			msg = strings.Trim(msg, " ")
			if source == "twitch" {
				twitch.SendMessage(msg, channel, reply)
			}
			if source == "kick" {
				kick.SendMessageIfChannelExist(msg, channel)
			}
			if source == "youtube" {
				youtube.SendMessage(channel, msg)
			}
			return 0
		},
		"get": func(L *lua.LState) int {
			key := L.CheckString(1)
			val := globals.GetState().GetData(key)
			helpers.Printf(helpers.Lua, "[LUA g.get] %s: %v", key, val)
			L.Push(mlua.ToLValue(L, val))
			return 1
		},
		"set": func(L *lua.LState) int {
			key := L.CheckString(1)
			val := mlua.FromLValue(L, L.Get(2))
			helpers.Printf(helpers.Lua, "[LUA g.set] %s: %v", key, val)
			globals.GetState().SetData(key, val)
			return 0
		},
		"kv_get": func(L *lua.LState) int {
			key := L.CheckString(1)
			v, err := globals.GetGlobalDB().KVGet(key)
			if err != nil {
				helpers.Logf(helpers.ERROR, "[LUA g.kv_get] %s", err.Error())
				return 0
			}
			helpers.Printf(helpers.Lua, "[LUA g.kv_get] %s: %v", key, v)
			L.Push(mlua.ToLValue(L, v))
			return 1
		},
		"kv_set": func(L *lua.LState) int {
			key := L.CheckString(1)
			val := mlua.FromLValue(L, L.Get(2))
			helpers.Printf(helpers.Lua, "[LUA g.kv_set] %s: %v", key, val)
			err := globals.GetGlobalDB().KVSet(key, val)
			if err != nil {
				helpers.Logf(helpers.ERROR, "[LUA g.kv_set] %s", err.Error())
				return 0
			}
			return 0
		},
	})

}

func RegisterServiceAPIs(L *lua.LState) {
	services.ExposeToLua(L, "twitch", TwitchFunctionList)
	services.ExposeToLua(L, "kick", KickFunctionList)
	services.ExposeToLua(L, "youtube", YouTubeFunctionList)
}
