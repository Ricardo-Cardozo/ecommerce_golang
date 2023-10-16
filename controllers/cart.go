package controllers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Ricardo-Cardozo/ecommerce_golang/database"
	"github.com/Ricardo-Cardozo/ecommerce_golang/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection *mongo.Collection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")

		if productQueryID == "" {
			log.Println("product id is empty")
			c.JSON(http.StatusNotFound, gin.H{
				"error":     "product id is empty",
				"productId": productQueryID,
			})
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))

			return
		}

		userQueryID := c.Query("userID")

		if userQueryID == "" {
			log.Println("user id is empty")
			c.JSON(http.StatusNotFound, gin.H{
				"error":     "user id is empty",
				"productId": userQueryID,
			})
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.AddProductToCart(
			ctx,
			app.prodCollection,
			app.userCollection,
			productID,
			userQueryID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.IndentedJSON(200, "Succesfully added to the cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")

		if productQueryID == "" {
			log.Println("product id is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))

			return
		}

		userQueryID := c.Query("userID")

		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.RemoveCartItem(
			ctx,
			app.prodCollection,
			app.userCollection,
			productID,
			userQueryID,
		)

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		c.IndentedJSON(200, "Succesfully remove item to the cart")
	}
}

func GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		if user_id == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Invalid Search Index",
			})
			return
		}

		usert_id, err := primitive.ObjectIDFromHex(user_id)

		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			100*time.Second,
		)

		defer cancel()

		var filledcart models.User

		err = UserCollection.FindOne(
			ctx,
			bson.D{
				primitive.E{
					Key:   "_id",
					Value: usert_id,
				},
			},
		).Decode(&filledcart)

		if err != nil {
			log.Println(err)
			c.IndentedJSON(404, "not found")
			return
		}

		pipeline := mongo.Pipeline{
			{{
				Key: "$match",
				Value: bson.M{
					"_id": usert_id,
				},
			}},
			{{
				Key: "$unwind",
				Value: bson.M{
					"path": "$usercart",
				},
			}},
			{{
				Key: "$group",
				Value: bson.M{
					"_id": "$_id",
					"total": bson.M{
						"$sum": "$usercart.price",
					},
				}},
			},
		}

		pointcursor, err := UserCollection.Aggregate(ctx, pipeline)

		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregation failed"})
			return
		}

		var listing []bson.M

		if err = pointcursor.All(ctx, &listing); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode results"})
			return
		}

		responseData := make([]gin.H, 0, len(listing))

		for _, json := range listing {
			responseData = append(responseData, gin.H{"total": json["total"], "cart": filledcart.UserCart})
		}

		c.JSON(200, responseData)
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")

		if userQueryID == "" {
			log.Panicln("user id empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserId is empty"))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		err := database.BuyItemFromCart(
			ctx,
			app.userCollection,
			userQueryID,
		)

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		c.IndentedJSON(200, "Succesfully placed the order")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")

		if productQueryID == "" {
			log.Println("product id is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))

			return
		}

		userQueryID := c.Query("userID")

		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.InstantBuyer(
			ctx,
			app.prodCollection,
			app.userCollection,
			productID,
			userQueryID,
		)

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		c.IndentedJSON(200, "Succesfully placed the order")
	}
}
