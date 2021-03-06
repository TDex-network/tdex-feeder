// Copyright (c) 2020 The VulpemVentures developers

// Feeder allows to connect an external price feed to the TDex Daemon to determine the current market price.
package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-feeder/config"
	"github.com/tdex-network/tdex-feeder/internal/adapters"
	"github.com/tdex-network/tdex-feeder/internal/application"
)

func main() {
	// Interrupt Notification.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	// retrieve feeder service from config file
	feeder := configFileToFeederService(config.GetConfigPath())

	log.Info("Start the feeder...")
	go func() {
		err := feeder.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// check for interupt
	<-interrupt
	log.Info("Shutting down the feeder...")
	err := feeder.Stop()
	log.Info("Feeder service stopped")
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func configFileToFeederService(configFilePath string) application.FeederService {
	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	configBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	config := &adapters.Config{}
	err = json.Unmarshal(configBytes, config)
	if err != nil {
		log.Fatal(err)
	}

	feeder, err := config.ToFeederService()
	if err != nil {
		log.Fatal(err)
	}

	return feeder
}
