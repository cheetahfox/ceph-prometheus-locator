package config

import (
	"encoding/json"
	"log"
	"os"
)

var Urls = []Config{}

type Config struct {
	HostUrl string `json:"hosts"`
}

func init() {
	if err := readHosts(); err != nil {
		panic("Failed to read hosts: " + err.Error())
	}
}

func readHosts() error {
	// This function should contain the logic to read the hosts from a configuration file or environment variables.
	// For now, it returns a placeholder list of hosts.
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

	return nil
}
