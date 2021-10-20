package main

import (
	"time"
)

type Ticker struct {
	From		string				`json:"from" bson:"from"`
	To			string				`json:"to" bson:"to"`
	Exchange	string				`json:"exchange" bson:"exchange"`
	Price		float64				`json:"price" bson:"price"`
	Timestamp	time.Time			`json:"ts,omitempty" bson:"ts"`
}

type MTicker struct {
	Ticker
	ID			string 
	TS			int64
}

type Tickers struct {
    Tickers     []Ticker    `json:"tickers"`
}

