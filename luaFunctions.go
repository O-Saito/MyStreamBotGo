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

func tableToMap(tbl *lua.LTable) map[string]interface{} {
	result := make(map[string]interface{})
	tbl.ForEach(func(key lua.LValue, value lua.LValue) {
		switch v := value.(type) {
		case lua.LString:
			result[key.String()] = string(v)
		case lua.LNumber:
			result[key.String()] = float64(v)
		case lua.LBool:
			result[key.String()] = bool(v)
		case *lua.LTable:
			result[key.String()] = tableToMap(v) // recursivo
		default:
			result[key.String()] = v.String()
		}
	})
	return result
}

func RegisterLuaFunctions(L *lua.LState) {
	mlua.ExposeServiceToLua(L, "g", map[string]func(*lua.LState) int{
		"log": func(L *lua.LState) int {
			if L.Get(2) == lua.LNil {
				helpers.Logf(helpers.Lua, "[LUA LOG] %s", L.CheckString(1))
				return 0
			}

			table := L.CheckTable(2)
			jsonData, _ := json.Marshal(tableToMap(table))
			helpers.Logf(helpers.Lua, "[LUA LOG] %s: %s", L.CheckString(1), jsonData)
			return 0
		},
		"print": func(L *lua.LState) int {
			if L.Get(2) == lua.LNil {
				helpers.Logf(helpers.Lua, "[LUA PRINT] %s", L.CheckString(1))
				return 0
			}
			table := L.CheckTable(2)
			jsonData, _ := json.Marshal(tableToMap(table))
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
			jsonData, _ := json.Marshal(tableToMap(L.CheckTable(2)))
			helpers.Logf(helpers.Lua, "[LUA SOCKET_SEND] %s; %s", t, jsonData)
			globals.WsBroadcast <- globals.SocketMessage{
				Type: t,
				Data: string(jsonData),
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
	})
}
