package main

import (
	"time"
	"context"
	"github.com/hashicorp/go-memdb"
)


type MemoryStorage struct {
	timeout	time.Duration
	ctx		context.Context
	db		*memdb.MemDB
}


func NewMemoryStorage(timeout time.Duration) (*MemoryStorage, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"person": &memdb.TableSchema{
				Name: "person",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"age": &memdb.IndexSchema{
						Name:    "age",
						Unique:  false,
						Indexer: &memdb.IntFieldIndex{Field: "Age"},
					},
				},
			},
		},
	}
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}
	return &MemoryStorage{db: db, timeout: timeout}, nil
}

func (m *MemoryStorage) Connect() (context.CancelFunc, error) {
    ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	m.ctx = ctx
	return cancel, nil
}

func (m *MemoryStorage) Close() {
	m.ctx = nil
}

func (m *MemoryStorage) InsertOne(o interface{}) ( string, error ) {

    cancel, err := m.Connect()
	defer cancel()

	if err != nil {
		return "", err
	}
    defer m.Close()
	return "TBFILL", nil
/*

    col := m.mc.Database("local").Collection("rates")
    bsonBytes, _ := bson.Marshal(o)

    res, err := col.InsertOne(m.ctx, bsonBytes)
    if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
*/
}


func (m *MemoryStorage) GetLatestPrice() (float64, time.Time, error) {
    cancel, err := m.Connect()
	defer cancel()
	if err != nil {
		return -1, time.Now(), err
	}
	return -1, time.Now(), nil
/*
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
*/
}



func (m *MemoryStorage) GetPriceByTimestamp(ts1 time.Time) (float64, error) {
    //ts2 := ts1.Add(1 * time.Second)

    cancel, err := m.Connect()
	defer cancel()

    if err != nil { 
		return -1, err
	}
	return -1, nil
/*
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
*/
}

func (m *MemoryStorage) GetAveragePrice(from, to time.Time) ( float64, error ) {

    cancel, err := m.Connect()
	defer cancel()
    if err != nil { 
		return -1, err
	}
	return -1, nil
/*
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
*/
}
