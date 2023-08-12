package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ricochet2200/go-disk-usage/du"
)

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)
	http.HandleFunc("/du", diskusage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defauting to port %s", port)
	}

	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "World"
	}
	_, _ = fmt.Fprintf(w, "Hello %s!\n", name)
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
)

func diskusage(w http.ResponseWriter, r *http.Request) {
	usage := du.NewDiskUsage("/tmp")

	m := make(map[string]any)
	setSize := func(key string, sizeInByte uint64) {
		m[key] = float64(sizeInByte) / float64(MB)
	}
	setSize("available", usage.Available())
	setSize("free", usage.Free())
	setSize("size", usage.Size())
	setSize("used", usage.Used())
	setSize("usage", uint64(usage.Usage()))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(m); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to call Statfs at path: %v", err)
		return
	}
}
