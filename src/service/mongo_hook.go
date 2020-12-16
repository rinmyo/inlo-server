package service

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type hooker struct {
	c *mongo.Collection
}

type M bson.M

func NewHookerFromCollection(collection *mongo.Collection) *hooker {
	return &hooker{c: collection}
}

func (h *hooker) Fire(entry *logrus.Entry) error {
	data := make(logrus.Fields)
	data["Level"] = entry.Level.String()
	data["Time"] = entry.Time
	data["Message"] = entry.Message

	for k, v := range entry.Data {
		if errData, isError := v.(error); logrus.ErrorKey == k && v != nil && isError {
			data[k] = errData.Error()
		} else {
			data[k] = v
		}
	}

	_, err := h.c.InsertOne(nil, M(data))
	if err != nil {
		return fmt.Errorf("failed to send log entry to mongodb: %v", err)
	}

	return nil
}

func (h *hooker) Levels() []logrus.Level {
	return logrus.AllLevels
}
