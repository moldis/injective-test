package model

type Prices struct {
	CreatedAt int64      `bson:"created_at"`
	Price     PricesInfo `bson:"price"`
}

type PricesInfo struct {
	Disclaimer string                `bson:"disclaimer"`
	ChartName  string                `bson:"chartName"`
	Bpi        PricesCurrentPriceBpi `bson:"bpi"`
}

type PricesCurrentPriceBpi struct {
	Usd PricesCurrentPriceRate `bson:"USD"`
	Gbp PricesCurrentPriceRate `bson:"GBP"`
	Eur PricesCurrentPriceRate `bson:"EUR"`
}

type PricesCurrentPriceRate struct {
	Code        string  `bson:"code"`
	Symbol      string  `bson:"symbol"`
	Rate        string  `bson:"rate"`
	Description string  `bson:"description"`
	RateFloat   float64 `bson:"rate_float"`
}
