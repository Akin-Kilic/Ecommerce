package main

import (
	"github.com/Akin-Kilic/ecommerce/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	routes.EndPoints(app)
	app.Listen(":8080")
}
