package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	listen = flag.String("listen", ":8080", "listen address")
	dir    = flag.String("dir", ".", "directory to serve")
)

func main() {
	flag.Parse()
	log.Printf("listening on %q...", *listen)
	http.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		goroot := os.Getenv("GOROOT")
		if goroot == "" {
			panic("No $GOROOT set, don't know where to find wasm_exec.js")
		}
		http.ServeFile(w, r, path.Join(goroot, "misc/wasm/wasm_exec.js"))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// if explicit file specified:
		// http.ServeFile(w, r, r.URL.Path[1:]+"index.html")
		w.Write([]byte(index_html()))
	})
	http.HandleFunc("/lib.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	log.Fatal(http.ListenAndServe(*listen, nil))
}
