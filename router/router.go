// Package router sets up the HTTP routes for the Ceph Prometheus Locator service.
package router

import (
	v1 "github.com/cheetahfox/ceph-prometheus-locator/api/v1"
	"github.com/cheetahfox/ceph-prometheus-locator/config"
	"github.com/cheetahfox/ceph-prometheus-locator/health"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ceph Locator Running") // Let us know this is running.
	})

	if config.Profile {
		app.Use(pprof.New())
	}

	app.Get("/healthz", health.GetHealthz)
	app.Get("/readyz", health.GetReadyz)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler())) // Prometheus metrics endpoint

	// API routes
	api := app.Group("/api/v1/", logger.New())
	api.Get("/", v1.GetActiveHost)
	api.Get("status", v1.GetActiveHost)

	// SD Path
	sdPath := app.Group("/sd/prometheus/", logger.New())
	sdPath.Get("sd-config", v1.GetLocation)
}
