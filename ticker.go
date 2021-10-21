package main

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type Ticker struct {
	ID			*primitive.ObjectID	`json:"_id,omitempty" bson:"_id"`
	From		string				`json:"from" bson:"from"`
	To			string				`json:"to" bson:"to"`
	Exchange	string				`json:"exchange" bson:"exchange"`
	Price		float64				`json:"price" bson:"price"`
	Timestamp	time.Time			`json:"ts,omitempty" bson:"ts"`
	TS			int64				`json:"-" bson:"-"`
}

type Tickers struct {
    Tickers     []Ticker    `json:"tickers"`
}

