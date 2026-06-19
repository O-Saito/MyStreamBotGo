package loader

/*
#include "hostapi.h"

extern bot_host_api_t* create_host_api(void);
extern void            destroy_host_api(bot_host_api_t*);
*/
import "C"
import (
	"encoding/json"
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
)

//export hostLog
func hostLog(level C.int, msg *C.char) {
	helpers.Logf(helpers.Level(level), "[PLUGIN] %s", C.GoString(msg))
}

//export hostEmitEvent
func hostEmitEvent(etype, data, target *C.char) *C.char {
	et := C.GoString(etype)
	d := []byte(C.GoString(data))
	t := C.GoString(target)

	var parsed map[string]any
	if len(d) > 0 {
		_ = json.Unmarshal(d, &parsed)
	}

	if t != "" {
		globals.LuaRequest <- globals.SocketMessage{
			Filter: t + ".lua",
			Type:   et,
			Data:   parsed,
		}
	} else {
		globals.EventQueue <- globals.Event{Type: et, Data: parsed}
	}

	return nil
}

func newHostAPI() *C.bot_host_api_t {
	return C.create_host_api()
}

func deleteHostAPI(api *C.bot_host_api_t) {
	C.destroy_host_api(api)
}
