package main

import (
	"os"
	"time"
	"context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
    default_mongodb_url = "mongodb://localhost:27017"
)

type MongoClient struct {
	db_url		string
	db_timeout	time.Duration
	ctx			context.Context
	mc			*mongo.Client
}

func NewMongoClient(timeout time.Duration) (*MongoClient) {
    db_url := os.Getenv("MONGODB_URL")
    if db_url == "" {
        db_url = default_mongodb_url
	}

	return &MongoClient{ db_url: db_url, db_timeout: timeout, mc: nil }
}

func (m *MongoClient) Connect() (context.CancelFunc, error) {
    ctx, cancel := context.WithTimeout(context.Background(), m.db_timeout)
	m.ctx = ctx

    mc, err := mongo.Connect(m.ctx, options.Client().ApplyURI(m.db_url))
	if err != nil {
		return cancel, err
	}
	m.mc = mc
	return cancel, nil
}

func (m *MongoClient) Close() {
	if m.ctx != nil {
		if m.mc != nil {
	    	m.mc.Disconnect(m.ctx)
		}
		m.ctx = nil
	}
	m.mc = nil
}

func (m *MongoClient) InsertOne(o interface{}) ( string, error ) {

    cancel, err := m.Connect()
	defer cancel()

	if err != nil {
		return "", err
	}
    defer m.Close()

    col := m.mc.Database("local").Collection("rates")
    bsonBytes, _ := bson.Marshal(o)

    res, err := col.InsertOne(m.ctx, bsonBytes)
    if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func unmarshalTicker(cur *mongo.Cursor) (*Ticker, error) {
    var t Ticker
    for cur.Next(context.Background()) {
        err := cur.Decode(&t)
        if err != nil { 
			return nil, err
		}
        return &t, nil
    }
    if err := cur.Err(); err != nil {
        return nil, err
    }
    return nil, nil
}

func (m *MongoClient) GetLatestPrice() (float64, time.Time, error) {
    cancel, err := m.Connect()
	defer cancel()

    col := m.mc.Database("local").Collection("rates")

    findOpts := options.Find()
    findOpts.SetSort(bson.D{{"ts", -1}})
    findOpts.SetLimit(1)

    cur, err := col.Find(context.Background(), bson.D{}, findOpts)
    if err != nil {
		return -1, time.Now(), err
	}

    t, err := unmarshalTicker(cur)
    if err != nil {
		return -1, time.Now(), err
	}

    return t.Price, t.Timestamp, nil
}



func (m *MongoClient) GetPriceByTimestamp(ts1 time.Time) (float64, error) {
    ts2 := ts1.Add(1 * time.Second)

    cancel, err := m.Connect()
	defer cancel()

    if err != nil { 
		return -1, err
	}
    col := m.mc.Database("local").Collection("rates")

    findOpts := options.Find()
    findOpts.SetLimit(1)

    //run in mongo console
    //db.rates.find({"from": "BTC", "to": "USD", "ts": {"$gte": ts1, "$lt":  ts2 }  })

    // Search by range for a particular second range
    filter0 := bson.M{"from": "BTC", "to": "USD", "ts": bson.M{"$gte": ts1, "$lt": ts2}}

    cur, err := col.Find(context.Background(), filter0, findOpts)
    if err != nil {
		return -1, err
	}
    defer cur.Close(context.Background())

    t0, err := unmarshalTicker(cur)

	if err != nil {
		return -1, err
	}

    if t0 != nil {
        return t0.Price, nil
    }
    // If no exact result here, I try to intepolate between two enquiry
    findOpts1 := options.Find()
    findOpts1.SetSort(bson.D{{"ts", -1}})
    findOpts1.SetLimit(1)
    filter1 := bson.M{"from": "BTC", "to": "USD", "ts": bson.M{"$lt": ts1}}

    findOpts2 := options.Find()
    findOpts2.SetSort(bson.D{{"ts", 1}})
    findOpts2.SetLimit(1)
    filter2 := bson.M{"from": "BTC", "to": "USD", "ts": bson.M{"$gt": ts1}}

    cur, err = col.Find(context.Background(), filter1, findOpts1)
	if err != nil {
		return -1, err
	}

    t1, err := unmarshalTicker(cur)
	if err != nil {
		return -1, err
	}

    cur, err = col.Find(context.Background(), filter2, findOpts2)
	if err != nil {
		return -1, err
	}

    t2, err := unmarshalTicker(cur)
	if err != nil {
		return -1, err
	}

    // If no earlier ticket is found
    if t1 == nil {
		return t2.Price, nil
    }

    if t2 == nil {
		return t1.Price, nil
    }

    dur := float64(t2.Timestamp.Sub(t1.Timestamp))
    return t1.Price * (float64(ts1.Sub(t1.Timestamp)) / dur ) +  t2.Price * (float64(t2.Timestamp.Sub(ts1)) / dur ), nil

}

func (m *MongoClient) GetAveragePrice(from, to time.Time) ( float64, error ) {

	ts_from := primitive.NewDateTimeFromTime(from)
    ts_to := primitive.NewDateTimeFromTime(to)

    cancel, err := m.Connect()
	defer cancel()
	if err != nil {
		return -1, err
	}
    defer m.Close()

    col := m.mc.Database("local").Collection("rates")

    // MongoDB Console Command Example
    // db.rates.aggregate( [ {$match: { from: "BTC", to: "USD", exchange: "Yobit", ts: { "$gte": ISODate("2021-10-18T17:03:52.647Z"), "$lte": ISODate("2021-10-19T17:03:52.647Z") } } }, { $group: { _id: null, price: { $avg: "$price" } } } ] )

    matchStage := bson.D{{"$match", bson.D{{"from", "BTC"}, {"to", "USD"}, {"ts", bson.D{{"$gte", ts_from}, {"$lte", ts_to}}}}}}
    groupStage := bson.D{{"$group", bson.D{{"_id", nil}, {"price", bson.D{{"$avg", "$price"}}}}}}

    cur, err := col.Aggregate(m.ctx, mongo.Pipeline{matchStage, groupStage})
    if err != nil {
        return -1, err
    }

    var avgPrice averagePrice
    for cur.Next(context.Background()) {
        err := cur.Decode(&avgPrice)
        if err != nil {
			return -1, err
        }
        return avgPrice.Price, nil
    }

    if err := cur.Err(); err != nil {
		return -1, err
    }

	return 0, nil
}

//func (m *MongoClient) Find(o interface{}) ( []interface{}, error ) {
//}

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
