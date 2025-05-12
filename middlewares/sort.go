package middlewares

import "github.com/gofiber/fiber/v2"

func sort(c *fiber.Ctx) error {
	sort := c.Query("sort")

	c.Locals("sort", sort)

	return c.Next()
}
