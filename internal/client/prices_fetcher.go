package client

import (
	"code.injective.org/service/pricefetcher/internal/model"
	"context"
	"time"
)

type PriceFetcher interface {
	GetPrice() (*model.CurrentPrice, error)
}

func RunPriceFetcher(ctx context.Context,
	fetcher PriceFetcher,
	receiver chan *model.CurrentPrice,
	errors chan error) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			price, err := fetcher.GetPrice()
			if err != nil {
				errors <- err
				continue
			}
			receiver <- price
		}
	}
}
