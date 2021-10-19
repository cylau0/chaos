package main

func main() {
	s := NewPollService()
	s.Start()
	go s.Loop()

	r := NewAPIService()
	r.Serve()

}



