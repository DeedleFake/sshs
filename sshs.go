package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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

func main() {
	log.SetOutput(os.Stdout)

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
	flag.Parse()

	var fs http.FileSystem
	fs = http.Dir(*root)
	if !*dirs {
		fs = &FileOnlyFS{FS: fs}
	}
	if *cache > 0 {
		fs = &FSCache{
			FS:      fs,
			Timeout: *cache,
		}
	}

	var handler http.Handler
	handler = http.FileServer(fs)
	handler = logger(handler)

	serve := httpServer(*addr, handler)
	if len(*tlscert)+len(*tlskey) > 0 {
		serve = httpsServer(*addr, *tlscert, *tlskey, handler)
	}

	serve()
}
