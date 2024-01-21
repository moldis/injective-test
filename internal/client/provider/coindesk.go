package provider

import (
	"code.injective.org/service/pricefetcher/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type coinDeskProvider struct {
	coin string
}

func NewCoinDeskProvider(coin string) *coinDeskProvider {
	return &coinDeskProvider{coin: coin}
}

func (p *coinDeskProvider) GetPrice() (*model.CurrentPrice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	url := "https://api.coindesk.com/v1/bpi/currentprice.json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("provider returned wrong status code %d", res.StatusCode)
	}

	var currentPrice model.CurrentPrice
	err = json.NewDecoder(res.Body).Decode(&currentPrice)
	if err != nil {
		return nil, err
	}
	return &currentPrice, nil
}
