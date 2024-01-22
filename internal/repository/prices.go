package repository

import (
	"context"
	"time"

	"code.injective.org/service/pricefetcher/internal/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collection = "prices"

//go:generate mockery --name=Prices --structname=MockPrices --outpkg=repository --output ./ --filename prices_mock.go
type Prices interface {
	Create(ctx context.Context, in *model.CurrentPrice) error
	GetSinceDate(ctx context.Context, date time.Time) ([]*model.CurrentPrice, error)
}

type prices struct {
	pool *mongo.Database
}

func NewPrices(conn *mongo.Database) *prices {
	return &prices{
		pool: conn,
	}
}

func (a *prices) Create(ctx context.Context, in *model.CurrentPrice) error {
	res, err := a.pool.Collection(collection).InsertOne(ctx, a.toPrice(in))
	if err != nil {
		return err
	}
	log.Debug().Msgf("inserted new collection with ID %d", res)
	return nil
}

func (a *prices) GetSinceDate(ctx context.Context, sinceDate time.Time) ([]*model.CurrentPrice, error) {
	filter := bson.M{"created_at": bson.M{"$gt": sinceDate.UTC().Unix()}}
	cursor, err := a.pool.Collection(collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var dbResult []model.Prices
	if err = cursor.All(ctx, &dbResult); err != nil {
		panic(err)
	}

	var res []*model.CurrentPrice
	for _, result := range dbResult {
		cursor.Decode(&result)
		curPrice := a.fromPrice(&result)
		res = append(res, &curPrice)
	}

	return res, nil
}

func (a *prices) toPrice(in *model.CurrentPrice) model.Prices {
	return model.Prices{
		CreatedAt: in.Time.UpdatedISO.UTC().Unix(),
		Price: model.PricesInfo{
			Disclaimer: in.Disclaimer,
			ChartName:  in.ChartName,
			Bpi: model.PricesCurrentPriceBpi{
				Usd: model.PricesCurrentPriceRate{
					Code:        in.Bpi.Usd.Code,
					Symbol:      in.Bpi.Usd.Symbol,
					Rate:        in.Bpi.Usd.Rate,
					Description: in.Bpi.Usd.Description,
					RateFloat:   in.Bpi.Usd.RateFloat,
				},
				Gbp: model.PricesCurrentPriceRate{
					Code:        in.Bpi.Gbp.Code,
					Symbol:      in.Bpi.Gbp.Symbol,
					Rate:        in.Bpi.Gbp.Rate,
					Description: in.Bpi.Gbp.Description,
					RateFloat:   in.Bpi.Gbp.RateFloat,
				},
				Eur: model.PricesCurrentPriceRate{
					Code:        in.Bpi.Eur.Code,
					Symbol:      in.Bpi.Eur.Symbol,
					Rate:        in.Bpi.Eur.Rate,
					Description: in.Bpi.Eur.Description,
					RateFloat:   in.Bpi.Eur.RateFloat,
				},
			},
		},
	}
}

func (a *prices) fromPrice(in *model.Prices) model.CurrentPrice {
	return model.CurrentPrice{
		Time: model.CurrentPriceTime{
			UpdatedISO: time.Unix(in.CreatedAt, 10),
		},
		Disclaimer: in.Price.Disclaimer,
		ChartName:  in.Price.ChartName,
		Bpi: model.CurrentPriceBpi{
			Usd: model.CurrentPriceRate{
				Code:        in.Price.Bpi.Usd.Code,
				Symbol:      in.Price.Bpi.Usd.Symbol,
				Rate:        in.Price.Bpi.Usd.Rate,
				Description: in.Price.Bpi.Usd.Description,
				RateFloat:   in.Price.Bpi.Usd.RateFloat,
			},
			Gbp: model.CurrentPriceRate{
				Code:        in.Price.Bpi.Gbp.Code,
				Symbol:      in.Price.Bpi.Gbp.Symbol,
				Rate:        in.Price.Bpi.Gbp.Rate,
				Description: in.Price.Bpi.Gbp.Description,
				RateFloat:   in.Price.Bpi.Gbp.RateFloat,
			},
			Eur: model.CurrentPriceRate{
				Code:        in.Price.Bpi.Eur.Code,
				Symbol:      in.Price.Bpi.Eur.Symbol,
				Rate:        in.Price.Bpi.Eur.Rate,
				Description: in.Price.Bpi.Eur.Description,
				RateFloat:   in.Price.Bpi.Eur.RateFloat,
			},
		},
	}
}
