package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/kick"
	"MyStreamBot/mlua"
	"MyStreamBot/twitch"
	"encoding/json"
	"runtime/debug"

	lua "github.com/yuin/gopher-lua"
)

func RegisterLuaFunctions(L *lua.LState) {
	mlua.ExposeServiceToLua(L, "g", map[string]func(*lua.LState) int{
		"log": func(L *lua.LState) int {
			if L.Get(2) == lua.LNil {
				helpers.Logf(helpers.Lua, "[LUA LOG] %s", L.CheckString(1))
				return 0
			}

			table := L.CheckTable(2)
			jsonData, _ := json.Marshal(mlua.TableToMap(table))
			helpers.Logf(helpers.Lua, "[LUA LOG] %s: %s", L.CheckString(1), jsonData)
			return 0
		},
		"print": func(L *lua.LState) int {
			if L.Get(2) == lua.LNil {
				helpers.Logf(helpers.Lua, "[LUA PRINT] %s", L.CheckString(1))
				return 0
			}
			table := L.CheckTable(2)
			jsonData, _ := json.Marshal(mlua.TableToMap(table))
			helpers.Logf(helpers.Lua, "[LUA PRINT] %s: %s", L.CheckString(1), jsonData)
			return 0
		},
		"socket_send": func(L *lua.LState) int {
			defer func() {
				if r := recover(); r != nil {
					helpers.Logf(helpers.Red, "[LUA SOCKET_SEND PANIC] %v", r)
					debug.PrintStack()
				}
			}()
			t := L.CheckString(1)
			t2 := L.CheckTable(2)
			data := mlua.TableToMap(t2)
			helpers.Logf(helpers.Lua, "[LUA SOCKET_SEND] %s; %v", t, data)
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
			helpers.Logf(helpers.Lua, "[LUA SEND_MESSAGE] {%s} %s: %s [%s]", source, channel, msg, reply)
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
			helpers.Logf(helpers.Lua, "[LUA GET] %s: %v", key, val)
			L.Push(mlua.ToLValue(L, val))
			return 1
		},
		"set": func(L *lua.LState) int {
			key := L.CheckString(1)
			val := mlua.FromLValue(L, L.Get(2))
			helpers.Logf(helpers.Lua, "[LUA SET] %s: %v", key, val)
			globals.GetState().SetData(key, val)
			return 0
		},
	})

	mlua.ExposeServiceToLua(L, "twitch", map[string]func(*lua.LState) int{
		"get_cache_user_chat_color": func(l *lua.LState) int {
			if L.Get(1) == lua.LNil {
				L.Push(lua.LNil)
				return 1
			}
			username := L.CheckString(1)
			d := twitch.GetCacheUserChatColor(username)
			helpers.Logf(helpers.Blue, "TESTE %v", d)
			L.Push(mlua.ToLValue(L, d))
			return 1
		},
		"get_user_data": func(l *lua.LState) int {
			if L.Get(1) == lua.LNil {
				L.Push(lua.LNil)
				return 1
			}
			username := L.CheckString(1)
			d, err := twitch.GetUserData(username)
			if err != nil {
				L.Push(lua.LNil)
				return 1
			}
			L.Push(mlua.ToLValue(L, d))
			return 1
		},
		"get_user_data_by_id": func(l *lua.LState) int {
			if L.Get(1) == lua.LNil {
				L.Push(lua.LNil)
				return 1
			}
			userid := L.CheckString(1)
			d, err := twitch.GetUserDataById(userid)
			if err != nil {
				L.Push(lua.LNil)
				return 1
			}
			L.Push(mlua.ToLValue(L, d))
			return 1
		},
		"get_followers_data": func(l *lua.LState) int {
			if L.Get(1) == lua.LNil {
				L.Push(lua.LNil)
				return 1
			}
			userid := L.CheckString(1)
			d, err := twitch.GetFollowersData("", userid)
			helpers.Logf(helpers.Red, "TESTE F %v %v", d, err)
			if err != nil {
				L.Push(lua.LNil)
				return 1
			}
			L.Push(mlua.ToLValue(L, d))
			return 1
		},
	})
}
