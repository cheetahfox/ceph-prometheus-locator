package v1

import (
	"log"

	"github.com/cheetahfox/ceph-prometheus-locator/cephlocator"
	"github.com/cheetahfox/ceph-prometheus-locator/config"
	"github.com/gofiber/fiber/v2"
)

// GetLocation handles the request to redirect to the active Ceph managed Prometheus server
func GetLocation(c *fiber.Ctx) error {
	serverUrl, _, err := getHostUrl()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve Ceph managed Prometheus server URL",
		})
	}

	return c.Redirect(serverUrl, fiber.StatusFound)
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
