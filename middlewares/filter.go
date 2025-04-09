package middlewares

import "github.com/gofiber/fiber/v2"

func filter(c *fiber.Ctx) error {
	offset := c.QueryInt("offset")
	limit := c.QueryInt("limit")
	query := c.Query("query")

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
	c.Locals("query", query)

	return c.Next()
}
