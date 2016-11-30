package main

import (
	"net/http"

	"golang.org/x/sys/unix"
)

type FileOnlyFS struct {
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
