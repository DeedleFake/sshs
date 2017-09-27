package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DeedleFake/wdte"
	"github.com/DeedleFake/wdte/std"
)

func configImporter(from string) (*wdte.Module, error) {
	switch from {
	case "simple":
		return &wdte.Module{
			Funcs: map[wdte.ID]wdte.Func{
				"server": wdte.GoFunc(func(frame wdte.Frame, args ...wdte.Func) wdte.Func {
					frame = frame.WithID("server")

					return Config{
						addr:      ":8080",
						redirects: make(map[string]string),
					}
				}),

				"root":     wdte.GoFunc(simpleRoot),
				"addr":     wdte.GoFunc(simpleAddr),
				"tls":      wdte.GoFunc(simpleTLS),
				"cache":    wdte.GoFunc(simpleCache),
				"dirs":     wdte.GoFunc(simpleDirs),
				"redirect": wdte.GoFunc(simpleRedirect),
			},
		}, nil

	default:
		return std.Import(from)
	}
}

func fromConfig(config string) (func(), error) {
	file, err := os.Open(config)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	m, err := std.Module().Parse(file, wdte.ImportFunc(configImporter))
	if err != nil {
		return nil, err
	}

	build, ok := m.Funcs["build"]
	if !ok {
		return nil, errors.New("No build function in config.")
	}
	server := build.Call(wdte.F())
	switch server := server.(type) {
	case wdteServer:
		return server.server(), nil
	case error:
		return nil, server
	}

	return nil, fmt.Errorf("Unexpected return type from build: %T", server)
}

type Config struct {
	root            string
	addr            string
	tlscert, tlskey string
	cache           time.Duration
	dirs            bool
	redirects       map[string]string
}

func (c Config) server() func() {
	var fs http.FileSystem
	fs = http.Dir(c.root)
	if c.cache > 0 {
		fs = &FSCache{
			FS:      fs,
			Timeout: c.cache,
		}
	}
	if !c.dirs {
		fs = &FileOnlyFS{FS: fs}
	}

	var handler http.Handler
	handler = http.FileServer(fs)
	handler = redirector{H: handler, R: c.redirects}
	handler = logger(handler)

	serve := httpServer(c.addr, handler)
	if len(c.tlscert)+len(c.tlskey) > 0 {
		serve = httpsServer(c.addr, c.tlscert, c.tlskey, handler)
	}

	return serve
}

func (c Config) Call(frame wdte.Frame, args ...wdte.Func) wdte.Func {
	return c
}

type wdteServer interface {
	server() func()
}

func simpleRoot(frame wdte.Frame, args ...wdte.Func) (next wdte.Func) {
	frame = frame.WithID("root")

	if len(args) == 0 {
		return wdte.GoFunc(simpleRoot)
	}

	root := args[0].Call(frame).(wdte.String)
	return wdte.GoFunc(func(frame wdte.Frame, args ...wdte.Func) wdte.Func {
		if len(args) == 0 {
			return next
		}

		c := args[0].Call(frame).(Config)
		c.root = string(root)
		return c
	})
}

func simpleAddr(frame wdte.Frame, args ...wdte.Func) (next wdte.Func) {
	frame = frame.WithID("addr")

	if len(args) == 0 {
		return wdte.GoFunc(simpleAddr)
	}

	addr := args[0].Call(frame).(wdte.String)
	return wdte.GoFunc(func(frame wdte.Frame, args ...wdte.Func) wdte.Func {
		if len(args) == 0 {
			return next
		}

		c := args[0].Call(frame).(Config)
		c.addr = string(addr)
		return c
	})
}

func simpleTLS(frame wdte.Frame, args ...wdte.Func) (next wdte.Func) {
	frame = frame.WithID("tls")

	switch len(args) {
	case 0:
		return wdte.GoFunc(simpleTLS)
	case 1:
		return wdte.GoFunc(func(frame wdte.Frame, innerargs ...wdte.Func) wdte.Func {
			return simpleTLS(frame, append(args, innerargs...)...)
		})
	}

	tlscert := args[0].Call(frame).(wdte.String)
	tlskey := args[1].Call(frame).(wdte.String)
	return wdte.GoFunc(func(frame wdte.Frame, args ...wdte.Func) wdte.Func {
		if len(args) == 0 {
			return next
		}

		c := args[0].Call(frame).(Config)
		c.tlscert = string(tlscert)
		c.tlskey = string(tlskey)
		return c
	})
}

func simpleCache(frame wdte.Frame, args ...wdte.Func) (next wdte.Func) {
	frame = frame.WithID("cache")

	if len(args) == 0 {
		return wdte.GoFunc(simpleCache)
	}

	cache, err := time.ParseDuration(string(args[0].Call(frame).(wdte.String)))
	if err != nil {
		return wdte.Error{Err: err, Frame: frame}
	}

	return wdte.GoFunc(func(frame wdte.Frame, args ...wdte.Func) wdte.Func {
		if len(args) == 0 {
			return next
		}

		c := args[0].Call(frame).(Config)
		c.cache = cache
		return c
	})
}

func simpleDirs(frame wdte.Frame, args ...wdte.Func) (next wdte.Func) {
	frame = frame.WithID("dirs")

	if len(args) == 0 {
		return wdte.GoFunc(simpleDirs)
	}

	dirs := strings.ToLower(string(args[0].Call(frame).(wdte.String)))
	return wdte.GoFunc(func(frame wdte.Frame, args ...wdte.Func) wdte.Func {
		if len(args) == 0 {
			return next
		}

		c := args[0].Call(frame).(Config)
		c.dirs = dirs == "true"
		return c
	})
}

func simpleRedirect(frame wdte.Frame, args ...wdte.Func) (next wdte.Func) {
	frame = frame.WithID("redirect")

	switch len(args) {
	case 0:
		return wdte.GoFunc(simpleRedirect)
	case 1:
		return wdte.GoFunc(func(frame wdte.Frame, innerargs ...wdte.Func) wdte.Func {
			return simpleRedirect(frame, append(args, innerargs...)...)
		})
	}

	from := args[0].Call(frame).(wdte.String)
	to := args[1].Call(frame).(wdte.String)
	return wdte.GoFunc(func(frame wdte.Frame, args ...wdte.Func) wdte.Func {
		if len(args) == 0 {
			return next
		}

		c := args[0].Call(frame).(Config)
		c.redirects[string(from)] = string(to)
		return c
	})
}
