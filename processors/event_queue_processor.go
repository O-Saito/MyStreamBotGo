package processors

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
)

func ProcessEventQueue() {
	helpers.Log(helpers.INFO, "Started event queue processor!")
	for ev := range globals.EventQueue {
		globals.WsBroadcast <- globals.SocketMessage{Type: ev.Type, Data: ev.Data}

		mlua.DyEventQueue <- mlua.DyEventQueueData{
			Type:     mlua.DyEventEvent,
			LuaEvent: ev,
		}
		mlua.HandleEvent(ev.Type, &ev)
	}
}
