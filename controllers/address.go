package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Akin-Kilic/ecommerce/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress(c *fiber.Ctx) error {

	user_id := c.Query("id")
	if user_id == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid User ID"})
	}

	address, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var addresses models.Address
	addresses.Address_id = primitive.NewObjectID()
	if err = c.BodyParser(&addresses); err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(err.Error())
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	match_filter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: address}}}}
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "&address"}}}}
	group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "&address_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}

	pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{match_filter, unwind, group})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	var addressinfo []bson.M
	if err = pointcursor.All(ctx, &addressinfo); err != nil {
		log.Println(err)
	}

	var size int32

	for _, address_no := range addressinfo {
		count := address_no["count"]
		size = count.(int32)
	}

	if size < 2 {
		filter := bson.D{primitive.E{Key: "_id", Value: address}}
		update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
		_, err := UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Println(err)
		}
	} else {
		return c.Status(fiber.StatusMethodNotAllowed).JSON("Not Allowed")
	}
	ctx.Done()
	return c.Status(fiber.StatusCreated).JSON("Created")
}

func EditHomeAddress(c *fiber.Ctx) error {
	user_id := c.Query("id")
	if user_id == "" {
		return c.Status(fiber.StatusNotFound).JSON("Not Found")
	}
	usert_id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	var editaddress models.Address
	if err := c.BodyParser(&editaddress); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
	update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.0.house_name", Value: editaddress.House}, {Key: "address.0.street_name", Value: editaddress.Street}, {Key: "address.0.city_name", Value: editaddress.City}, {Key: "address.0.pin_code", Value: editaddress.Pincode}}}}
	_, err = UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("Something went wrong")
	}
	defer cancel()
	ctx.Done()
	return c.Status(fiber.StatusOK).JSON("Successfully Updated the Home address")
}

func EditWorkAddress(c *fiber.Ctx) error {
	user_id := c.Query("id")
	if user_id == "" {
		return c.Status(fiber.StatusNotFound).JSON("Wrong id not provided")
	}
	usert_id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	var editaddress models.Address
	if err := c.BodyParser(&editaddress); err != nil {
		c.Status(http.StatusBadRequest).JSON(err.Error())
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
	update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.1.house_name", Value: editaddress.House}, {Key: "address.1.street_name", Value: editaddress.Street}, {Key: "address.1.city_name", Value: editaddress.City}, {Key: "address.1.pin_code", Value: editaddress.Pincode}}}}
	_, err = UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON("something Went wrong")
	}
	defer cancel()
	ctx.Done()
	return c.Status(fiber.StatusOK).JSON("Successfully updated the Work Address")

}

func DeleteAddress(c *fiber.Ctx) error {
	user_id := c.Query("id")
	if user_id == "" {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid Search Index")
	}
	addresses := make([]models.Address, 0)
	usert_id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return c.Status(500).JSON("Internal Server Error")
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
	update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
	_, err = UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON("Wromg")
	}
	defer cancel()
	ctx.Done()
	return c.Status(fiber.StatusOK).JSON("Successfully Deleted!")
}
