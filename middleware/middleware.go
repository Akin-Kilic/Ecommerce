package middleware

import (
	"net/http"

	token "github.com/Akin-Kilic/ecommerce/tokens"
	"github.com/gofiber/fiber/v2"
)

func Authentication(c *fiber.Ctx) error {
	ClientToken := c.Get("token")
	if ClientToken == "" {
		return c.Status(http.StatusInternalServerError).JSON("No Authorization Header Provided")
	}
	claims, err := token.ValidateToken(ClientToken)
	if err != "" {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}
	c.Set("email", claims.Email)
	c.Set("uid", claims.Uid)
	return c.Next()
}
