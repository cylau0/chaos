package main
import (
	"time"
	"context"
)

type DataStorage interface {
	Connect() (context.CancelFunc, error)
	Close() 
	InsertOne(o interface{}) ( string, error ) 
	GetLatestPrice() (float64, time.Time, error)
	GetPriceByTimestamp(ts1 time.Time) (float64, error) 
	GetAveragePrice(from, to time.Time) ( float64, error ) 
}
