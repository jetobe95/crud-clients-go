package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/jetobe95/crud-clients-go/models"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var collection *mongo.Collection
var ctx = context.TODO()

func main() {
	godotenv.Load()
	connect()
	app := &cli.App{
		Name:  "Clients",
		Usage: "A simple CLI program to manage your clients",
		Commands: []*cli.Command{
			{
				Name:    "Add",
				Aliases: []string{"a"},
				Usage:   "Add a client to the list",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true, Aliases: []string{"n"}},
					&cli.StringFlag{Name: "email", Required: true, Aliases: []string{"e"}},
					&cli.StringFlag{Name: "gender", Required: true, Aliases: []string{"g"}},
				},
				Action: func(c *cli.Context) error {
					name := c.String("name")
					email := c.String("email")
					gender := c.String("gender")

					client := models.Client{
						ID:     primitive.NewObjectID(),
						Name:   name,
						Email:  email,
						Gender: gender,
					}
					return Create(&client)
				},
			},
			{
				Name:    "all",
				Aliases: []string{"l"},
				Usage:   "List all clients",
				Action: func(c *cli.Context) error {
					clients, err := getAll()
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("Nothing to see here.\nRun `add 'task'` to add a task")
							return nil
						}

						return err
					}
					printClients(clients)
					return nil
				},
			},
			{
				Name:    "read",
				Aliases: []string{"r"},
				Usage:   "Read a client by id",

				Action: func(c *cli.Context) error {
					id := c.Args().First()
					err := read(id)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Update a client",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Required: true, Aliases: []string{"i"}},
					&cli.StringFlag{Name: "name", Required: true, Aliases: []string{"n"}},
					&cli.StringFlag{Name: "email", Required: true, Aliases: []string{"e"}},
					&cli.StringFlag{Name: "gender", Required: true, Aliases: []string{"g"}},
				},

				Action: func(c *cli.Context) error {
					id := c.String("id")
					name := c.String("name")
					email := c.String("email")
					gender := c.String("gender")
					err := edit(id, name, email, gender)
					if err != nil {
						return (err)
					}

					return nil
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Action: func(c *cli.Context) error {
					id := c.Args().First()
					err := delete(id)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func connect() {
	uri := os.Getenv("MONGO_URI")
	// Create a new client and connect to the server
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	collection = client.Database("crud-clients-go").Collection("clients")

}

func Create(client *models.Client) error {
	_, err := collection.InsertOne(ctx, client)
	return err
}
func getAll() ([]*models.Client, error) {
	var clients []*models.Client

	cur, err := collection.Find(ctx, bson.D{{}})
	if err != nil {
		return clients, nil
	}
	for cur.Next(ctx) {
		var c models.Client
		err := cur.Decode(&c)
		if err != nil {
			return clients, err
		}
		clients = append(clients, &c)
	}

	if err := cur.Err(); err != nil {
		return clients, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(clients) == 0 {
		return clients, mongo.ErrNoDocuments
	}

	return clients, nil

}

func printClients(clients []*models.Client) {
	for _, v := range clients {
		color.Green("%s %s %s %s %s %s %s\n", v.ID.Hex(), "Name:", v.Name, "Email:", v.Email, "Gender:", v.Gender)
	}
}

func read(id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(err)
	}
	filter := bson.D{
		{
			Key: "_id", Value: objectId,
		},
	}

	var result models.Client

	errF := collection.FindOne(ctx, filter, options.FindOne()).Decode(&result)

	if errF != nil {
		panic(errF)
	}

	printClients([]*models.Client{&result})

	return nil
}

func edit(id, name, email, gender string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{
				Key:   "name",
				Value: name,
			},
			{
				Key:   "email",
				Value: email,
			},
			{
				Key:   "gender",
				Value: gender,
			},
		}},
	}
	updateResult, errUpdate := collection.UpdateByID(ctx, objectId, update)
	if errUpdate != nil {
		panic(errUpdate)
	}
	if updateResult.MatchedCount == 0 {
		return errors.New("client not found")
	}

	color.Green("Client was updated")
	return nil
}

func delete(id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.D{
		{
			Key:   "_id",
			Value: objectId,
		},
	}
	deleteResult, deleteErr := collection.DeleteOne(ctx, filter)
	if deleteErr != nil {
		panic(deleteErr)
	}
	if deleteResult.DeletedCount == 0 {
		return errors.New("client not found")
	}

	color.Green("Client was deleted")
	return nil

}
