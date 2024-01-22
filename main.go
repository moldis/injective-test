package main

import (
	"code.injective.org/service/pricefetcher/internal/client"
	"code.injective.org/service/pricefetcher/internal/repository"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"code.injective.org/service/pricefetcher/internal/client/provider"
	"code.injective.org/service/pricefetcher/internal/config"
	"code.injective.org/service/pricefetcher/internal/model"
	"code.injective.org/service/pricefetcher/internal/server"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	coin := provider.NewCoinDeskProvider("BTC")

	receiver := make(chan *model.CurrentPrice)
	errors := make(chan error)
	defer close(receiver)
	defer close(errors)

	// setup DB connection
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDBURL))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = mongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	err = mongoClient.Ping(ctx, &readpref.ReadPref{})
	if err != nil {
		panic(err)
	}

	db := mongoClient.Database(cfg.DBName)

	// run fetcher to receive prices
	pricesRepo := repository.NewPrices(db)
	fetcher := client.NewPriceFetcher(pricesRepo, coin, cfg.FetchInterval)
	go fetcher.RunPriceFetcher(ctx, receiver, errors)

	// run ws
	srv, err := server.NewServer(receiver, errors, cfg, pricesRepo)
	if err != nil {
		panic(err)
	}
	err = srv.Run()
	if err != nil {
		log.Err(err).Msg("server failed to start")
	}
}
