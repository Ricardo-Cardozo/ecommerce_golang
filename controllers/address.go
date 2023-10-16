package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Ricardo-Cardozo/ecommerce_golang/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		if user_id == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Invalid Search index",
			})
			return
		}

		address, err := primitive.ObjectIDFromHex(user_id)

		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		var addresses models.Address

		addresses.Address_ID = primitive.NewObjectID()

		if err = c.BindJSON(&addresses); err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		pipeline := mongo.Pipeline{
			{{
				Key: "$match",
				Value: bson.M{
					"_id": address,
				},
			}},
			{{
				Key: "$unwind",
				Value: bson.M{
					"path": "$address",
				},
			}},
			{{
				Key: "$group",
				Value: bson.M{
					"_id": "$address_id",
					"count": bson.M{
						"$sum": 1,
					},
				}},
			},
		}

		pointcursor, err := UserCollection.Aggregate(ctx, pipeline)

		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
			return
		}

		var addressinfo []bson.M

		if err = pointcursor.All(ctx, &addressinfo); err != nil {
			panic(err)
		}

		var size int32

		for _, address_no := range addressinfo {
			count := address_no["count"]
			size = count.(int32)
		}

		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: address}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: address}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)

			if err != nil {
				fmt.Println(err)
			}

		} else {
			c.IndentedJSON(400, "not allowed")
		}
	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		if user_id == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Invalid",
			})
			return
		}

		var editaddress models.Address

		if err := c.BindJSON(&editaddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}

		usert_id, err := primitive.ObjectIDFromHex(user_id)

		if err != nil {
			c.IndentedJSON(500, "InternalServer Error")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		filter := bson.M{"_id": usert_id}

		update := bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "address.0.house_name", Value: editaddress.House},
					{Key: "address.0.street_name", Value: editaddress.Street},
					{Key: "address.0.city_name", Value: editaddress.City},
					{Key: "address.0.pin_code", Value: editaddress.Pincode},
				},
			},
		}

		if _, err := UserCollection.UpdateOne(ctx, filter, update); err != nil {
			c.IndentedJSON(500, "Something went wrong")
			return
		}

		c.IndentedJSON(200, "Succesfuly update address")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		if user_id == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Invalid",
			})
			return
		}

		var editaddress models.Address

		if err := c.BindJSON(&editaddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}

		usert_id, err := primitive.ObjectIDFromHex(user_id)

		if err != nil {
			c.IndentedJSON(500, "InternalServer Error")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		filter := bson.M{"_id": usert_id}

		update := bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "address.1.house_name", Value: editaddress.House},
					{Key: "address.1.street_name", Value: editaddress.Street},
					{Key: "address.1.city_name", Value: editaddress.City},
					{Key: "address.1.pin_code", Value: editaddress.Pincode},
				},
			},
		}

		if _, err := UserCollection.UpdateOne(ctx, filter, update); err != nil {
			c.IndentedJSON(500, "Something went wrong")
			return
		}

		c.IndentedJSON(200, "Succesfuly update address")
	}
}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Invalid Search index",
			})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)

		usert_id, err := primitive.ObjectIDFromHex(user_id)

		if err != nil {
			c.IndentedJSON(500, "internal Server Error")
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			100*time.Second,
		)

		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}

		update := bson.D{
			{
				Key: "$set",
				Value: bson.D{
					primitive.E{
						Key:   "address",
						Value: addresses,
					},
				},
			},
		}

		_, err = UserCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.IndentedJSON(404, "wrong command")
			return
		}

		c.IndentedJSON(200, "Succesfully Deleted")
	}
}
