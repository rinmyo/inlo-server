package service

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func NewMongoClient(uri string) (*mongo.Client, func() error, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://" + uri))
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, nil, err
	}
	return client, func() error {
		return client.Disconnect(ctx)
	}, nil
}

func NewLogCollection(db *mongo.Database, logFormat string) *mongo.Collection {
	return db.Collection(time.Now().Format(logFormat))
}

type UserCollection struct {
	*mongo.Collection
}

func NewUserCollection(db *mongo.Database) *UserCollection {
	return &UserCollection{db.Collection("users")}
}

func (uc *UserCollection) Find(username string) (*User, error) {
	var result User
	if err := uc.Collection.
		FindOne(nil, bson.M{
			"username": username,
		}).Decode(&result); err != nil {
		return nil, err
	}
	return result.Clone(), nil
}

func (uc *UserCollection) Save(user *User) error {
	result, err := uc.InsertOne(nil, user)
	if err != nil {
		return err
	}
	log.Info("inserted: ", result.InsertedID)
	return nil
}
