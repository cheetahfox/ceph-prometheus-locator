// Package health provides health and readiness check handlers for the Ceph Prometheus Locator service.
package health

import (
	"sync"

	"github.com/gofiber/fiber/v2"
)

var (
	ready bool = false
	mu    sync.RWMutex
)

// SetReady sets the ready state with mutex protection
func SetReady(state bool) {
	mu.Lock()
	defer mu.Unlock()
	ready = state
}

// IsReady returns the ready state with mutex protection
func IsReady() bool {
	mu.RLock()
	defer mu.RUnlock()
	return ready
}

func GetHealthz(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

func GetReadyz(c *fiber.Ctx) error {
	if !IsReady() {
		return c.SendStatus(503)
	}
	return c.SendStatus(200)
}
