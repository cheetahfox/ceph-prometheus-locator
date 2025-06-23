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
	"os"
	"os/signal"
	"syscall"

	"github.com/cheetahfox/ceph-prometheus-locator/cephlocator"
	"github.com/cheetahfox/ceph-prometheus-locator/health"
	"github.com/cheetahfox/ceph-prometheus-locator/router"

	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Println("Starting ceph prometheus locator...")
	locator := fiber.New()

	// Setup routes
	router.SetupRoutes(locator)

	// Start the server.
	go func() {
		if err := locator.Listen(":8080"); err != nil {
			panic(err)
		}
	}()

	// Start the locator service.
	err := cephlocator.StartLocator()
	if err != nil {
		fmt.Println("Failed to start locator service:", err)
		panic(err)
	}

	// Set the service as ready.
	health.Ready = true

	// Listen for Sigint or SigTerm and exit if you get them.
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	<-done
	fmt.Println("Shutdown Started")
}
