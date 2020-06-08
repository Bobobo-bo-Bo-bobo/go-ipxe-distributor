package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	//    log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

func handleHTTP(cfg *Configuration) error {
	parsed, err := url.Parse(cfg.Global.URL)
	if err != nil {
		return err
	}

	if parsed.Scheme != "http" {
		return fmt.Errorf("Invalid or unsupported scheme %s", parsed.Scheme)
	}

	prefix := strings.TrimRight(strings.TrimLeft(parsed.Path, "/"), "/")

	router := mux.NewRouter()
	router.HandleFunc(prefix+defaultPath, defaultHandler)
	router.HandleFunc(prefix+defaultPath+"/", defaultHandler)
	router.HandleFunc(prefix+groupPath, groupHandler)
	router.HandleFunc(prefix+macPath, macHandler)
	router.HandleFunc(prefix+serialPath, serialHandler)

	server := &http.Server{
		Handler:      router,
		Addr:         parsed.Host,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	server.ListenAndServe()
	return nil
}

func defaultHandler(writer http.ResponseWriter, request *http.Request) {
}

func groupHandler(writer http.ResponseWriter, request *http.Request) {
}

func macHandler(writer http.ResponseWriter, request *http.Request) {
}

func serialHandler(writer http.ResponseWriter, request *http.Request) {
}
