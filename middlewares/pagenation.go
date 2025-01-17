package middlewares

import "github.com/gofiber/fiber/v2"

func pagination(c *fiber.Ctx) error {
	offset := c.QueryInt("offset")
	limit := c.QueryInt("limit")
	if offset < 0 {
		offset = 0
	}

	if limit == 0 || limit < 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	c.Locals("offset", offset)
	c.Locals("limit", limit)

	return c.Next()
}
