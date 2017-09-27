package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("%q Request from %q for %q", req.Method, req.RemoteAddr, req.URL)

		h.ServeHTTP(rw, req)
	})
}

func httpServer(addr string, handler http.Handler) func() {
	return func() {
		log.Println("Starting HTTP server...")
		err := http.ListenAndServe(addr, handler)
		if err != nil {
			log.Fatalf("Error: Failed to start server on %q: %v", addr, err)
		}
	}
}

func httpsServer(addr, tlscert, tlskey string, handler http.Handler) func() {
	return func() {
		log.Println("Starting HTTPS server...")
		err := http.ListenAndServeTLS(addr, tlscert, tlskey, handler)
		if err != nil {
			log.Fatalf("Error: Failed to start server on %q: %v", addr, err)
		}
	}
}

type mapFlag map[string]string

func (flag mapFlag) String() string {
	if len(flag) == 0 {
		return ""
	}

	var buf bytes.Buffer

	for k, v := range flag {
		buf.WriteString(k)
		buf.WriteByte(':')
		buf.WriteString(v)
		buf.WriteByte(',')
	}

	buf.Truncate(buf.Len() - 1)

	return buf.String()
}

func (flag *mapFlag) Set(val string) error {
	mappings := strings.Split(val, ",")
	if *flag == nil {
		*flag = make(mapFlag, len(mappings))
	}

	for _, m := range mappings {
		parts := strings.Split(m, "~")
		if len(parts) != 2 {
			return fmt.Errorf("Invalid mapping: %q", m)
		}

		(*flag)[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return nil
}

func main() {
	log.SetOutput(os.Stdout)

	var (
		redirects mapFlag
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [options]\n", os.Args[0])
		fmt.Fprintln(os.Stderr)

		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	root := flag.String("root", "", "The root directory of the server. Defaults to the current directory.")
	addr := flag.String("addr", ":8080", "The address to listen on.")
	tlscert := flag.String("tlscert", "", "TLS Certificate.")
	tlskey := flag.String("tlskey", "", "TLS Key.")
	cache := flag.Duration("cache", 0, "Amount of time to cache files for. 0, the default, disables caching.")
	dirs := flag.Bool("dirs", false, "List directory contents when accessed.")
	config := flag.String("config", "", "The config file to use. If this is specified, all other options are ignored.")
	flag.Var(&redirects, "redirects", "Comma-seperated list of from~to redirect mappings.")
	flag.Parse()

	if *config != "" {
		serve, err := fromConfig(*config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config from %q: %v\n", *config, err)
			os.Exit(1)
		}
		serve()
		return
	}

	var fs http.FileSystem
	fs = http.Dir(*root)
	if *cache > 0 {
		fs = &FSCache{
			FS:      fs,
			Timeout: *cache,
		}
	}
	if !*dirs {
		fs = &FileOnlyFS{FS: fs}
	}

	var handler http.Handler
	handler = http.FileServer(fs)
	handler = redirector{H: handler, R: redirects}
	handler = logger(handler)

	serve := httpServer(*addr, handler)
	if len(*tlscert)+len(*tlskey) > 0 {
		serve = httpsServer(*addr, *tlscert, *tlskey, handler)
	}

	serve()
}
