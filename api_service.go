package main

import (
	"time"
	"strconv"
	"context"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type APIService struct {
	router *chi.Mux
}

func NewAPIService() *APIService {
	// Prepare Router
	return &APIService{router: chi.NewRouter()}
}

func (r *APIService) Serve() {
	r.router.Use(middleware.Logger)

	r.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

    // "/price" resource
	r.router.Route("/price", func(rr chi.Router) {
		// GET /price 
		rr.With(paginate).Get("/", getLatestPrice)

		// GET /articles/2021-01-09T12:34:56
		rr.With(paginate).Get("/{year:[0-9]+}-{month:[0-9]+}-{day:[0-9]+}T{hour:[0-9]+}:{minute:[0-9]+}:{second:[0-9]+}", getPriceByTimestamp) 
    })

    r.router.Get("/average", getAveragePrice)
	http.ListenAndServe(":80", r.router)
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

func getTicker(w http.ResponseWriter, r *http.Request, cur *mongo.Cursor) *Ticker {
	var t Ticker
	for cur.Next(context.Background()) {
	  	err := cur.Decode(&t)
    	if err != nil { render.Respond(w, r, errInternalServerError) }
		return &t
	}
    if err := cur.Err(); err != nil {
		render.Respond(w, r, errInternalServerError)
    }
	return nil
}

func getLatestPrice(w http.ResponseWriter, r *http.Request) {
	mc, err := getMongoClient(10)
    if err != nil {
		render.Respond(w, r, errInternalServerError)
	}
	col := mc.Database("local").Collection("rates")

	findOpts := options.Find()
	findOpts.SetSort(bson.D{{"ts", -1}})
	findOpts.SetLimit(1)

	cur, err := col.Find(context.Background(), bson.D{}, findOpts)
	if err != nil { render.Respond(w, r, errInternalServerError) }
	defer cur.Close(context.Background())

	t := getTicker(w, r, cur)
	render.JSON(w, r, t)
}

func getPriceByTimestamp(w http.ResponseWriter, r *http.Request) {
    // The regex should take care of it
	year, _ := strconv.Atoi(chi.URLParam(r, "year"))
	month, _ := strconv.Atoi(chi.URLParam(r, "month"))
	day, _ := strconv.Atoi(chi.URLParam(r, "day"))
	hour, _ := strconv.Atoi(chi.URLParam(r, "hour"))
	minute, _ := strconv.Atoi(chi.URLParam(r, "minute"))
	second, _ := strconv.Atoi(chi.URLParam(r, "second"))

    // Using timezone UTC right now for simplicity
    ts1 := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
    ts2 := time.Date(year, time.Month(month), day, hour, minute, second + 1, 0, time.UTC)

	mc, err := getMongoClient(10)
    if err != nil { render.Respond(w, r, errInternalServerError) }
	col := mc.Database("local").Collection("rates")

	findOpts := options.Find()
	findOpts.SetLimit(1)

	//run in mongo console
    //db.rates.find({"from": "BTC", "to": "USD", "ts": {"$gte": ts1, "$lt":  ts2 }  })
	// Search by range for a particular second range
    filter0 := bson.M{"from": "BTC", "to": "USD", "ts": bson.M{"$gte": ts1, "$lt": ts2}}

	cur, err := col.Find(context.Background(), filter0, findOpts)
	if err != nil { render.Respond(w, r, errInternalServerError) }
	defer cur.Close(context.Background())
    t0 := getTicker(w, r, cur)

	if t0 != nil {
		render.JSON(w, r, t0)
		return
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
    t1:= getTicker(w, r, cur)

	cur, err = col.Find(context.Background(), filter2, findOpts2)
    t2:= getTicker(w, r, cur)

	// If no earlier ticket is found
    if t1 == nil {
		t2.Timestamp = ts1
		render.JSON(w, r, t2)
		return
	}

    if t2 == nil {
		t1.Timestamp = ts1
		render.JSON(w, r, t1)
		return
	}

    t3 := Ticker{ From: t1.From, To: t1.To, Exchange: t1.Exchange, Timestamp: ts1 }
    dur := float64(t2.Timestamp.Sub(t1.Timestamp))
	t3.Price = t1.Price * float64(ts1.Sub(t1.Timestamp)) / dur + t2.Price * float64(t2.Timestamp.Sub(ts1)) / dur
	render.JSON(w, r, t3)
}

func getAveragePrice(w http.ResponseWriter, r *http.Request) {
    //command in mongodb console

	fromStr := r.FormValue("from")
	toStr := r.FormValue("to")
    if fromStr == "" || toStr == "" {
        render.Respond(w, r, &ErrResponse{HTTPStatusCode: 400, StatusText: "Both 'from' and 'to' url parameter is required."})
		return
	}

	//ts_from, err := time.Parse(time.RFC3339, fromStr)
	time_from, err := time.Parse("2006-01-02T15:04:05", fromStr)
	if err != nil {
        render.Respond(w, r, &ErrResponse{HTTPStatusCode: 400, StatusText: "from = '" + fromStr + "' does not in format like YYYY-MM-DDTHH:MM:SS"})
		return
	}

	time_to, err := time.Parse("2006-01-02T15:04:05", toStr)
	if err != nil {
        render.Respond(w, r, &ErrResponse{HTTPStatusCode: 400, StatusText: "to= '" + toStr + "' does not in format like YYYY-MM-DDTHH:MM:SS"})
		return
	}
	ts_from := primitive.NewDateTimeFromTime(time_from)
	ts_to := primitive.NewDateTimeFromTime(time_to)

	mc, err := getMongoClient(10)
    if err != nil {
		render.Respond(w, r, errInternalServerError)
		return
	}
	col := mc.Database("local").Collection("rates")

	// MongoDB Console Command Example
    // db.rates.aggregate( [ {$match: { from: "BTC", to: "USD", exchange: "Yobit", ts: { "$gte": ISODate("2021-10-18T17:03:52.647Z"), "$lte": ISODate("2021-10-19T17:03:52.647Z") } } }, { $group: { _id: null, price: { $avg: "$price" } } } ] )

	matchStage := bson.D{{"$match", bson.D{{"from", "BTC"}, {"to", "USD"}, {"ts", bson.D{{"$gte", ts_from}, {"$lte", ts_to}}}}}}
	groupStage := bson.D{{"$group", bson.D{{"_id", nil}, {"price", bson.D{{"$avg", "$price"}}}}}}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
    cur, err := col.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})

    var avgPrice averagePrice
    for cur.Next(context.Background()) {
        err := cur.Decode(&avgPrice)
        if err != nil {
			render.Respond(w, r, errInternalServerError)
			return
		}
		render.JSON(w, r, &averageResponse{ From: time_from, To: time_to, Price: avgPrice.Price})
		return
    }

    if err := cur.Err(); err != nil {
        render.Respond(w, r, errInternalServerError)
		return
    }
	render.JSON(w, r, &averageResponse{ From: time_from, To: time_to, Price: 0})

}
