package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

type FSCache struct {
	once sync.Once

	FS      http.FileSystem
	Timeout time.Duration

	m sync.RWMutex
	c map[string]*cacheEntry
}

func (fs *FSCache) init() {
	fs.once.Do(func() {
		fs.c = make(map[string]*cacheEntry)
	})
}

func (fs *FSCache) Open(name string) (http.File, error) {
	fs.init()

	fs.m.RLock()
	cache, ok := fs.c[name]
	fs.m.RUnlock()

	if !ok || (time.Since(cache.ts) > fs.Timeout) {
		file, err := fs.FS.Open(name)
		if err != nil {
			return nil, err
		}

		fi, err := file.Stat()
		if err != nil {
			return nil, err
		}

		if fi.IsDir() {
			return file, nil
		}

		defer file.Close()

		cache, err = newCacheEntry(file, fi)
		if err != nil {
			return nil, err
		}

		fs.m.Lock()
		if _, ok := fs.c[name]; !ok {
			fs.c[name] = cache
		}
		fs.m.Unlock()
	}

	return cache.file(), nil
}

type cacheEntry struct {
	data []byte
	fi   os.FileInfo
	ts   time.Time
}

func newCacheEntry(r io.Reader, fi os.FileInfo) (*cacheEntry, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &cacheEntry{
		data: data,
		fi:   fi,
		ts:   time.Now(),
	}, nil
}

type cacheFile struct {
	*bytes.Reader
	io.Closer

	fi os.FileInfo
}

func (c *cacheEntry) file() http.File {
	f := &cacheFile{
		Reader: bytes.NewReader(c.data),
		fi:     c.fi,
	}
	f.Closer = ioutil.NopCloser(f)

	return f
}

func (f *cacheFile) Readdir(n int) ([]os.FileInfo, error) {
	// This might not be the right error. In other words, this error
	// might be invalid.
	return nil, os.ErrInvalid
}

func (f *cacheFile) Stat() (os.FileInfo, error) {
	return f.fi, nil
}
