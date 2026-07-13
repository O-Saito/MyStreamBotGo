package loader

import (
	"MyStreamBot/helpers"
	"path/filepath"
	"strings"
)

type LoadedLibrary struct {
	Path   string
	Handle uintptr
	Funcs  map[string]uintptr
}

func LoadPlugin(path string) (*LoadedLibrary, error) {
	handle, fns, err := openLibrary(path)
	if err != nil {
		return nil, err
	}

	return &LoadedLibrary{Path: path, Handle: handle, Funcs: fns}, nil
}

func LoadDirectory(dir string) ([]*LoadedLibrary, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}

	var result []*LoadedLibrary
	for _, entry := range entries {
		ext := strings.ToLower(filepath.Ext(entry))
		if ext != ".dll" && ext != ".so" && ext != ".dylib" {
			continue
		}

		lib, err := LoadPlugin(entry)
		if err != nil {
			helpers.Logf(helpers.WARN, "[LOADER] skipping %s: %v", entry, err)
			continue
		}
		result = append(result, lib)
	}

	return result, nil
}

func CloseLibrary(lib *LoadedLibrary) {
	if lib != nil && lib.Handle != 0 {
		closeLibrary(lib.Handle)
		lib.Handle = 0
	}
}

type loadError struct {
	path string
	msg  string
}

func (e *loadError) Error() string {
	return e.path + ": " + e.msg
}
