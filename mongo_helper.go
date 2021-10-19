package main

import (
	"os"
	"time"
	"context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

const (
    default_mongodb_url = "mongodb://localhost:27017"
)

func getMongoClient(timeout time.Duration) (*mongo.Client, error) {
    db_url := os.Getenv("MONGODB_URL")
    if db_url == "" {
        db_url = default_mongodb_url
    }
    ctx, cancel := context.WithTimeout(context.Background(), timeout * time.Second)
    defer cancel()
    mc, err := mongo.Connect(ctx, options.Client().ApplyURI(db_url))
    if err != nil {
        return nil, err
    }
    return mc, nil
}

