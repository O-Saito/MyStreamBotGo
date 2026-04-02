package processors

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
)

func ProcessCommandQueue() {
	helpers.Log(helpers.INFO, "Started command queue processor!")
	for ev := range globals.CommandQueue {
		mlua.DyEventQueue <- mlua.DyEventQueueData{
			Type:       mlua.DyEventCommand,
			LuaCommand: ev,
		}
		mlua.HandleCommand(ev.Name, &ev)
	}
}
