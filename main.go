package main

import (
	"time"
)

func main() {
	mc := NewMongoClient(10 * time.Second)
	s := NewPollService(mc)
	s.Start()
	go s.Loop()

	r := NewAPIService(mc)
	r.Serve()

}
