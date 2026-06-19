//go:build windows

package loader

/*
#include <stdlib.h>
#include "hostapi.h"
*/
import "C"
import (
	"MyStreamBot/helpers"
	"syscall"
)

func openLibrary(path string) (uintptr, map[string]uintptr, error) {
	handle, err := syscall.LoadLibrary(path)
	if err != nil {
		return 0, nil, &loadError{path, "LoadLibrary: " + err.Error()}
	}

	required := []string{
		"plugin_name", "plugin_api_version", "plugin_set_host", "plugin_free",
	}
	optional := []string{
		"plugin_init", "plugin_start", "plugin_stop",
		"plugin_on_chat", "plugin_on_event",
		"plugin_get_actions", "plugin_call_action",
	}

	fns := make(map[string]uintptr, len(required)+len(optional))

	for _, name := range required {
		addr, err := syscall.GetProcAddress(handle, name)
		if err != nil {
			syscall.FreeLibrary(handle)
			return 0, nil, &loadError{path, "missing required export: " + name}
		}
		fns[name] = addr
	}

	for _, name := range optional {
		addr, err := syscall.GetProcAddress(handle, name)
		if err == nil {
			fns[name] = addr
		}
	}

	helpers.Logf(helpers.DEBUG, "[LOADER] opened %s (handle=%v)", path, handle)
	return uintptr(handle), fns, nil
}

func closeLibrary(handle uintptr) {
	if handle != 0 {
		syscall.FreeLibrary(syscall.Handle(handle))
	}
}
