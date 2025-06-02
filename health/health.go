package health

import (
	"github.com/gofiber/fiber/v2"
)

var Ready bool = false

func GetHealthz(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

func GetReadyz(c *fiber.Ctx) error {
	if !Ready {
		return c.SendStatus(503)
	}
	return c.SendStatus(200)
}
