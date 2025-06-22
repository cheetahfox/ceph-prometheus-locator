package cephlocator

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/cheetahfox/ceph-prometheus-locator/config"
)

var Hosts map[string]*Host

type Host struct {
	HostUrl  string
	HostName string
	Active   bool
	Shutdown chan bool
}

func init() {
	Hosts = make(map[string]*Host)
}

// GetActiveHost returns the URL of the first active host if available.
// If no hosts are configured, it returns an empty string and false
func GetActiveHost() (string, bool, error) {
	var activeHostUrl, activeHost string
	var foundActive bool

	if len(Hosts) == 0 {
		err := fmt.Errorf("No active Ceph managed Prometheus server found. Please configure hosts in the config file.")
		return "", false, err
	}

	for _, host := range Hosts {
		if host.Active {
			activeHost = host.HostUrl
			foundActive = true
			break // We only need the first active host, should only be one in any case.
		}
	}

	if foundActive {
		hostUrl, err := stripHttpParam(activeHost)
		if err != nil {
			log.Printf("Error stripping HTTP parameters from host URL %s: %v\n", activeHost, err)
			return "", false, err
		}
		activeHostUrl = hostUrl
	}

	return activeHostUrl, foundActive, nil
}

func StartLocator() error {
	for _, url := range config.Urls {
		if config.Debug {
			log.Printf("Starting check for %s\n", url.HostName)
		}
		go setupHost(url.HostUrl, url.HostName)
	}

	return nil
}

// setupHost initializes a host with its URL and name, and calls checkHosts to monitor it.
// It also sets up a shutdown channel to gracefully stop monitoring when needed. It blocks
// until something writes to the shutdown channel.
func setupHost(hostUrl string, hostName string) {
	if config.Debug {
		log.Printf("registered host %s at %s\n", hostName, hostUrl)
	}

	Hosts[hostName] = &Host{
		HostUrl:  hostUrl,
		HostName: hostName,
		Active:   false,
		Shutdown: make(chan bool, 1),
	}

	go checkHosts(hostName)

	<-Hosts[hostName].Shutdown
	if config.Debug {
		log.Printf("Shutting down host %s at %s\n", hostName, hostUrl)
	}
	delete(Hosts, hostName)
}

// checkHosts periodically checks if the host is still active and exists in the Hosts map.
// Then it will check if the hosts is serving up the prometheus metrics endpoint. If it is
// we set the host to active.
func checkHosts(hostName string) {
	if config.Debug {
		log.Println("Starting check for host:", hostName)
	}

	connection := &http.Client{
		Timeout: time.Duration(30) * time.Second,
		Transport: &http.Transport{
			// Use a custom transport to disable keep-alive connections.
			DisableKeepAlives:   true,
			MaxIdleConnsPerHost: 1,
		},
	}

	ticker := time.NewTicker(time.Duration(config.RefreshInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// Check that the host still exists so we should monitor it.
		if _, exists := Hosts[hostName]; !exists {
			if config.Debug {
				log.Printf("Host %s no longer exists, stopping check.\n", hostName)
			}
			return
		}

		resp, err := connection.Get(Hosts[hostName].HostUrl)
		if err != nil {
			if config.Debug {
				log.Printf("Error checking host %s: %v\n", hostName, err)
			}
			Hosts[hostName].Active = false
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if config.Debug {
				log.Printf("Host %s is active at %s\n", hostName, Hosts[hostName].HostUrl)
			}
			Hosts[hostName].Active = true
		} else {
			if config.Debug {
				log.Printf("Host %s is not active at %s, status code: %d\n", hostName, Hosts[hostName].HostUrl, resp.StatusCode)
			}
			Hosts[hostName].Active = false
		}
		// Close this directly here since normally we don't exit the loop
		resp.Body.Close()
	}
}

// stripHttpParam removes any http parameters from the host URL.
// My configfile will have URL with a query parameter so we need to strip it
// and best to do it here so we can use the host URL in the redirect and append
// the query parameters that the incoming request has.
func stripHttpParam(baseUrl string) (string, error) {
	if baseUrl == "" {
		return "", errors.New("base URL is empty")
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %s: %w", baseUrl, err)
	}

	hostUrl := u.Host + u.Path

	return hostUrl, nil
}
