package model

import "time"

type CurrentPrice struct {
	Time       CurrentPriceTime `json:"time"`
	Disclaimer string           `json:"disclaimer"`
	ChartName  string           `json:"chartName"`
	Bpi        CurrentPriceBpi  `json:"bpi"`
}

type CurrentPriceTime struct {
	Updated    string    `json:"updated"` // TODO might be good to code custom Unmarshal
	UpdatedISO time.Time `json:"updatedISO"`
	Updateduk  string    `json:"updateduk"` // TODO might be good to code custom Unmarshal
}

type CurrentPriceBpi struct {
	Usd CurrentPriceRate `json:"USD"`
	Gbp CurrentPriceRate `json:"GBP"`
	Eur CurrentPriceRate `json:"EUR"`
}

type CurrentPriceRate struct {
	Code        string  `json:"code"`
	Symbol      string  `json:"symbol"`
	Rate        string  `json:"rate"`
	Description string  `json:"description"`
	RateFloat   float64 `json:"rate_float"`
}
