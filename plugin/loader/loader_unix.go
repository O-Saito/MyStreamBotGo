//go:build linux || darwin

package loader

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
*/
import "C"
import (
	"MyStreamBot/helpers"
	"unsafe"
)

func openLibrary(path string) (uintptr, map[string]uintptr, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	handle := C.dlopen(cpath, C.RTLD_NOW)
	if handle == nil {
		errStr := C.GoString(C.dlerror())
		return 0, nil, &loadError{path, "dlopen: " + errStr}
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
		cname := C.CString(name)
		addr := C.dlsym(handle, cname)
		C.free(unsafe.Pointer(cname))
		if addr == nil {
			C.dlclose(handle)
			return 0, nil, &loadError{path, "missing required export: " + name}
		}
		fns[name] = uintptr(addr)
	}

	for _, name := range optional {
		cname := C.CString(name)
		addr := C.dlsym(handle, cname)
		C.free(unsafe.Pointer(cname))
		if addr != nil {
			fns[name] = uintptr(addr)
		}
	}

	helpers.Logf(helpers.DEBUG, "[LOADER] opened %s (handle=%p)", path, handle)
	return uintptr(handle), fns, nil
}

//go:nocheckptr
func closeLibrary(handle uintptr) {
	if handle != 0 {
		C.dlclose(unsafe.Pointer(handle))
	}
}
