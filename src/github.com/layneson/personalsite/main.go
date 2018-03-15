package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	lfile, err := os.OpenFile("personalsite.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(lfile)

	log.Println("Starting...")

	go func() {
		server := &http.Server{}
		mux := &http.ServeMux{}

		mux.HandleFunc("/", handleRowsOfB)

		server.Handler = mux
		server.Addr = ":4041"

		server.ListenAndServe()
	}()

	http.HandleFunc("/", handlePage)

	http.ListenAndServe(":4040", nil)
}

func serveDirectory(dir string, path string, rw http.ResponseWriter) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	log.Printf("Serving directory file %s/%s\n", dir, path)

	if strings.TrimSpace(path) == "" {
		path = "index.html"
	}

	splitPath := strings.Split(path, "/")

	for _, part := range splitPath {
		if strings.TrimSpace(part) == ".." {
			rw.WriteHeader(404)
			return
		}
	}

	splitPath = append([]string{dir}, splitPath...)

	fpath := filepath.Join(splitPath...)

	buffer := make([]byte, 1024*1024)

	file, err := os.Open(fpath)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(404)
		return
	}

	var contentType string
	if strings.HasSuffix(fpath, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(fpath, ".html") {
		contentType = "text/html"
	} else {
		initialBuffer := make([]byte, 512)
		file.Read(initialBuffer)
		contentType = http.DetectContentType(initialBuffer)
		file.Seek(0, 0)
	}

	rw.Header().Set("Content-Type", contentType)

	io.CopyBuffer(rw, file, buffer)

	file.Close()
}

func handlePage(rw http.ResponseWriter, req *http.Request) {
	serveDirectory("./site", req.URL.Path, rw)
}

func handleRowsOfB(rw http.ResponseWriter, req *http.Request) {
	serveDirectory("rowsofb", req.URL.Path, rw)
}
