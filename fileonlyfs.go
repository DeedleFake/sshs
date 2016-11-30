package main

import (
	"net/http"

	"golang.org/x/sys/unix"
)

// FileOnlyFS is an http.FileSystem that wraps another
// http.FileSystem, restricting access to directories. Attempting to
// access directories through it results in a 403 error.
type FileOnlyFS struct {
	// The wrapped http.FileSystem.
	FS http.FileSystem
}

func (fs FileOnlyFS) Open(name string) (http.File, error) {
	file, err := fs.FS.Open(name)
	if err != nil {
		return file, err
	}

	fi, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	if fi.IsDir() {
		file.Close()
		return nil, unix.EPERM
	}

	return file, nil
}
