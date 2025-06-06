package locator

import (
	"log"

	"github.com/cheetahfox/ceph-prometheus-locator/config"
)

type Host struct {
	Targets string `json:"targets"`
	Lables  Lables `json:"labels"`
}

type Lables struct {
	Instance string `json:"instance"`
}

func GetActiveHost() (string, bool, error) {
	// This function should contain the logic to determine the host URL of the Ceph managed Prometheus server.
	// For now, it returns a placeholder URL.
	Urls := config.Urls
	if len(Urls) == 0 {
		// If no URLs are configured, return an error or a default value.
		return "", false, nil
	}
	activeHost := "http://example-ceph-prometheus-server:9090"
	if len(Urls) > 0 {
		// If there are hosts configured, return the first one.
		activeHost = string(Urls[0].HostUrl)
	}
	return activeHost, true, nil
}

func StartLocator() error {
	// This function should contain the logic to start the locator service.
	// For now, it does nothing.

	for _, url := range config.Urls {
		if config.Debug {
			log.Printf("Starting check for %s\n", url.HostName)
		}
		go checkHost(url.HostUrl, url.HostName)
	}

	return nil
}

func checkHost(hostUrl string, hostName string) {
	if config.Debug {
		log.Printf("registered host %s at %s\n", hostName, hostUrl)
	}
	// Here you would implement the actual logic to check the host.
}
