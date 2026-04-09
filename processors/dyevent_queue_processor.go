package processors

import (
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"sync"
	"time"
)

func ProcessDyEventQueue() {
	helpers.Log(helpers.INFO, "Started dyevents processor!")
	for deq := range mlua.DyEventQueue {
		events := mlua.GetDyEvents()
		wait := sync.WaitGroup{}
		wait.Add(len(events))
		for _, dev := range events {
			go func() {
				start := time.Now()
				switch deq.Type {
				case mlua.DyEventChat:
					dev.ProcessChat(&deq.MessageFromStream)
				case mlua.DyEventCommand:
					dev.ProcessCommand(&deq.LuaCommand)
				case mlua.DyEventEvent:
					dev.ProcessEvent(&deq.LuaEvent)
				case mlua.DyEventRequest:
					dev.ProcessRequest(&deq.SocketMessage)
				default:
				}
				elapsed := time.Since(start)
				helpers.Logf(helpers.DEBUG, "PROCESSED DY EVENT %s[%d] IN %v", dev.Name, deq.Type, elapsed)
				wait.Done()
			}()
		}
		helpers.Logf(helpers.DEBUG, "WAINTING DY PROCESSOR")
		wait.Wait()
		helpers.Logf(helpers.DEBUG, "WAINTED DY PROCESSOR")
	}
}
