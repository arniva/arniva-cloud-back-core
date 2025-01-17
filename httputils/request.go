package httputils

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func BodyParser[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data T

		if err := c.BodyParser(&data); err != nil {
			log.Print(err)
			return PrepareParseError(err.Error())
		}
		c.Locals("data", &data)

		return c.Next()
	}
}
