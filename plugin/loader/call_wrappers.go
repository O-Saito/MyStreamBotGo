package loader

/*
#include <stdlib.h>
#include "hostapi.h"

extern const char* call_plugin_name_ull(unsigned long long fn);
extern uint32_t call_plugin_api_ver_ull(unsigned long long fn);
extern void call_plugin_set_host_ull(unsigned long long fn, const bot_host_api_t* api);
extern void call_plugin_free_ull(unsigned long long fn, void* ptr);
extern const char* call_plugin_init_ull(unsigned long long fn, const char* json);
extern void call_plugin_start_ull(unsigned long long fn);
extern void call_plugin_stop_ull(unsigned long long fn);
extern void call_plugin_on_chat_ull(unsigned long long fn, const char* json);
extern void call_plugin_on_event_ull(unsigned long long fn, const char* json);
extern const char* call_plugin_get_actions_ull(unsigned long long fn);
extern const char* call_plugin_call_action_ull(unsigned long long fn, const char* name, const char* json);
*/
import "C"
import (
	"errors"
	"unsafe"
)

func CallPluginName(fn, free uintptr) string {
	r := C.call_plugin_name_ull(C.ulonglong(fn))
	if r == nil {
		return ""
	}
	defer C.call_plugin_free_ull(C.ulonglong(free), unsafe.Pointer(r))
	return C.GoString(r)
}

func CallPluginAPIVersion(fn uintptr) uint32 {
	return uint32(C.call_plugin_api_ver_ull(C.ulonglong(fn)))
}

func CallPluginSetHost(fn uintptr, api unsafe.Pointer) {
	C.call_plugin_set_host_ull(C.ulonglong(fn), (*C.bot_host_api_t)(api))
}

func CallPluginInit(fn uintptr, configJSON string, freeFn uintptr) error {
	cjson := C.CString(configJSON)
	defer C.free(unsafe.Pointer(cjson))
	r := C.call_plugin_init_ull(C.ulonglong(fn), cjson)
	if r == nil {
		return nil
	}
	defer C.call_plugin_free_ull(C.ulonglong(freeFn), unsafe.Pointer(r))
	s := C.GoString(r)
	if s == "" {
		return nil
	}
	return errors.New(s)
}

func CallPluginStart(fn uintptr) {
	C.call_plugin_start_ull(C.ulonglong(fn))
}

func CallPluginStop(fn uintptr) {
	C.call_plugin_stop_ull(C.ulonglong(fn))
}

func CallPluginOnChat(fn uintptr, msgJSON string) {
	cmsg := C.CString(msgJSON)
	defer C.free(unsafe.Pointer(cmsg))
	C.call_plugin_on_chat_ull(C.ulonglong(fn), cmsg)
}

func CallPluginOnEvent(fn uintptr, evtJSON string) {
	cevt := C.CString(evtJSON)
	defer C.free(unsafe.Pointer(cevt))
	C.call_plugin_on_event_ull(C.ulonglong(fn), cevt)
}

func CallPluginGetActions(fn, free uintptr) string {
	r := C.call_plugin_get_actions_ull(C.ulonglong(fn))
	if r == nil {
		return ""
	}
	defer C.call_plugin_free_ull(C.ulonglong(free), unsafe.Pointer(r))
	return C.GoString(r)
}

func CallPluginAction(fn uintptr, name, argsJSON string, free uintptr) string {
	cname := C.CString(name)
	cargs := C.CString(argsJSON)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cargs))
	r := C.call_plugin_call_action_ull(C.ulonglong(fn), cname, cargs)
	if r == nil {
		return ""
	}
	defer C.call_plugin_free_ull(C.ulonglong(free), unsafe.Pointer(r))
	return C.GoString(r)
}
