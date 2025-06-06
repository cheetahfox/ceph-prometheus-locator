/*
Simple Go Program to locate which server ceph is running it's managed Prometheus.

This might only be useful for me; since I want to pull metrics from a ceph managed
Prometheus server and I don't want to make any modifications to the ceph cluster. I
just want to use what they are already running. This issue is that ceph can run this
on any of the nodes in the cluster. If something changes in the cluster it can move
to a new node without warning and it will break my monitoring setup.

Could I solve this differently? Yeah, I do different things like scraping the Prometheus
node exporters directly, but it's nice that ceph manges everything so I just need to
pull the metrics from one place.
*/
package main

import (
	"fmt"

	"github.com/cheetahfox/ceph-prometheus-locator/health"
	"github.com/cheetahfox/ceph-prometheus-locator/router"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("Starting ceph prometheus locator...")
	locator := fiber.New()

	prometheus := fiberprometheus.New("ceph_prometheus_locator")
	prometheus.RegisterAt(locator, "/metrics")
	locator.Use(prometheus.Middleware)

	health.Ready = true

	// Setup routes
	router.SetupRoutes(locator)

}
