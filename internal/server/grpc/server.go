package grpc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"code.injective.org/service/pricefetcher/internal/config"
	"code.injective.org/service/pricefetcher/internal/model"
	"code.injective.org/service/pricefetcher/internal/repository"
	pb "code.injective.org/service/pricefetcher/proto/prices"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const queueBufferSize = 100

type PricesServer struct {
	pb.UnimplementedPricesStreamingServiceServer

	receiverCh chan *model.CurrentPrice
	errorCh    chan error
	cfg        *config.Config
	repo       repository.Prices
	mutex      sync.Mutex
	multi      map[string]chan *model.CurrentPrice
}

func NewPricesServer(
	receiverCh chan *model.CurrentPrice,
	errorCh chan error,
	cfg *config.Config,
	repo repository.Prices) *PricesServer {
	multi := make(map[string]chan *model.CurrentPrice)

	return &PricesServer{receiverCh: receiverCh, errorCh: errorCh, cfg: cfg, repo: repo, multi: multi}
}

func (s PricesServer) GetDataStreaming(req *pb.PricesRequest, srv pb.PricesStreamingService_GetDataStreamingServer) error {
	// buffered channel to keep prices in queue
	queueMsg := make(chan *model.CurrentPrice, queueBufferSize)
	defer func() {
		s.mutex.Lock()
		delete(s.multi, uuid.NewString())
		s.mutex.Unlock()
		close(queueMsg)
	}()

	s.mutex.Lock()
	s.multi[uuid.NewString()] = queueMsg
	s.mutex.Unlock()

	currency := req.GetCurrency()
	if req.GetSinceDate() != 0 {
		legacy, err := s.repo.GetSinceDate(context.Background(), time.Unix(int64(req.GetSinceDate()), 10))
		if err != nil {
			log.Err(err).Msgf("something wrong with legacy data %v", err)
		}
		for _, rate := range legacy {
			err = s.sendReceivedPrice(srv, rate, currency)
			if err != nil {
				log.Err(err).Msgf("something wrong with connections %v", err)
			}
		}
	}

	for {
		select {
		case rate := <-queueMsg:
			log.Info().Msgf("received rate %v", rate)
			err := s.sendReceivedPrice(srv, rate, currency)
			if err != nil {
				log.Err(err).Msgf("error sending pricing data")
			}
		case errMsg := <-s.errorCh:
			if errMsg != nil {
				log.Err(errMsg).Msgf("something goes wrong with channel %s", errMsg)
			}
		}
	}

	return nil
}

func (s PricesServer) sendReceivedPrice(conn pb.PricesStreamingService_GetDataStreamingServer,
	rate *model.CurrentPrice, currency []string) error {
	res := &pb.PricesResponse{
		TimeDate: rate.Time.UpdatedISO.Unix(),
		Price:    fmt.Sprintf("%f", rate.Bpi.Usd.RateFloat),
	}

	for _, cur := range currency {
		if strings.EqualFold(cur, "usd") {
			res.PriceUsd = fmt.Sprintf("%f", rate.Bpi.Usd.RateFloat)
		}

		if strings.EqualFold(cur, "eur") {
			res.PriceEur = fmt.Sprintf("%f", rate.Bpi.Eur.RateFloat)
		}

		if strings.EqualFold(cur, "gbp") {
			res.PriceGbp = fmt.Sprintf("%f", rate.Bpi.Gbp.RateFloat)
		}
	}
	return conn.Send(res)
}
