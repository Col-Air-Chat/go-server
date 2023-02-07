package common

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	url         = "mongodb://localhost:27017"
)

func SetMongoDBConfig(u string) {
	url = u
}

func InitMongodb() {
	var err error
	mongoClient, err = mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		log.Fatal(err)
	}
	err = mongoClient.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func GetMongoDBClient() *mongo.Client {
	if mongoClient == nil {
		InitMongodb()
	}
	return mongoClient
}

func GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

