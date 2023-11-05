package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Akin-Kilic/ecommerce/database"
	"github.com/Akin-Kilic/ecommerce/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart(c *fiber.Ctx) error {
	productQueryID := c.Query("id")
	if productQueryID == "" {
		log.Println("product id is empty")
		return c.Status(fiber.StatusBadRequest).JSON("product id is empty")
	}
	userQueryID := c.Query("userID")
	if userQueryID == "" {
		log.Println("user id is empty")
		return c.Status(fiber.StatusBadRequest).JSON("user id is empty")
	}
	productID, err := primitive.ObjectIDFromHex(productQueryID)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = database.AddProductToCart(ctx, app.prodCollection, app.userCollection, productID, userQueryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON("Successfully Added to the cart")

}

func (app *Application) RemoveItem(c *fiber.Ctx) error {
	productQueryID := c.Query("id")
	if productQueryID == "" {
		log.Println("product id is inavalid")
		return c.Status(fiber.StatusBadRequest).JSON("product id is empty")
	}

	userQueryID := c.Query("userID")
	if userQueryID == "" {
		log.Println("user id is empty")
		return c.Status(fiber.StatusBadRequest).JSON("UserID is empty")
	}

	ProductID, err := primitive.ObjectIDFromHex(productQueryID)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = database.RemoveCartItem(ctx, app.prodCollection, app.userCollection, ProductID, userQueryID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}
	return c.Status(200).JSON("Successfully removed from cart")

}

func GetItemFromCart(c *fiber.Ctx) error {
	user_id := c.Query("id")
	if user_id == "" {
		return c.Status(fiber.StatusNotFound).JSON("invalid id")
	}

	usert_id, _ := primitive.ObjectIDFromHex(user_id)

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var filledcart models.User
	err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: usert_id}}).Decode(&filledcart)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON("not id found")
	}

	filter_match := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: usert_id}}}}
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
	pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind, grouping})
	if err != nil {
		log.Println(err)
	}
	var listing []bson.M
	if err = pointcursor.All(ctx, &listing); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}
	for _, json := range listing {
		c.Status(fiber.StatusOK).JSON(json["total"])
		return c.Status(fiber.StatusOK).JSON(filledcart.UserCart)
	}
	ctx.Done()
	return c.Status(fiber.StatusOK).JSON("success")
}

func (app *Application) BuyFromCart(c *fiber.Ctx) error {
	userQueryID := c.Query("id")
	if userQueryID == "" {
		log.Panicln("user id is empty")
		return c.Status(fiber.StatusBadRequest).JSON("UserID is empty")
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err := database.BuyItemFromCart(ctx, app.userCollection, userQueryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON("Successfully Placed the order")

}

func (app *Application) InstantBuy(c *fiber.Ctx) error {
	UserQueryID := c.Query("userid")
	if UserQueryID == "" {
		log.Println("UserID is empty")
		return c.Status(fiber.StatusBadRequest).JSON("UserID is empty")
	}
	ProductQueryID := c.Query("pid")
	if ProductQueryID == "" {
		log.Println("Product_ID id is empty")
		return c.Status(fiber.StatusBadRequest).JSON("product_id is empty")
	}
	productID, err := primitive.ObjectIDFromHex(ProductQueryID)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = database.InstantBuyer(ctx, app.prodCollection, app.userCollection, productID, UserQueryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON("Successully placed the order")
}
