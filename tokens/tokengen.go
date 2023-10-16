package token

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Ricardo-Cardozo/ecommerce_golang/database"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	SECRET_KEY string            = os.Getenv("SECRET_KEY")
	UserData   *mongo.Collection = database.UserData(database.Client, "Users")
)

type SignedDetails struct {
	Email      string
	First_Name string
	Uid        string
	jwt.StandardClaims
}

func TokenGenerator(
	email string,
	firstname string,
	lastname string,
	uid string,
) (accesstoken string, refreshtoken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_Name: firstname,
		Uid:        uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshclaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	refreshtoken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshclaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return token, refreshtoken, nil
}

func ValidateToken(accesstoken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		accesstoken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)

	if !ok {
		msg = "Token claims are not of expected type"
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "token is already expired"
		return
	}

	return claims, msg

}

func UpdateAllTokens(accestoken string, refreshtoken string, userid string) (string, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		100*time.Second,
	)

	defer cancel()

	var updateobj primitive.D

	updateobj = append(updateobj, bson.E{Key: "token", Value: accestoken})
	updateobj = append(updateobj, bson.E{Key: "refresh_token", Value: refreshtoken})
	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateobj = append(updateobj, bson.E{Key: "updated_at", Value: updated_at})
	upsert := true
	filter := bson.M{"user_id": userid}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := UserData.UpdateOne(ctx, filter, bson.D{{
		Key:   "$set",
		Value: updateobj,
	}}, &opt)

	if err != nil {
		log.Panic(err.Error())
	}

	return "Succesfully token updated!", nil
}
