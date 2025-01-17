package middlewares

import (
	"github.com/gofiber/fiber/v2"
)

func parseToken(c *fiber.Ctx) error {
	// Token'ı al
	// token := c.Get("Authorization")
	// if token == "" {
	// 	return httputils.PrepareUnauthorizedRequestError("Token is required")
	// }

	// // Token'ı parse et
	// claims, err := jwtutils.ParseClaims(token)
	// if err != nil {
	// 	return httputils.PrepareUnauthorizedRequestError("Invalid token")
	// }

	// // Token'ı context'e ekle
	// c.Locals("user", claims)
	return c.Next()
}
