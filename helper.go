package main
import (
	"time"
)
// 

func InterpolatePrice(t1 time.Time, v1 float64, t2 time.Time, v2 float64, tm time.Time) float64 {
    return ( v1 * float64(tm.Sub(t1)) +  v2 * float64(t2.Sub(tm)) ) / float64(t2.Sub(t1)) 
}

func PriceTimeProductSum(t1 time.Time, v1 float64, t2 time.Time, v2 float64) float64 {
	if t1 == t2 {
		return 0.0
	}
	return ( v1 + v2 ) * float64(t2.Sub(t1)) / 2.0
}
