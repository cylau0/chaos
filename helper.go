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
		return float64(0.0)
	}
	return ( v1 + v2 ) * float64(t2.Sub(t1)) / 2.0
}

func CalculateAveragePrice(keys []int64, values []string, db map[string]*Ticker, from, to time.Time) ( float64, error ) {
	var price_from, price_to float64

	TS_from := from.UnixNano()
	TS_to := to.UnixNano()
	
    area_sum := float64(0.0)
	started := false
	for i, t := range keys {
		if TS_from == t {
			if TS_from == TS_to {
				return float64(0.0), nil
			}
			price_from = db[values[i]].Price
			started = true
			continue
		}
		if i == 0 {
			continue
		}
		lt := db[values[i-1]]
		ut := db[values[i]]

		if TS_to == t {
			price_to = db[values[i]].Price
			area_sum += PriceTimeProductSum(lt.Timestamp, lt.Price, ut.Timestamp, ut.Price)
			break
		}
		// Special Catering
		if !started && TS_from < t && TS_to < t {
    		price_from = InterpolatePrice(lt.Timestamp, lt.Price, ut.Timestamp, ut.Price, from)
    		price_to = InterpolatePrice(lt.Timestamp, lt.Price, ut.Timestamp, ut.Price, to)
			return PriceTimeProductSum(from, price_from, to, price_to), nil
		}
		if !started && TS_from < t {
    		price_from = InterpolatePrice(lt.Timestamp, lt.Price, ut.Timestamp, ut.Price, from)
			area_sum += PriceTimeProductSum(from, price_from, ut.Timestamp, ut.Price)
			started = true
			continue
		}
		if started && TS_to < t {
    		price_to = InterpolatePrice(lt.Timestamp, lt.Price, ut.Timestamp, ut.Price, to)
			area_sum += PriceTimeProductSum(lt.Timestamp, lt.Price, to, price_to)
			break;
		}
		if started {
			area_sum += PriceTimeProductSum(lt.Timestamp, lt.Price, ut.Timestamp, ut.Price)
		}
	}

	return area_sum / float64(to.Sub(from)), nil
}
