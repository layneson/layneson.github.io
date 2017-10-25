package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"golang.org/x/crypto/acme/autocert"
)

var password string

func main() {
	lfile, err := os.OpenFile("sitestrapper.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	log.SetOutput(lfile)

	if err := loadPassword(); err != nil {
		log.Fatalf("Failed to load password: %v\n", err)
	}

	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	go func() {
		server := &http.Server{}
		mux := &http.ServeMux{}

		mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			http.Redirect(rw, req, "https://"+req.Host+req.URL.Path, 302)
		})

		server.Addr = globalConfig.HTTPAddress
		server.Handler = mux

		log.Fatal(server.ListenAndServe())
	}()

	server := &http.Server{}
	mux := &http.ServeMux{}

	mux.HandleFunc("/", handlePage)

	server.Handler = mux

	hosts := []string{}

	for h := range globalConfig.HostMap {
		hosts = append(hosts, h)
	}

	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("sitestrapper-autocert-cache"),
		HostPolicy: autocert.HostWhitelist(hosts...),
		Email:      globalConfig.Email,
	}

	rproxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf("127.0.0.1:%d", globalConfig.HostMap[req.Host])
		},
	}

	server.Addr = globalConfig.HTTPSAddress
	server.TLSConfig = &tls.Config{GetCertificate: manager.GetCertificate}

	log.Fatal(server.ListenAndServeTLS("", ""))
}

func loadPassword() error {
	conts, err := ioutil.ReadFile("sitestrapper-password.txt")
	if err != nil {
		return err
	}

	password = string(conts)

	return nil
}

type config struct {
	HostMap      map[string]int // maps host (rowsofb.laynegustafson.com) to port (8081)
	Email        string
	HTTPSAddress string
	HTTPAddress  string
}

var globalConfig = &config{}

func loadConfig() error {
	conts, err := ioutil.ReadFile("sitestrapper-config.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(conts, globalConfig)

	return err
}

var rproxy *httputil.ReverseProxy

func handlePage(rw http.ResponseWriter, req *http.Request) {
	log.Println("Handling request")
	if _, ok := globalConfig.HostMap[req.Host]; !ok {
		rw.WriteHeader(404)
		return
	}

	log.Println(req.Host)

	rproxy.ServeHTTP(rw, req)
}
