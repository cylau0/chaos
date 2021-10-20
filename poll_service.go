package main

import (
	"fmt"
	"time"
	"os"
	"io"
	"log"
	"net/http"
	"encoding/json"
	"github.com/go-co-op/gocron"

)
const (
	poll_url = "https://api.coinstats.app/public/v1/tickers?exchange=yobit&pair=BTC-USD"
	cronSec = 60
)

type PollService struct {
	scheduler	*gocron.Scheduler
	errChannel	chan error
	msgChannel	chan string
	mc			*MongoClient
}

func NewPollService(mc *MongoClient) *PollService {
	return &PollService{
		scheduler:	gocron.NewScheduler(time.UTC),
		errChannel:	make(chan error),
		msgChannel:	make(chan string),
		mc:			mc,
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

	for i := 0; i < len(tkts); i++ {
		tkts[i].Timestamp = ts
		id, err := p.mc.InsertOne(tkts[i])
		if err != nil { p.errChannel <- err ; continue }
		p.msgChannel <- id
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
				log.Println("Record created : " + id)
		}
	}
}
