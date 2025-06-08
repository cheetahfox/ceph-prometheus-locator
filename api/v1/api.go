package v1

import (
	"log"

	"github.com/cheetahfox/ceph-prometheus-locator/cephlocator"
	"github.com/cheetahfox/ceph-prometheus-locator/config"
	"github.com/gofiber/fiber/v2"
)

func GetLocation(c *fiber.Ctx) error {
	// This is a placeholder for the actual logic to find the ceph managed Prometheus server.
	// In a real implementation, you would query the Ceph cluster or configuration to find the
	// appropriate server and return its address.

	serverUrl, _, err := getHostUrl()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve Ceph managed Prometheus server URL",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ceph managed Prometheus server location",
		"server":  serverUrl,
	})
}

func getHostUrl() (string, bool, error) {
	// This function should contain the logic to determine the host URL of the Ceph managed Prometheus server.
	// For now, it returns a placeholder URL.
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
