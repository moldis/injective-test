package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"code.injective.org/service/pricefetcher/internal/model"

	testdb "code.injective.org/service/pricefetcher/internal/repository/test_db"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
)

type PricesRepositorySuite struct {
	suite.Suite
	repository *prices
	db         *mongo.Database
	now        time.Time
	cleanups   []func()
}

func (suite *PricesRepositorySuite) SetupSuite() {
	db, cancel, err := testdb.NewMongoDB()
	if err != nil {
		suite.Failf("Error creating ClickHouse Client", fmt.Sprintf("%v", err))
	}
	suite.db = db
	suite.cleanups = append(suite.cleanups, cancel)
	repo := NewPrices(suite.db)
	suite.repository = repo
}

func (suite *PricesRepositorySuite) TestCreate() {
	ctx := context.Background()
	createdDate := time.Now()

	err := suite.repository.Create(ctx, &model.CurrentPrice{
		Time:       model.CurrentPriceTime{UpdatedISO: createdDate},
		Disclaimer: "disclamer",
		ChartName:  "chart_name",
		Bpi: model.CurrentPriceBpi{
			Usd: model.CurrentPriceRate{
				Code:        "USD",
				Symbol:      "USD",
				Rate:        "1000",
				Description: "desc",
				RateFloat:   1.000,
			},
			Gbp: model.CurrentPriceRate{
				Code:        "GBP",
				Symbol:      "GBP",
				Rate:        "1000",
				Description: "desc",
				RateFloat:   1.000,
			},
			Eur: model.CurrentPriceRate{
				Code:        "EUR",
				Symbol:      "EUR",
				Rate:        "1000",
				Description: "desc",
				RateFloat:   1.000,
			},
		},
	})
	suite.Assert().NoError(err)

	result, err := suite.repository.GetSinceDate(ctx, createdDate.Add(-1*time.Minute))
	suite.Assert().Len(result, 1)
	suite.Assert().NoError(err)
}

func (suite *PricesRepositorySuite) TearDownSuite() {
	for i := range suite.cleanups {
		suite.cleanups[i]()
	}
}

func TestPricesIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PricesRepositorySuite))
}
