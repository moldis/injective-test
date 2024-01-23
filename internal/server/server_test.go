package server

import (
	"code.injective.org/service/pricefetcher/internal/config"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"code.injective.org/service/pricefetcher/internal/model"
	mockRepo "code.injective.org/service/pricefetcher/internal/repository/mocks"
	"github.com/gorilla/websocket"
)

func TestSuccessReceived(t *testing.T) {
	duration := time.Now().Add(10 * time.Second)

	// Create a context that is both manually cancellable and will signal
	// a cancel at the specified duration.
	_, cancel := context.WithDeadline(context.Background(), duration)
	defer cancel()

	receiver := make(chan *model.CurrentPrice)
	errors := make(chan error)
	defer close(receiver)
	defer close(errors)

	mockPrices := mockRepo.NewMockPrices(t)

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	srv, err := NewServer(receiver, errors, cfg, mockPrices)
	if err != nil {
		panic(err)
	}

	// run the actual server
	go func() {
		srv.Run()
	}()
	time.Sleep(5 * time.Second)

	ws, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:8080/ws", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	date := time.Unix(0, 10)
	samplePrice := &model.CurrentPrice{
		Time:       model.CurrentPriceTime{UpdatedISO: date},
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
	}

	receiver <- samplePrice

	msgType, msg, err := ws.ReadMessage()

	require.NoError(t, err)
	require.Equal(t, msgType, 1)
	require.Equal(t, string(msg), `{"timedate":"1970-01-01T08:00:00.00000001+08:00","price":1}`)

	sinceDate := time.Unix(1706015736, 10)
	samplePrice.Bpi.Usd.RateFloat = 2
	mockPrices.On("GetSinceDate", mock.Anything, sinceDate).Return([]*model.CurrentPrice{samplePrice}, nil).Once()

	wsClient2, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:8080/ws?since_date=1706015736", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	//second client must receive message
	msgType, msg, err = wsClient2.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, msgType, 1)
	require.Equal(t, string(msg), `{"timedate":"1970-01-01T08:00:00.00000001+08:00","price":2}`)

	receiver <- samplePrice
	// first client must receive message
	msgType, msg, err = ws.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, msgType, 1)
	require.Equal(t, string(msg), `{"timedate":"1970-01-01T08:00:00.00000001+08:00","price":2}`)

	msgType, msg, err = wsClient2.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, msgType, 1)
	require.Equal(t, string(msg), `{"timedate":"1970-01-01T08:00:00.00000001+08:00","price":2}`)

	// TODO tests with db and currencies
	wsClient3, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:8080/ws?currency=USD&currency=EUR", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	receiver <- samplePrice
	// first client must receive message
	msgType, msg, err = ws.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, msgType, 1)
	require.Equal(t, string(msg), `{"timedate":"1970-01-01T08:00:00.00000001+08:00","price":2}`)

	msgType, msg, err = wsClient2.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, msgType, 1)
	require.Equal(t, string(msg), `{"timedate":"1970-01-01T08:00:00.00000001+08:00","price":2}`)

	msgType, msg, err = wsClient3.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, msgType, 1)
	require.Equal(t, string(msg), `{"timedate":"1970-01-01T08:00:00.00000001+08:00","price":2,"price_usd":2,"price_eur":1}`)
}
