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

	// Get image label for node
	node, found := config.Nodes[nlabel]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"group": rgrp,
			"label": nlabel,
		}).Error("Group maps to node label but no configuration found for this label")

		fmt.Fprintf(writer, "Group %s maps to node label %s but no configuration found for this label", rgrp, nlabel)
		return
	}

	if node.Image == "" {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"group": rgrp,
			"label": nlabel,
		}).Error("Group maps to node label but no configuration found for this label")

		fmt.Fprintf(writer, "Group %s maps to node label %s, but node label contains no image name", rgrp, nlabel)
		return
	}

	// Get image data for image label
	image, found := config.Images[node.Image]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"group": rgrp,
			"label": nlabel,
		}).Error("Group maps to node label which references image label, but no such image label exist")

		fmt.Fprintf(writer, "Group %s maps to node label %s which references image label %s, but no such image label exist", rgrp, nlabel, node.Image)
		return
	}

	writer.WriteHeader(http.StatusOK)

	fmt.Fprintf(writer, "#!ipxe\n")
	fmt.Fprintf(writer, strings.Join(config.Default.IPXEPrepend, "\n"))
	fmt.Fprintf(writer, strings.Join(image.Action, "\n"))
	fmt.Fprintf(writer, strings.Join(config.Default.IPXEAppend, "\n"))
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

	// Get image label for node
	node, found := config.Nodes[nlabel]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"mac":            _rmac,
			"normalised_mac": rmac,
			"label":          nlabel,
		}).Error("MAC maps to node label but no configuration found for this label")

		fmt.Fprintf(writer, "MAC %s maps to node label %s but no configuration found for this label", _rmac, nlabel)
		return
	}

	if node.Image == "" {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"mac":            _rmac,
			"normalised_mac": rmac,
			"label":          nlabel,
		}).Error("MAC maps to node label but no configuration found for this label")

		fmt.Fprintf(writer, "MAC %s maps to node label %s, but node label contains no image name", _rmac, nlabel)
		return
	}

	// Get image data for image label
	image, found := config.Images[node.Image]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"mac":            _rmac,
			"normalised_mac": rmac,
			"label":          nlabel,
			"image":          node.Image,
		}).Error("MAC maps to node label which references image label, but no such image label exist")

		fmt.Fprintf(writer, "MAC %s maps to node label %s which references image label %s, but no such image label exist", _rmac, nlabel, node.Image)
		return
	}

	writer.WriteHeader(http.StatusOK)

	fmt.Fprintf(writer, "#!ipxe\n")
	fmt.Fprintf(writer, strings.Join(config.Default.IPXEPrepend, "\n"))
	fmt.Fprintf(writer, strings.Join(image.Action, "\n"))
	fmt.Fprintf(writer, strings.Join(config.Default.IPXEAppend, "\n"))
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

	// Get image label for node
	node, found := config.Nodes[nlabel]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"serial": rsrl,
			"label":  nlabel,
		}).Error("Serial number maps to node label but no configuration found for this label")

		fmt.Fprintf(writer, "Serial number %s maps to node label %s but no configuration found for this label", rsrl, nlabel)
		return
	}

	if node.Image == "" {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"serial": rsrl,
			"label":  nlabel,
		}).Error("Serial number maps to node label but no configuration found for this label")

		fmt.Fprintf(writer, "Serial number %s maps to node label %s, but node label contains no image name", rsrl, nlabel)
		return
	}

	// Get image data for image label
	image, found := config.Images[node.Image]
	if !found {
		writer.WriteHeader(http.StatusNotFound)

		log.WithFields(log.Fields{
			"serial": rsrl,
			"label":  nlabel,
		}).Error("Serial number maps to node label which references image label, but no such image label exist")

		fmt.Fprintf(writer, "Serial number %s maps to node label %s which references image label %s, but no such image label exist", rsrl, nlabel, node.Image)
		return
	}

	writer.WriteHeader(http.StatusOK)

	fmt.Fprintf(writer, "#!ipxe\n")
	fmt.Fprintf(writer, strings.Join(config.Default.IPXEPrepend, "\n"))
	fmt.Fprintf(writer, strings.Join(image.Action, "\n"))
	fmt.Fprintf(writer, strings.Join(config.Default.IPXEAppend, "\n"))
}
