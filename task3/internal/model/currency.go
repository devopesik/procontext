package model

import "time"

type CurrencyRate struct {
	Name string
	Rate float64
	Date time.Time
}
