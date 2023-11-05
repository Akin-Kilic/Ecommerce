package routes

import (
	"github.com/Akin-Kilic/ecommerce/controllers"
	"github.com/gofiber/fiber/v2"
)

func EndPoints(app *fiber.App) {
	app.Post("/users/signup", controllers.SignUp)
	app.Post("/users/login", controllers.Login)
	app.Post("/admin/addprpducts", controllers.ProductViewerAdmin)
	app.Get("/users/productview", controllers.SearchProduct)
	app.Get("/users/search", controllers.SearchProductByQuery)
}
