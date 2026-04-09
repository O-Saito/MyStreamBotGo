package processors

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
)

func ProcessLuaRequest() {
	helpers.Log(helpers.INFO, "Started lua requests queue!")
	defer func() {
		if r := recover(); r != nil {
			helpers.Logf(helpers.ERROR, "[LuaRequest] panic: %v", r)
		}
	}()
	for ev := range globals.LuaRequest {
		mlua.DyEventQueue <- mlua.DyEventQueueData{
			Type:          mlua.DyEventRequest,
			SocketMessage: ev,
		}
	}
}
