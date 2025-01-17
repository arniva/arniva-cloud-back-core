package middlewares

import "github.com/gofiber/fiber/v2"

func SetupMiddlewares(app *fiber.App) {
	// Middleware'leri ekle

	app.Use(parseToken)
	app.Use(databaseSelect)
	app.Use(pagination)
}
