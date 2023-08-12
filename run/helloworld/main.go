package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ricochet2200/go-disk-usage/du"
)

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)
	http.HandleFunc("/du", diskusage)
	http.HandleFunc("/empty", emptyFile)
	http.HandleFunc("/dummy", dummyFile)

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

func emptyFile(w http.ResponseWriter, r *http.Request) {
	sizeInByte := int64(300 * MB)
	var msg string

	if fileExists("/tmp/empty") {
		msg += "before: file exits\n"
	} else {
		msg += "before: file not exists\n"
	}

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
	msg += fmt.Sprintf("truncated empty file: size=%s\n", strFileSize(uint64(sizeInByte)))

	if fileExists("/tmp/empty") {
		msg += "after: file exits\n"
	} else {
		msg += "after: file not exists\n"
	}

	_, _ = fmt.Fprintf(w, msg)
}

func dummyFile(w http.ResponseWriter, r *http.Request) {
	const sizeLimit = 512
	const dir = "/tmp/dummy"

	sizeInMB := getSize(r)
	if sizeInMB > sizeLimit {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "dummy size limit(%dMB) exceeded: %dMB", sizeLimit, sizeInMB)
		return
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		log.Fatal(err)
	}
	if err := createDummyFile(dir, sizeInMB); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to create dummy file: %v", err)
		return
	}

	s, err := statDummyFiles(dir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to count dummy file: %v", err)
		return
	}

	_, _ = fmt.Fprintf(w, "dummy files count: %d, total size: %s\n", s.count, strFileSize(uint64(s.totalSize)))
}

func getSize(r *http.Request) int {
	const defaultSize = 5

	strSize := r.URL.Query().Get("size")
	s, err := strconv.Atoi(strSize)
	if err != nil {
		return defaultSize
	}
	return s
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

const dummyFilePrefix = "dummy"

func createDummyFile(dir string, sizeInMB int) error {
	f, err := os.CreateTemp(dir, dummyFilePrefix)
	if err != nil {
		return err
	}

	b := make([]byte, 1*KB)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	until := sizeInMB * MB / len(b)
	for i := 0; i < until; i++ {
		if _, err := f.Write(b); err != nil {
			return err
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	return nil
}

type dummyFilesStat struct {
	count     int
	totalSize int64
}

func statDummyFiles(dir string) (*dummyFilesStat, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	count := 0
	totalSize := int64(0)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), dummyFilePrefix) {
			count++
			totalSize += info.Size()
		}
	}

	return &dummyFilesStat{
		count:     count,
		totalSize: totalSize,
	}, nil
}
