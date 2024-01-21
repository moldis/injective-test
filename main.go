package main

import (
	"context"

	"code.injective.org/service/pricefetcher/internal/client"
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

	go client.RunPriceFetcher(ctx, coin, receiver, errors)

	srv, err := server.NewServer(receiver, errors, cfg)
	if err != nil {
		panic(err)
	}
	err = srv.Run()
	if err != nil {
		log.Err(err).Msg("server failed to start")
	}
}
