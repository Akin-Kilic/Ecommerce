package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Akin-Kilic/ecommerce/database"
	"github.com/Akin-Kilic/ecommerce/models"
	generate "github.com/Akin-Kilic/ecommerce/tokens"
	"github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")

var Validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Println(err)
	}
	return string(bytes)
}

func VerifyPassword(userpassword string, givenpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenpassword), []byte(userpassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "Login Or Password is Incorrect"
		valid = false
	}
	return valid, msg
}

func SignUp(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	validationErr := Validate.Struct(user)
	if validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": validationErr})
	}

	count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	if count > 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "User already exists"})
	}
	count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
	defer cancel()
	if err != nil {
		log.Println(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	if count > 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Phone is already in use"})
	}
	password := HashPassword(*user.Password)
	user.Password = &password

	user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.ID = primitive.NewObjectID()
	user.User_ID = user.ID.Hex()
	token, refreshtoken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
	user.Token = &token
	user.Refresh_Token = &refreshtoken
	user.UserCart = make([]models.ProductUser, 0)
	user.Address_Details = make([]models.Address, 0)
	user.Order_Status = make([]models.Order, 0)
	_, inserterr := UserCollection.InsertOne(ctx, user)
	if inserterr != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "not created"})
	}
	defer cancel()
	return c.Status(http.StatusCreated).JSON("Successfully Signed Up!!")

}

func Login(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var user models.User
	var founduser models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err})
	}
	err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
	defer cancel()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "login or password incorrect"})
	}
	PasswordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
	defer cancel()
	if !PasswordIsValid {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}
	token, refreshToken, _ := generate.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, founduser.User_ID)
	defer cancel()
	generate.UpdateAllTokens(token, refreshToken, founduser.User_ID)
	return c.Status(http.StatusFound).JSON(founduser)

}

func ProductViewerAdmin(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var products models.Product
	defer cancel()
	if err := c.BodyParser(&products); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	products.Product_ID = primitive.NewObjectID()
	_, anyerr := ProductCollection.InsertOne(ctx, products)
	if anyerr != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Not Created"})
	}
	defer cancel()
	return c.Status(http.StatusOK).JSON("Successfully added our Product Admin!!")

}

func SearchProduct(c *fiber.Ctx) error {
	var productlist []models.Product
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	cursor, err := ProductCollection.Find(ctx, bson.D{{}})
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON("Something Went Wrong Please Try After Some Time")
		return err
	}
	err = cursor.All(ctx, &productlist)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer cursor.Close(ctx)
	if err := cursor.Err(); err != nil {
		log.Println(err)
		c.JSON(fiber.Map{
			"status": 400,
			"erorr":  "invalid"})
		return err
	}
	defer cancel()
	return c.Status(fiber.StatusOK).JSON(productlist)

}

func SearchProductByQuery(c *fiber.Ctx) error {
	var searchproducts []models.Product
	queryParam := c.Query("name")
	if queryParam == "" {
		log.Println("query is empty")
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"Error": "Invalid Search Index"})
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	searchquerydb, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": queryParam}})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON("something went wrong in fetching the dbquery")
	}
	err = searchquerydb.All(ctx, &searchproducts)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON("invalid")
	}
	defer searchquerydb.Close(ctx)
	if err := searchquerydb.Err(); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON("invalid request")
	}
	defer cancel()
	return c.Status(fiber.StatusOK).JSON(searchproducts)

}
