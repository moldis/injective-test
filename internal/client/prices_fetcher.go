package client

import (
	"context"
	"time"

	"code.injective.org/service/pricefetcher/internal/model"
	"code.injective.org/service/pricefetcher/internal/repository"
	"github.com/rs/zerolog/log"
)

type PriceProvider interface {
	GetPrice() (*model.CurrentPrice, error)
}

type PriceFetcher interface {
	RunPriceFetcher(ctx context.Context, fetcher PriceProvider, receiver chan *model.CurrentPrice, errors chan error, saveData bool)
}

type priceFetcher struct {
	pricesRepo     repository.Prices
	fetcher        PriceProvider
	tickerInterval int
	saveData       bool
}

func NewPriceFetcher(pricesRepo repository.Prices, fetcher PriceProvider, tickerInterval int, saveData bool) *priceFetcher {
	return &priceFetcher{
		pricesRepo:     pricesRepo,
		fetcher:        fetcher,
		tickerInterval: tickerInterval,
		saveData:       saveData,
	}
}

func (p *priceFetcher) RunPriceFetcher(ctx context.Context, receiver chan *model.CurrentPrice, errors chan error) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("exiting")
			return
		case <-ticker.C:
			price, err := p.fetcher.GetPrice()
			if err != nil {
				log.Err(err).Msgf("error receiving price %v", err)
				errors <- err
				continue
			}

			// we can run it in separate go routine
			log.Info().Msgf("received price %v", price)
			if p.saveData {
				err = p.pricesRepo.Create(ctx, price)
				if err != nil {
					log.Err(err).Msg("error saving to DB")
				}
			}

			// if nobody read it, it will stack, so using unblocking writes
			go func() {
				receiver <- price
			}()
		}
	}
}
