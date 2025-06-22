package v1

import (
	"log"

	"github.com/cheetahfox/ceph-prometheus-locator/cephlocator"
	"github.com/cheetahfox/ceph-prometheus-locator/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/gofiber/fiber/v2"
)

var (
	apiRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_requests_total",
		Help: "Total number of API requests",
	}, []string{"method", "endpoint", "status"})
)

// GetLocation handles the request to redirect to the active Ceph managed Prometheus server
// It retrieves the active host URL and appends any query parameters from the request.
func GetLocation(c *fiber.Ctx) error {
	var header string = "http://"

	url, _, err := getHostUrl()
	if err != nil {
		apiRequestsTotal.WithLabelValues(c.Method(), c.Path(), "500").Inc()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve Ceph managed Prometheus server URL",
		})
	}

	hostUrl := header + url
	qparms := c.Queries()
	if len(qparms) > 0 {
		// If there are query parameters, append them to the host URL.
		hostUrl += "?"
		for key, value := range qparms {
			hostUrl += key + "=" + value + "&"
		}
	}

	if config.Debug {
		log.Printf("Redirecting to active Ceph managed Prometheus server: %s\n", hostUrl)
	}

	apiRequestsTotal.WithLabelValues(c.Method(), c.Path(), "302").Inc()
	return c.Redirect(hostUrl, fiber.StatusFound)
}

func GetActiveHost(c *fiber.Ctx) error {
	// This function is similar to GetLocation but returns the active host URL without redirecting.
	url, running, err := getHostUrl()
	if err != nil {
		apiRequestsTotal.WithLabelValues(c.Method(), c.Path(), "500").Inc()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve Ceph managed Prometheus server URL",
		})
	}

	if !running {
		apiRequestsTotal.WithLabelValues(c.Method(), c.Path(), "404").Inc()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No active Ceph managed Prometheus server found",
		})
	}

	apiRequestsTotal.WithLabelValues(c.Method(), c.Path(), "200").Inc()
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
