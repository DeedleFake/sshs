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

func main() {
	log.SetOutput(os.Stdout)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [options]\n", os.Args[0])
		fmt.Fprintln(os.Stderr)

		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	root := flag.String("root", "", "The root directory of the server.")
	addr := flag.String("addr", ":8080", "The address to listen on.")
	flag.Parse()

	http.Handle("/", logger(http.FileServer(http.Dir(*root))))

	log.Println("Starting server...")
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalf("Error: Failed to start server on %q: %v", *addr, err)
	}
}
