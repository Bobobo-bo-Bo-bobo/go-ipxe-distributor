package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func handleHTTP(cfg Configuration) error {
	var w time.Duration

	parsed, err := url.Parse(cfg.Global.URL)
	if err != nil {
		return err
	}

	if parsed.Scheme != "http" {
		return fmt.Errorf("Invalid or unsupported scheme %s", parsed.Scheme)
	}

	prefix := strings.TrimRight(strings.TrimLeft(parsed.Path, "/"), "/")
	if len(prefix) != 0 {
		if prefix[0] != '/' {
			prefix = "/" + prefix
		}
	}

	router := mux.NewRouter()
	router.HandleFunc(prefix+defaultPath, defaultHandler)
	router.HandleFunc(prefix+groupPath, groupHandler)
	router.HandleFunc(prefix+macPath, macHandler)
	router.HandleFunc(prefix+serialPath, serialHandler)

	server := &http.Server{
		Handler:      router,
		Addr:         parsed.Host,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.WithFields(log.Fields{
		"default": prefix + defaultPath,
		"group":   prefix + groupPath,
		"mac":     prefix + macPath,
		"serial":  prefix + serialPath,
		"address": server.Addr,
	}).Info("Setting up HTTP handlers and starting web server")

	// start a separate thread so we can listen for signals
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.WithFields(log.Fields{
				"address": server.Addr,
				"error":   err.Error(),
			}).Fatal("Can't start web server")
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, os.Kill)
	<-sigchan

	ctx, cancel := context.WithTimeout(context.Background(), w)
	defer cancel()
	server.Shutdown(ctx)

	return nil
}

func logRequest(request *http.Request) {
	log.WithFields(log.Fields{
		"method":   request.Method,
		"protocol": request.Proto,
		"address":  request.RemoteAddr,
		"host":     request.Host,
		"url":      request.URL.String(),
	}).Info("Request received")
}

func setHeader(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "text/plain")
	writer.Header().Add("Expires", "Thu, 01 Jan 1970 12:00:00 AM GMT")
	writer.Header().Add("X-Clacks-Overhead", "GNU Terry Pratchett")
	writer.Header().Add("X-Content-Type-Options", "nosniff")
}

func defaultHandler(writer http.ResponseWriter, request *http.Request) {
	logRequest(request)

	setHeader(writer)
	writer.WriteHeader(http.StatusOK)

	fmt.Fprintf(writer, "#!ipxe\n")
	fmt.Fprintf(writer, strings.Join(config.Default.IPXEPrepend, "\n"))
	fmt.Fprintf(writer, strings.Join(config.Default.DefaultImage, "\n"))
	fmt.Fprintf(writer, strings.Join(config.Default.IPXEAppend, "\n"))
}

func getiPXEFromLabel(l string) (string, error) {
	var ipxe string

	// Get image label for node
	node, found := config.Nodes[l]
	if !found {
		return "", fmt.Errorf("No node data found for label")
	}

	if node.Image == "" {
		return "", fmt.Errorf("Node for label contains no image name")
	}

	// Get image data for image label
	image, found := config.Images[node.Image]
	if !found {
		// special case: "default" will load the default iPXE data
		if node.Image != "default" {
			return "", fmt.Errorf("No image data found for image name")
		}
	}

	ipxe += fmt.Sprintf("#!ipxe\n")
	ipxe += fmt.Sprintf(strings.Join(config.Default.IPXEPrepend, "\n"))

	if node.Image == "default" {
		ipxe += fmt.Sprintf(strings.Join(config.Default.DefaultImage, "\n"))
	} else {
		ipxe += fmt.Sprintf(strings.Join(image.Action, "\n"))
	}

	ipxe += fmt.Sprintf(strings.Join(config.Default.IPXEAppend, "\n"))

	return ipxe, nil
}

func groupHandler(writer http.ResponseWriter, request *http.Request) {
	logRequest(request)

	setHeader(writer)

	mvars := mux.Vars(request)
	rgrp := mvars["group"]

	// Map provided group to a node label
	nlabel, found := config.GroupNodeMap[rgrp]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"group": rgrp,
		}).Error("Can't find node label for group")

		fmt.Fprintf(writer, "Can't find node label for group %s", rgrp)
		return
	}

	data, err := getiPXEFromLabel(nlabel)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"group": rgrp,
			"label": nlabel,
		}).Error(err.Error())

		return
	}

	writer.WriteHeader(http.StatusOK)

	fmt.Fprintf(writer, data)
}

func macHandler(writer http.ResponseWriter, request *http.Request) {
	logRequest(request)

	setHeader(writer)

	mvars := mux.Vars(request)
	_rmac := mvars["mac"]

	// normalize MAC
	rmac := strings.ToLower(strings.Replace(strings.Replace(_rmac, ":", "", -1), "-", "", -1))

	// Map provided group to a node label
	nlabel, found := config.MACNodeMap[rmac]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"mac":            _rmac,
			"normalised_mac": rmac,
		}).Error("Can't find node label for MAC")

		fmt.Fprintf(writer, "Can't find node label for MAC %s", _rmac)
		return
	}

	data, err := getiPXEFromLabel(nlabel)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"mac":            _rmac,
			"normalised mac": rmac,
			"label":          nlabel,
		}).Error(err.Error())

		return
	}

	writer.WriteHeader(http.StatusOK)

	fmt.Fprintf(writer, data)
}

func serialHandler(writer http.ResponseWriter, request *http.Request) {
	logRequest(request)

	setHeader(writer)

	mvars := mux.Vars(request)
	rsrl := mvars["serial"]

	// Map provided serial number to a node label
	nlabel, found := config.SerialNodeMap[rsrl]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"serial": rsrl,
		}).Error("Can't find node label for serial number")

		fmt.Fprintf(writer, "Can't find node label for serial number %s", rsrl)
		return
	}

	data, err := getiPXEFromLabel(nlabel)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"serial_number": rsrl,
			"label":         nlabel,
		}).Error(err.Error())

		return
	}

	writer.WriteHeader(http.StatusOK)

	fmt.Fprintf(writer, data)
}
