package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	db "github.com/sumit-behera-in/go-storage-handler/db"
)

func main() {
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

	var dbCollection db.DBCollection
	dbCollection.Project = "GO-STORAGE-HANDLER"

	mongoCollection := client.Database(database).Collection(collectionName)

	filter := bson.D{
		{Key: "project", Value: dbCollection.Project},
	}

	err = mongoCollection.FindOne(ctx, filter).Decode(&dbCollection)
	if err != nil {
		log.Fatal(err)
	}

	myclient, err := db.New(dbCollection)

	if err != nil {
		log.Fatal(err)
	}

	data := db.Data{
		FileName: "README.md",
		FileType: "md",
	}
	data.File, _ = os.ReadFile("README.md")

	myclient.Upload(data, 0)
	
}
