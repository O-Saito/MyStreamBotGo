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
				defer func() {
					if r := recover(); r != nil {
						helpers.Logf(helpers.ERROR, "[DYEVENT] goroutine panicked in %s: %v", dev.Name, r)
					}
				}()
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
				dev.Statistics.AddTiming(elapsed)
				wait.Done()
			}()
		}
		wait.Wait()
	}
}
