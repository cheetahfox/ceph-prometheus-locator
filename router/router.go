package router

import (
	v1 "github.com/cheetahfox/ceph-prometheus-locator/api/v1"
	"github.com/cheetahfox/ceph-prometheus-locator/health"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Singularity Running running") // send text
	})

	app.Get("/healthz", health.GetHealthz)
	app.Get("/readyz", health.GetReadyz)

	// API routes
	api := app.Group("/api/v1/", logger.New())
	api.Get("/", v1.GetLocation)
}
