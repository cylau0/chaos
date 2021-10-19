package main

import (
	"fmt"
	"time"
	"os"
	"io"
	"log"
	"context"
	"net/http"
	"encoding/json"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

)
const (
	poll_url = "https://api.coinstats.app/public/v1/tickers?exchange=yobit&pair=BTC-USD"
	cronSec = 60
)

type PollService struct {
	scheduler  *gocron.Scheduler
	errChannel chan error
	msgChannel chan primitive.ObjectID
}

func NewPollService() *PollService {
	return &PollService{
		scheduler:	gocron.NewScheduler(time.UTC),
		errChannel:	make(chan error),
		msgChannel:	make(chan primitive.ObjectID),
	}
}

func (p *PollService) pollURL() {
	ts := time.Now()
	resp, err := http.Get(poll_url)
	if err != nil { p.errChannel <- err ; return }

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil { p.errChannel <- err ; return }

	var tickers Tickers
	err = json.Unmarshal([]byte(body), &tickers)
	tkts := tickers.Tickers
	if err != nil { p.errChannel <- err ; return }

	// Validate
	if len(tkts) == 0 {
		p.errChannel <- fmt.Errorf("Zero Length of Returned array.")
		return
	}

	if tkts[0].From != "BTC" && tkts[0].To != "USD" {
		p.errChannel <- fmt.Errorf("Is not valid BTC-USD Pair Result")
		return
	}

	mc, err := getMongoClient(10)

	if err != nil { p.errChannel <- err ; return }

	col := mc.Database("local").Collection("rates")

	for i := 0; i < len(tkts); i++ {
		tkts[i].Timestamp = ts
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		bsonBytes, _ := bson.Marshal(tkts[i])
		res, err := col.InsertOne(ctx, bsonBytes)
		if err != nil { p.errChannel <- err ; continue }
		id := res.InsertedID
		p.msgChannel <- id.(primitive.ObjectID)
	}
}

func (p *PollService) Start() {
	p.scheduler.Every(cronSec).Seconds().Do(p.pollURL)
	p.scheduler.StartAsync()
}

func (p *PollService) Loop() {
	for ;; {
		select {
       		case err := <- p.errChannel:
				log.Println(err)
				if err == nil {
					os.Exit(0)
				}
			// here comes the created object in mongodb
			case id := <- p.msgChannel:
				log.Println("Record created : " + id.Hex())
		}
	}
}
