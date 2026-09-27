// Package v1 implements the API version 1 handlers for the Ceph Prometheus Locator service.
package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cheetahfox/ceph-prometheus-locator/cephlocator"
	"github.com/cheetahfox/ceph-prometheus-locator/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/gofiber/fiber/v2"
)

var (
	sdClient = &http.Client{Timeout: 30 * time.Second}

	apiRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_requests_total",
		Help: "Total number of API requests",
	}, []string{"method", "endpoint", "status"})
)

// GetLocation proxies the active Ceph managed Prometheus server's service-discovery response.
// It rewrites advertised target domains when HOST_DOMAIN is configured.
func GetLocation(c *fiber.Ctx) error {
	hostURL, running, err := getHostUrl()
	if err != nil {
		apiRequestsTotal.WithLabelValues(c.Method(), "/sd/prometheus/sd-config", "500").Inc()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve Ceph managed Prometheus server URL",
		})
	}

	if !running {
		apiRequestsTotal.WithLabelValues(c.Method(), "/sd/prometheus/sd-config", "404").Inc()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No active Ceph managed Prometheus server found",
		})
	}

	upstreamURL, err := url.Parse("http://" + hostURL)
	if err != nil {
		log.Printf("Failed to parse active Ceph Prometheus URL %q: %v", hostURL, err)
		apiRequestsTotal.WithLabelValues(c.Method(), "/sd/prometheus/sd-config", "502").Inc()
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to contact Ceph managed Prometheus server",
		})
	}
	upstreamURL.RawQuery = string(c.Context().URI().QueryString())

	req, err := http.NewRequestWithContext(c.Context(), c.Method(), upstreamURL.String(), nil)
	if err != nil {
		log.Printf("Failed to create request for Ceph Prometheus URL %q: %v", upstreamURL, err)
		apiRequestsTotal.WithLabelValues(c.Method(), "/sd/prometheus/sd-config", "502").Inc()
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to contact Ceph managed Prometheus server",
		})
	}
	resp, err := sdClient.Do(req)
	if err != nil {
		log.Printf("Failed to fetch Ceph service-discovery response from %q: %v", upstreamURL, err)
		apiRequestsTotal.WithLabelValues(c.Method(), "/sd/prometheus/sd-config", "502").Inc()
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to contact Ceph managed Prometheus server",
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read Ceph service-discovery response from %q: %v", upstreamURL, err)
		apiRequestsTotal.WithLabelValues(c.Method(), "/sd/prometheus/sd-config", "502").Inc()
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to read Ceph managed Prometheus response",
		})
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		body, err = rewriteTargetDomains(body, config.HostDomain)
		if err != nil {
			log.Printf("Failed to rewrite Ceph service-discovery targets: %v", err)
			apiRequestsTotal.WithLabelValues(c.Method(), "/sd/prometheus/sd-config", "502").Inc()
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "Invalid service-discovery response from Ceph managed Prometheus server",
			})
		}
	}

	apiRequestsTotal.WithLabelValues(c.Method(), "/sd/prometheus/sd-config", strconv.Itoa(resp.StatusCode)).Inc()
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		c.Set("Content-Type", contentType)
	}
	return c.Status(resp.StatusCode).Send(body)
}

func GetActiveHost(c *fiber.Ctx) error {
	// This endpoint returns the active host URL without fetching its service-discovery response.
	url, running, err := getHostUrl()
	if err != nil {
		apiRequestsTotal.WithLabelValues(c.Method(), "/api/v1/status", "500").Inc()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve Ceph managed Prometheus server URL",
		})
	}

	if !running {
		apiRequestsTotal.WithLabelValues(c.Method(), "/api/v1/status", "404").Inc()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No active Ceph managed Prometheus server found",
		})
	}

	apiRequestsTotal.WithLabelValues(c.Method(), "/api/v1/status", "200").Inc()
	return c.JSON(fiber.Map{
		"url": url,
	})
}

func getHostUrl() (string, bool, error) {
	activeHostUrl, running, err := cephlocator.GetActiveHost()
	if err != nil {
		// If there is an error retrieving the active host, return an error.
		return "", false, err
	}
	if !running {
		// We couldn't find an active host, return an empty string and false.
		if config.Debug {
			// If debug mode is enabled, log the error.
			log.Println("No active Ceph managed Prometheus server found.")
		}
		return "", false, nil
	}

	return activeHostUrl, true, nil
}

func rewriteTargetDomains(body []byte, domain string) ([]byte, error) {
	if domain == "" {
		return body, nil
	}

	domain = strings.TrimSuffix(domain, ".")
	if !validDomain(domain) {
		return nil, fmt.Errorf("invalid HOST_DOMAIN %q", domain)
	}

	var groups []map[string]json.RawMessage
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("failed to decode target groups: %w", err)
	}

	for _, group := range groups {
		rawTargets, exists := group["targets"]
		if !exists {
			continue
		}
		var targets []string
		if err := json.Unmarshal(rawTargets, &targets); err != nil {
			return nil, fmt.Errorf("failed to decode target list: %w", err)
		}
		for i, target := range targets {
			rewritten, err := applyHostDomain(target, domain)
			if err != nil {
				return nil, fmt.Errorf("failed to rewrite target %q: %w", target, err)
			}
			targets[i] = rewritten
		}
		encodedTargets, err := json.Marshal(targets)
		if err != nil {
			return nil, fmt.Errorf("failed to encode rewritten targets: %w", err)
		}
		group["targets"] = encodedTargets
	}

	rewrittenBody, err := json.Marshal(groups)
	if err != nil {
		return nil, fmt.Errorf("failed to encode target groups: %w", err)
	}
	return rewrittenBody, nil
}

func applyHostDomain(target, domain string) (string, error) {
	u, err := url.Parse("//" + target)
	if err != nil {
		return "", fmt.Errorf("failed to parse target: %w", err)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return "", fmt.Errorf("target has no hostname")
	}
	if net.ParseIP(hostname) == nil {
		firstLabel := strings.SplitN(hostname, ".", 2)[0]
		if firstLabel == "" {
			return "", fmt.Errorf("target %q has an invalid hostname", target)
		}
		hostname = firstLabel + "." + domain
	}

	if port := u.Port(); port != "" {
		u.Host = net.JoinHostPort(hostname, port)
	} else {
		u.Host = hostname
	}
	return u.Host + u.EscapedPath(), nil
}

func validDomain(domain string) bool {
	if domain == "" || len(domain) > 253 {
		return false
	}
	for _, label := range strings.Split(domain, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' {
				return false
			}
		}
	}
	return true
}
