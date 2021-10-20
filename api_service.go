package main

import (
	"time"
	"strconv"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/render"
)

type APIService struct {
	router	*chi.Mux
	mc		DataStorage
}

func NewAPIService(mc DataStorage) *APIService {
	// Prepare Router
	return &APIService{router: chi.NewRouter(), mc: mc,}
}

func (api *APIService) Serve() {
	api.router.Use(middleware.Logger)

	api.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

    // "/price" resource
	api.router.Route("/price", func(r chi.Router) {
		// GET /price 
		r.With(paginate).Get("/", api.getLatestPrice)

		// GET /articles/2021-01-09T12:34:56
		r.With(paginate).Get("/{year:[0-9]+}-{month:[0-9]+}-{day:[0-9]+}T{hour:[0-9]+}:{minute:[0-9]+}:{second:[0-9]+}", api.getPriceByTimestamp) 
    })

    api.router.Get("/average", api.getAveragePrice)
	http.ListenAndServe(":80", api.router)
}

type averagePrice struct{
	ID			int64       `json:"_id"`
	Price		float64		`json:"price"`
}

type averageResponse  struct{
	From		time.Time	`json:"from"`
	To			time.Time   `json:"to"`
	Price		float64		`json:"price"`
}

type priceResponse  struct{
	TimeStamp	time.Time   `json:"ts"`
	Price		float64		`json:"price"`
}

type ErrResponse struct {
	Err				error	`json:"-"` // low-level runtime error
	HTTPStatusCode	int		`json:"http_code"` // http response status code
	StatusText		string	`json:"status"`          // user-level status message
	AppCode			int64	`json:"code,omitempty"`  // application-specific error code
	ErrorText		string	`json:"error,omitempty"` // application-level error message, for debugging
}


func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

var errInternalServerError = &ErrResponse{HTTPStatusCode: 500, StatusText: "Internal Server Error."}

func (api *APIService) getLatestPrice(w http.ResponseWriter, r *http.Request) {
	price, ts, err := api.mc.GetLatestPrice()
    if err != nil {
		render.Respond(w, r, errInternalServerError)
	}
	render.JSON(w, r, &priceResponse{TimeStamp: ts, Price: price})
}

func (api *APIService) getPriceByTimestamp(w http.ResponseWriter, r *http.Request) {
    // The regex should take care of it
	year, _ := strconv.Atoi(chi.URLParam(r, "year"))
	month, _ := strconv.Atoi(chi.URLParam(r, "month"))
	day, _ := strconv.Atoi(chi.URLParam(r, "day"))
	hour, _ := strconv.Atoi(chi.URLParam(r, "hour"))
	minute, _ := strconv.Atoi(chi.URLParam(r, "minute"))
	second, _ := strconv.Atoi(chi.URLParam(r, "second"))

    // Using timezone UTC right now for simplicity
    ts1 := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)

	price, err := api.mc.GetPriceByTimestamp(ts1)
    if err != nil {
        render.Respond(w, r, errInternalServerError)
		return
    }

	render.JSON(w, r, &priceResponse{TimeStamp: ts1, Price: price})
}

func (api *APIService) getAveragePrice(w http.ResponseWriter, r *http.Request) {
    //command in mongodb console

	fromStr := r.FormValue("from")
	toStr := r.FormValue("to")
    if fromStr == "" || toStr == "" {
        render.Respond(w, r, &ErrResponse{HTTPStatusCode: 400, StatusText: "Both 'from' and 'to' url parameter is required."})
		return
	}

	//ts_from, err := time.Parse(time.RFC3339, fromStr)
	from, err := time.Parse("2006-01-02T15:04:05", fromStr)
	if err != nil {
        render.Respond(w, r, &ErrResponse{HTTPStatusCode: 400, StatusText: "from = '" + fromStr + "' does not in format like YYYY-MM-DDTHH:MM:SS"})
		return
	}

	to, err := time.Parse("2006-01-02T15:04:05", toStr)
	if err != nil {
        render.Respond(w, r, &ErrResponse{HTTPStatusCode: 400, StatusText: "to= '" + toStr + "' does not in format like YYYY-MM-DDTHH:MM:SS"})
		return
	}

	price, err := api.mc.GetAveragePrice(from, to)
    if err != nil {
		if price < 0 {
	        render.Respond(w, r, errInternalServerError)
		}
	    render.Respond(w, r, &ErrResponse{HTTPStatusCode: 400, StatusText: err.Error()})
		return
    }

	render.JSON(w, r, &averageResponse{ From: from, To: to, Price: price})
}
