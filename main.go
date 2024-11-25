package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sumit-behera-in/go-storage-handler-api/controllers"
	db "github.com/sumit-behera-in/go-storage-handler/db"
)

var (
	DbCollection db.DBCollection
	DBController controllers.DBController
	server       *gin.Engine
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connectionString := os.Getenv("CONNECTION_STRING")
	database := os.Getenv("DATABASE_NAME")
	collectionName := os.Getenv("COLLECTION_NAME")

	clientOptions := options.Client().ApplyURI(connectionString)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	DbCollection.Project = "GO-STORAGE-HANDLER"

	mongoCollection := client.Database(database).Collection(collectionName)

	filter := bson.D{
		{Key: "project", Value: DbCollection.Project},
	}

	err = mongoCollection.FindOne(ctx, filter).Decode(&DbCollection)
	if err != nil {
		log.Fatal(err)
	}

	// create a gin server
	server = gin.Default()
}

func main() {
	dbClient, err := db.New(DbCollection)
	fmt.Println(DbCollection)
	if err != nil {
		log.Fatal("Db client init failed with error:", err.Error())
	}

	defer dbClient.Close()

	// create db controller

	DBController = *controllers.New(&dbClient)

	// base path for gin
	basePath := server.Group("/v1")

	DBController.RegisterUserRoutes(basePath)

	log.Fatal(server.Run(":9090"))
}
