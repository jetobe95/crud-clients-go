package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Client struct {
	ID     primitive.ObjectID `bson:"_id"`
	Name   string             `bson:"name"`
	Email  string             `bson:"email"`
	Gender string             `bson:"gender"`
}
