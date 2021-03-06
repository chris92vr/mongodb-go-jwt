package helpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chris92vr/mongodb-go-jwt/database"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SignedDetails
type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

// GenerateAllTokens generates both teh detailed token and refresh token
func GenerateAllTokens(email string, firstName string, lastName string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

//ValidateToken validates the jwt token
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}

	return claims, msg
}

//UpdateAllTokens renews the user tokens when they login
func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)
	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}

	return
}

func DeleteAllTokens(userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	filter := bson.M{"user_id": userId}

	_, err := userCollection.DeleteOne(
		ctx,
		filter,
	)
	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}

	return
}

//GetUserToken returns the user token
func GetUserToken(userId string) (token string, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var userToken struct {
		Token string `bson:"token"`
	}

	filter := bson.M{"user_id": userId}

	err = userCollection.FindOne(
		ctx,
		filter,
	).Decode(&userToken)
	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}

	return userToken.Token, err
}

//GetUserRefreshToken returns the user refresh token
func GetUserRefreshToken(userId string) (token string, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var userToken struct {
		Refresh_token string `bson:"refresh_token"`
	}

	filter := bson.M{"user_id": userId}

	err = userCollection.FindOne(
		ctx,
		filter,
	).Decode(&userToken)
	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}

	return userToken.Refresh_token, err
}

//GetUserId returns the user id
func GetUserId(signedToken string) (userId string, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var userToken struct {
		User_id string `bson:"user_id"`
	}

	filter := bson.M{"token": signedToken}

	err = userCollection.FindOne(
		ctx,
		filter,
	).Decode(&userToken)
	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}

	return userToken.User_id, err
} //GetUserId

