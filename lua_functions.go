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

	lua "github.com/yuin/gopher-lua"
)

var KickFunctionList = []services.LuaFunction{
	{Name: "get_user", Fn: kick.GetUser},
	{Name: "get_channel", Fn: kick.GetChannel},
	{Name: "get_chatroom", Fn: kick.GetChatroom},
	{Name: "post_message", Fn: kick.PostMessage},
}

var YouTubeFunctionList = []services.LuaFunction{
	{Name: "get_current_youtube_channel", Fn: youtube.GetCurrentYouTubeChannel},
	{Name: "get_current_streamings", Fn: youtube.GetCurrentStreamings},
	{Name: "refresh_token", Fn: youtube.RefreshToken},
}

var TwitchFunctionList = []services.LuaFunction{
	{Name: "get_cache_user_chat_color", Fn: twitch.GetCacheUserChatColor},
	{Name: "get_user_data", Fn: twitch.GetUserData},
	{Name: "get_user_data_by_id", Fn: twitch.GetUserDataById},
	{Name: "get_followers_data", Fn: twitch.GetFollowersData},
	{Name: "get_follower_data", Fn: func(userId string) ([]twitch.TwitchViewerData, error) {
		return twitch.GetFollowersData("", userId)
	}},
	{Name: "get_channel_stream_data", Fn: twitch.GetChannelStreamData},
}

func RegisterLuaFunctions(L *lua.LState) {
	mlua.ExposeServiceToLua(L, "g", map[string]func(*lua.LState) int{
		"log": func(L *lua.LState) int {
			if L.Get(2) == lua.LNil {
				helpers.Logf(helpers.ERROR, "[LUA g.log] %s", L.CheckString(1))
				return 0
			}

			table := L.CheckTable(2)
			jsonData, _ := json.Marshal(mlua.TableToMap(table))
			helpers.Logf(helpers.DEBUG, "[LUA g.log] %s: %s", L.CheckString(1), jsonData)
			return 0
		},
		"print": func(L *lua.LState) int {
			if L.Get(2) == lua.LNil {
				helpers.Logf(helpers.ERROR, "[LUA g.print] %s", L.CheckString(1))
				return 0
			}
			table := L.CheckTable(2)
			jsonData, _ := json.Marshal(mlua.TableToMap(table))
			helpers.Printf(helpers.Lua, "[LUA g.print] %s: %s", L.CheckString(1), jsonData)
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
			if source == "twitch" {
				twitch.SendMessage(msg, channel, reply)
			}
			if source == "kick" {
				kick.SendMessageIfChannelExist(msg, channel)
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
