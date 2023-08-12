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
	http.HandleFunc("/empty", truncateEmptyFile)

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
	GB = 1024 * MB
)

func strFileSize(sizeInByte uint64) string {
	switch {
	case sizeInByte >= GB:
		return fmt.Sprintf("%.2fGB", float64(sizeInByte)/float64(GB))
	case sizeInByte >= MB:
		return fmt.Sprintf("%.2fMB", float64(sizeInByte)/float64(MB))
	case sizeInByte >= KB:
		return fmt.Sprintf("%.2fKB", float64(sizeInByte)/float64(KB))
	default:
		return fmt.Sprintf("%dB", sizeInByte)
	}
}

func diskusage(w http.ResponseWriter, r *http.Request) {
	usage := du.NewDiskUsage("/tmp")

	m := make(map[string]any)
	m["size"] = strFileSize(usage.Size())
	m["used"] = strFileSize(usage.Used())
	m["available"] = strFileSize(usage.Available())
	m["free"] = strFileSize(usage.Free())

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(m); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to call Statfs at path: %v", err)
		return
	}
}

func truncateEmptyFile(w http.ResponseWriter, r *http.Request) {
	sizeInByte := int64(1 * MB)
	f, err := os.Create("/tmp/empty")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to create empty file: %v", err)
		return
	}
	if err := f.Truncate(sizeInByte); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to truncate empty file: %v", err)
		return
	}
	_, _ = fmt.Fprintf(w, "truncated empty file: size=%s", strFileSize(uint64(sizeInByte)))
}
