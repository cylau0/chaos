package main

import (
	"fmt"
	"time"
	"context"
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type MemoryStorage struct {
	db map[string]*Ticker
	tree *redblacktree.Tree
}


func NewMemoryStorage() (*MemoryStorage) {
	db := make(map[string]*Ticker)
	tree := redblacktree.NewWith(utils.Int64Comparator)
	return &MemoryStorage{db: db, tree: tree}
}

func (m *MemoryStorage) Connect() (context.CancelFunc, error) {
	return func() {}, nil
}

func (m *MemoryStorage) Close() { }

func (m *MemoryStorage) InsertOne(o interface{}) ( string, error ) {
	n := o.(Ticker)
	n.Timestamp = n.Timestamp.UTC()
	ID := primitive.NewObjectID()
	n.ID = &ID
	n.TS = n.Timestamp.UnixMicro()
	m.db[n.ID.Hex()] = &n
	m.tree.Put(n.TS, n.ID.Hex())

	return n.ID.Hex(), nil
}


func (m *MemoryStorage) GetLatestPrice() (float64, time.Time, error) {
	values := m.tree.Values()
	ID := values[len(values)-1].(string)
	t := m.db[ID]
	return t.Price, t.Timestamp, nil
}



func (m *MemoryStorage) GetPriceByTimestamp(ts1 time.Time) (float64, error) {
	keys := m.tree.Keys()
	values := m.tree.Values()
	TS := ts1.UnixMicro()

	if TS < keys[0].(int64) || TS > keys[len(keys)-1].(int64) {
		return -1 , fmt.Errorf("Price is out of the data range: from = " + ts1.String())
	}

	for i, t := range keys {
		if TS == t.(int64) {
			return m.db[values[i].(string)].Price, nil
		}
		if i == 0 {
			continue
		}
		if TS < t.(int64) {
			lt := m.db[values[i-1].(string)]
			ut := m.db[values[i].(string)]
    		return InterpolatePrice(lt.Timestamp, lt.Price, ut.Timestamp, ut.Price, ts1), nil
		}
	}
	return -1, nil
}

func (m *MemoryStorage) GetAveragePrice(from, to time.Time) ( float64, error ) {
	okeys := m.tree.Keys()
	ovalues := m.tree.Values()
	keys := []int64{}
	values := []string{}
	for _, v := range okeys {
		keys = append(keys, v.(int64))
	}

	for _, v := range ovalues {
		values = append(values, v.(string))
	}

	if from.UnixMicro() < keys[0] {
		return -1 , fmt.Errorf("Price is out of the data range: from = " + from.String())
	}

	if to.UnixMicro() > keys[len(keys)-1] {
		return -1 , fmt.Errorf("Price is out of the data range: to = " + to.String())
	}
	
	return CalculateAveragePrice(keys, values, m.db, from, to)
}
