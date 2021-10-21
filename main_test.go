package main

import (
    "testing"
	"log"
	"time"
	"io"
	"net/http"
	"encoding/json"
	"github.com/stretchr/testify/assert"
)

func TestAPIServer(t *testing.T) {
	log.Println("TestAPIServer")

	mc := NewMemoryStorage()
	start_time := time.Date(2021, time.August, 1, 0, 0, 0, 0, time.UTC)
	ts := start_time
	end_time := start_time
	prices := []float64{64000, 63000, 62000, 61000, 60000, 59000}
	
	for _, p := range prices {
		t := Ticker{From: "BTC", To: "USD", Exchange: "Yobit", Price: p, Timestamp: ts}
		mc.InsertOne(t)
		log.Printf("Insert : %v\n", t)
		end_time = ts
		ts = ts.Add(time.Minute)
	}
	r := NewAPIService(mc)
	go r.Serve(":8080")

	mid := int(len(prices)/2)
	mid_time := start_time
    for i := 0; i < mid -1 ; i++ {
		mid_time = mid_time.Add(time.Minute)
	}
	mid_time = mid_time.Add(30 * time.Second)

	midPrice := (prices[mid-1] + prices[mid]) / float64(2.0)

	time.Sleep(2 * time.Second)

	latestPrice := prices[len(prices)-1] 

	// Get Average
	sumPrice := float64(0.0)
	for _, p := range prices {
		sumPrice += p	
	}
	averagePrice := sumPrice / float64(len(prices))

	testLatestPrice(t, latestPrice)
	testTimePrice(t, midPrice, mid_time)
	testAveragePrice(t, averagePrice, start_time, end_time)
}

func getJsonOutput(url string) map[string]interface{} {
	log.Println("Using URL = " + url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	blob , err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(blob), &result); err != nil {
		log.Fatal(err)
	}
	return result
}

func testLatestPrice(t *testing.T, latestPrice float64) {
	url := "http://localhost:8080/price"
	result := getJsonOutput(url)
	assert.Equal(t, latestPrice, result["price"], "Latest Price is incorrect")
}

func testTimePrice(t *testing.T, midPrice float64, midTime time.Time) {
	url := "http://localhost:8080/price/" + midTime.Format("2006-01-02T15:04:05")
	result := getJsonOutput(url)
	assert.Equal(t, midPrice, result["price"], "/price/YYYY-mm-ddTHH:MM:SS")
}

func testAveragePrice(t *testing.T, averagePrice float64, from, to time.Time) {
	fromStr := from.Format("2006-01-02T15:04:05")
	toStr := to.Format("2006-01-02T15:04:05")
	url := "http://localhost:8080/average?from="+fromStr+"&to="+toStr
	result := getJsonOutput(url)
	assert.Equal(t, averagePrice, result["price"], "Average Price is not correct")
}



/*
func TestMain(t *testing.T) {
	log.Println("TestMain")
	mc := NewMemoryStorage()
	mc.InsertOne
	mc := NewMongoStorage(10 * time.Second)
    main_impl(mc)
	for;; {}
}
*/
