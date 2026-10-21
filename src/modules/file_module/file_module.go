package file_module

import (
	"joodpakpaw/pkg/storage"
)

type module struct {
	fileSystem *storage.FileSystem
}

func NewModule() module {
	return module{
		fileSystem: storage.CurrentFileStorage(),
	}
}
