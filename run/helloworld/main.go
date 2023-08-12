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
	http.HandleFunc("/dummy", statsDummyFile)
	http.HandleFunc("/dummy:add", addDummyFile)

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

const dummyDir = "/tmp/dummy"

func addDummyFile(w http.ResponseWriter, r *http.Request) {
	const sizeLimitInMB = 512

	sizeInMB := getSize(r)
	if sizeInMB > sizeLimitInMB {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "dummy size limit(%dMB) exceeded: %dMB", sizeLimitInMB, sizeInMB)
		return
	}

	if err := os.MkdirAll(dummyDir, 0750); err != nil {
		log.Fatal(err)
	}
	if err := createDummyFile(dummyDir, sizeInMB); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to create dummy file: %v", err)
		return
	}

	s, err := statDummyFiles(dummyDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to get stats: %v", err)
		return
	}

	_, _ = fmt.Fprintf(w, "dummy files: %v\n", s)
}

func statsDummyFile(w http.ResponseWriter, r *http.Request) {
	s, err := statDummyFiles(dummyDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "failed to get stats: %v", err)
		return
	}

	_, _ = fmt.Fprintf(w, "dummy files: %v\n", s)
}

func getSize(r *http.Request) int {
	const fallbackSize = 1

	strSize := r.URL.Query().Get("size")
	s, err := strconv.Atoi(strSize)
	if err != nil {
		return fallbackSize
	}
	return s
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func dirExists(dirname string) bool {
	s, err := os.Stat(dirname)
	return err == nil && s.IsDir()
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

func (s *dummyFilesStat) String() string {
	return fmt.Sprintf("count: %d, total size: %s", s.count, strFileSize(uint64(s.totalSize)))
}

func statDummyFiles(dir string) (*dummyFilesStat, error) {
	if !dirExists(dir) {
		return &dummyFilesStat{
			count:     0,
			totalSize: 0,
		}, nil
	}

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
