package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

var Urls = []Config{}
var Debug, Profile bool
var RefreshInterval int = 60

type Config struct {
	HostUrl  string `json:"url"`
	HostName string `json:"hostname"`
}

func init() {
	if err := readHosts(); err != nil {
		panic("Failed to read hosts: " + err.Error())
	}
	if os.Getenv("DEBUG") == "true" {
		Debug = true
		log.Println("Debug mode is enabled")
	}
	if os.Getenv("PROFILE") == "true" {
		Profile = true
		log.Println("Profiling mode is enabled")
	}
	if interval := os.Getenv("REFRESH_INTERVAL"); interval != "" {
		if parsedInterval, err := strconv.Atoi(interval); err == nil {
			RefreshInterval = parsedInterval
			log.Printf("Refresh interval set to %d seconds\n", RefreshInterval)
		} else {
			log.Printf("Invalid REFRESH_INTERVAL value: %s, using default %d seconds\n", interval, RefreshInterval)
		}
	}
}

// readHosts reads the configuration file and populates the Urls slice with the host configurations.
func readHosts() error {
	var configFile string
	if os.Getenv("LOCATOR_CONFIG") != "" {
		// If the environment variable is set, use it.
		configFile = os.Getenv("LOCATOR_CONFIG")
	} else {
		configFile = "config.json"
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// If the config file does not exist, return an error.
		return err
	}

	configData, err := os.ReadFile(configFile)
	if err != nil { // If there is an error reading the file, return it.
		log.Println("Error reading config file:", err)
		return err
	}

	err = json.Unmarshal(configData, &Urls)
	if err != nil {
		log.Println("Error unmarshalling config data:", err)
		return err
	}

	if Debug {
		for _, url := range Urls {
			log.Printf("Loaded URL: %s, Hostname: %s\n", url.HostUrl, url.HostName)
		}
	}

	return nil
}
