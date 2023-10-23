package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/fizzify/todoify/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

// Create a single global client instance
var mongoClient *mongo.Client

// SetupMongo creates the MongoDB client connection
func SetupMongo() error {

	ctx := context.TODO()

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))

	// Configure connection pooling options
	clientOptions.SetMaxPoolSize(50)
	clientOptions.SetMinPoolSize(25)
	clientOptions.SetMaxConnIdleTime(60 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	mongoClient = client
	return nil
}

// GetMongoDB returns a handle to the database
func GetMongoDB(dbName string) (*mongo.Database, error) {

	if mongoClient == nil {
		return nil, errors.New("MongoDB connection not initialized")
	}

	return mongoClient.Database(dbName), nil
}

// DisconnectMongo closes the MongoDB connection
func DisconnectMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if mongoClient == nil {
		return nil
	}

	err := mongoClient.Disconnect(ctx)
	mongoClient = nil
	return err
}

func main() {

	envErr := godotenv.Load(".env")
	if envErr != nil {
		fmt.Println("Could not load env file")
		os.Exit(1)
	}

	err := SetupMongo()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := DisconnectMongo()
		if err != nil {

		}
	}()

	mg, err := GetMongoDB("todoify")

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		filter := bson.D{}
		cursor, err := mg.Collection("Todo").Find(c.Context(), filter)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		var todos []models.Todo

		if err := cursor.All(c.Context(), &todos); err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.Render("index", fiber.Map{
			"Todos": todos,
		})
	})

	app.Post("/", func(c *fiber.Ctx) error {
		todo := new(models.Todo)

		if err := c.BodyParser(todo); err != nil {
			return c.Status(500).SendString(err.Error())
		}

		fmt.Println("HELLAUR???? 2")

		todo.ID = ""
		todo.Done = false

		newTodo, err := mg.Collection("Todo").InsertOne(context.TODO(), todo)

		if err != nil {
			fmt.Println(err.Error())
			return c.Status(500).SendString(err.Error())
		}

		filter := bson.D{{Key: "_id", Value: newTodo.InsertedID}}
		createdRecord := mg.Collection("Todo").FindOne(c.Context(), filter)

		createdTodo := &models.Todo{}

		err = createdRecord.Decode(createdTodo)
		if err != nil {
			return err
		}

		return c.Status(201).JSON(createdTodo)
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}
