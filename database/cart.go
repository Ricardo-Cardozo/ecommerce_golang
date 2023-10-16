package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Ricardo-Cardozo/ecommerce_golang/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("can't find the product")
	ErrCantDecodeProducts = errors.New("cant find the product")
	ErrUserIdIsNotValid   = errors.New("this user is not valid")
	ErrCantUpdateUser     = errors.New("cannot add this product to the cart")
	ErrCantRemoveItemCart = errors.New("cannot remove this add from the cart")
	ErrCantGetItem        = errors.New("was unable to get the item from the cart")
	ErrCantBuyCartItem    = errors.New("cannot update the purchase")
)

func AddProductToCart(
	ctx context.Context,
	prodCollection *mongo.Collection,
	userCollection *mongo.Collection,
	productID primitive.ObjectID,
	userID string,
) error {
	fmt.Println("Product Collection:", prodCollection.Name())
	fmt.Println("User Collection:", userCollection.Name())

	searchfromdb, err := prodCollection.Find(ctx, bson.M{"_id": productID})

	if err != nil {
		log.Println(err.Error(), searchfromdb)
		return ErrCantFindProduct
	}

	var productcart []models.ProductUser

	err = searchfromdb.All(ctx, &productcart)

	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}

	id, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	filter := bson.D{{Key: "_id", Value: id}}

	update := bson.D{{Key: "$push", Value: bson.D{{Key: "usercart", Value: bson.D{{
		Key: "$each", Value: productcart,
	}}}}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Println(err)
		return ErrCantUpdateUser
	}

	return nil

}

func RemoveCartItem(
	ctx context.Context,
	prodCollection *mongo.Collection,
	userCollection *mongo.Collection,
	productID primitive.ObjectID,
	userID string,
) error {
	id, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}

	_, err = userCollection.UpdateMany(ctx, filter, update)

	if err != nil {
		return ErrCantRemoveItemCart
	}

	return nil
}

func BuyItemFromCart(
	ctx context.Context,
	userCollection *mongo.Collection,
	userID string,
) error {
	id, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	var getcartitems models.User
	var ordercart models.Order

	ordercart.Order_ID = primitive.NewObjectID()
	ordercart.Ordered_At = time.Now()
	ordercart.Order_Cart = make([]models.ProductUser, 0)
	ordercart.Payment_Method.COD = true

	unwind := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$usercart"}}}}
	grouping := bson.D{
		{
			Key: "$group",
			Value: bson.D{{
				Key:   "_id",
				Value: "$usercart",
			},
			},
		},
		{
			Key: "total",
			Value: bson.D{{
				Key:   "$sum",
				Value: "$usercart.price",
			}},
		},
	}

	currentresult, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})

	if err != nil {
		log.Panic(err)
	}

	var getusercart []bson.M

	if err = currentresult.All(ctx, &getusercart); err != nil {
		log.Panic(err)
	}

	var total_price int32

	for _, user_item := range getusercart {
		price := user_item["total"]
		total_price = price.(int32)
	}

	ordercart.Price = int(total_price)

	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{{Key: "orders", Value: ordercart}}}}

	_, err = userCollection.UpdateMany(ctx, filter, update)

	if err != nil {
		log.Println(err)
	}

	err = userCollection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&getcartitems)

	if err != nil {
		log.Println(err)
	}

	filter2 := bson.D{{Key: "_id", Value: id}}
	update2 := bson.M{
		"$push": bson.M{
			"orders.$[].order_list": bson.M{
				"$each": getcartitems.UserCart,
			},
		},
	}

	_, err = userCollection.UpdateOne(ctx, filter2, update2)

	if err != nil {
		log.Println(err)
	}

	usercart_empty := make([]models.ProductUser, 0)
	filter3 := bson.D{{Key: "_id", Value: id}}
	update3 := bson.D{{
		Key: "$set", Value: bson.D{{
			Key: "usercart", Value: usercart_empty,
		}},
	}}

	_, err = userCollection.UpdateOne(ctx, filter3, update3)

	if err != nil {
		return ErrCantBuyCartItem
	}

	return nil
}

func InstantBuyer(
	ctx context.Context,
	prodCollection *mongo.Collection,
	userCollection *mongo.Collection,
	productID primitive.ObjectID,
	userID string,
) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	var product_details models.ProductUser
	var orders_details models.Order

	orders_details.Order_ID = primitive.NewObjectID()
	orders_details.Ordered_At = time.Now()
	orders_details.Order_Cart = make([]models.ProductUser, 0)
	orders_details.Payment_Method.COD = true

	err = prodCollection.FindOne(ctx, bson.D{{Key: "_id", Value: productID}}).Decode(&product_details)

	if err != nil {
		log.Println(err)
	}

	orders_details.Price = product_details.Price

	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{{Key: "orders", Value: orders_details}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Println(err)
	}

	filter2 := bson.D{{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": product_details}}

	_, err = userCollection.UpdateOne(ctx, filter2, update2)

	if err != nil {
		log.Println(err)
	}

	return nil
}
