package main

import (
	"time"
)

func main_impl(mc DataStorage) {
	s := NewPollService(mc)
	s.Start()
	go s.Loop()

	r := NewAPIService(mc)
	r.Serve()

}

func main() {
	mc := NewMongoStorage(10 * time.Second)
	main_impl(mc)
}
