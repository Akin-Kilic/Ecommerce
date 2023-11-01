package routes

import "github.com/gofiber/fiber/v2"

func EndPoints(app *fiber.App) {
	app.Post("/users/signup", SignUp)
	app.Post("/users/login", Login)
	app.Post("/admin/addprpducts", ProductViewerAdmin)
	app.Get("/users/productview", SearchProduct)
	app.Get("/users/search", SearchProductByQuery)
}
