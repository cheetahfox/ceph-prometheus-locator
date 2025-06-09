package router

import (
	v1 "github.com/cheetahfox/ceph-prometheus-locator/api/v1"
	"github.com/cheetahfox/ceph-prometheus-locator/health"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ceph Locator Running") // Let us know this is running.
	})

	app.Get("/healthz", health.GetHealthz)
	app.Get("/readyz", health.GetReadyz)

	// API routes
	api := app.Group("/api/v1/", logger.New())
	api.Get("/", v1.GetLocation)

	// SD Path
	sdPath := app.Group("/sd/prometheus/", logger.New())
	sdPath.Get("sd-config", v1.GetLocation)
}
