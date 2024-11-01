// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

const (
	configFilePath            = "config/config.json"
	DefaultIndexerBlockWindow = 1024
)

type Config struct {
	ExternalSubscriberAddr string `json:"external_subscriber_addr"`
	IndexerBlockWindow     int    `json:"indexer_block_window"`
}

func NewDefaultConfig() Config {
	return Config{
		ExternalSubscriberAddr: "",
		IndexerBlockWindow:     DefaultIndexerBlockWindow,
	}
}

func LoadConfig() Config {
	config := NewDefaultConfig()

	// Resolve absolute path for config file
	absPath, err := filepath.Abs(configFilePath)
	if err != nil {
		log.Printf("Error resolving config path: %v", err)
	} else {
		file, err := os.Open(absPath)
		if err == nil {
			defer file.Close()
			if err := json.NewDecoder(file).Decode(&config); err != nil {
				log.Printf("Error decoding config file, using defaults: %v", err)
			}
		} else {
			log.Printf("Config file not found at %s, using defaults.", absPath)
		}
	}

	// Environment variable overrides
	if envAddress := os.Getenv("NUKLAIVM_EXTERNAL_SUBSCRIBER_SERVER_ADDRESS"); envAddress != "" {
		config.ExternalSubscriberAddr = envAddress
		log.Printf("External subscriber address set from environment variable: %s", envAddress)
	}
	if envBlockWindow := os.Getenv("NUKLAIVM_INDEXER_BLOCK_WINDOW"); envBlockWindow != "" {
		if val, err := strconv.Atoi(envBlockWindow); err == nil {
			config.IndexerBlockWindow = val
			log.Printf("Indexer block window set from environment variable: %d", val)
		}
	}

	log.Printf("External Subscriber address: %s", config.ExternalSubscriberAddr)
	log.Printf("Indexer block window: %d", config.IndexerBlockWindow)

	return config
}
