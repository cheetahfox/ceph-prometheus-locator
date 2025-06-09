package v1

import (
	"log"

	"github.com/cheetahfox/ceph-prometheus-locator/cephlocator"
	"github.com/cheetahfox/ceph-prometheus-locator/config"
	"github.com/gofiber/fiber/v2"
)

// GetLocation handles the request to redirect to the active Ceph managed Prometheus server
// It retrieves the active host URL and appends any query parameters from the request.
func GetLocation(c *fiber.Ctx) error {
	hostUrl, _, err := getHostUrl()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve Ceph managed Prometheus server URL",
		})
	}

	qparms := c.Queries()
	if len(qparms) > 0 {
		// If there are query parameters, append them to the host URL.
		hostUrl += "?"
		for key, value := range qparms {
			hostUrl += key + "=" + value + "&"
		}
	}

	return c.Redirect(hostUrl, fiber.StatusFound)
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
