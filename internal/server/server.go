package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"code.injective.org/service/pricefetcher/internal/config"
	"code.injective.org/service/pricefetcher/internal/model"
	"code.injective.org/service/pricefetcher/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const queueBufferSize = 100

type Server struct {
	receiverCh chan *model.CurrentPrice
	errorCh    chan error
	cfg        *config.Config
	wsUpgrader websocket.Upgrader
	repo       repository.Prices
	mutex      sync.Mutex
	multi      map[string]chan *model.CurrentPrice
}

func NewServer(
	receiverCh chan *model.CurrentPrice,
	errorCh chan error,
	cfg *config.Config,
	repo repository.Prices) (Server, error) {

	wsUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	multi := make(map[string]chan *model.CurrentPrice)
	return Server{receiverCh: receiverCh, errorCh: errorCh, cfg: cfg, wsUpgrader: wsUpgrader, repo: repo, multi: multi}, nil
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// reading income messages
	go func(conn *websocket.Conn) {
		for {
			// Read a message from the client.
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			// Print the message to the console.
			fmt.Println("Received:", message)
			fmt.Println("Received messageType:", messageType)
		}
	}(conn)

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

	if r.URL.Query().Has("since_date") {
		since := r.URL.Query().Get("since_date")
		d, err := strconv.Atoi(since)
		if err != nil {
			log.Err(err).Msgf("wrong since date, ignoring")
			conn.Close()
			return
		}

		legacy, err := s.repo.GetSinceDate(context.Background(), time.Unix(int64(d), 10))
		if err != nil {
			log.Err(err).Msgf("something wrong with legacy data %v", err)
		}
		for _, rate := range legacy {
			err = s.sendReceivedPrice(conn, rate)
			if err != nil {
				log.Err(err).Msgf("something wrong with connections %v", err)
				conn.Close()
				return
			}
		}
	}

	for {
		select {
		case rate := <-queueMsg:
			log.Info().Msgf("received rate %v", rate)
			err = s.sendReceivedPrice(conn, rate)
		case errMsg := <-s.errorCh:
			log.Err(errMsg).Msgf("something goes wrong with channel")
		}
	}
}

func (s *Server) sendReceivedPrice(conn *websocket.Conn, rate *model.CurrentPrice) error {
	type PriceMsg struct {
		TimeDate time.Time `json:"timedate"`
		Price    float64   `json:"price"`
	}
	message := PriceMsg{
		TimeDate: rate.Time.UpdatedISO,
		Price:    rate.Bpi.Usd.RateFloat,
	}
	marshalled, err := json.Marshal(message)
	if err != nil {
		log.Err(err).Msg("error marshaling structure")
	}

	err = conn.WriteMessage(1, marshalled)
	if err != nil {
		log.Err(err).Msg("error writing message to client")
		return err
	}
	return nil
}

func (s *Server) runPipe() {
	for {
		select {
		case rate := <-s.receiverCh:
			log.Info().Msgf("received rate %v", rate)
			for s2 := range s.multi {
				s.multi[s2] <- rate
			}
		case errMsg := <-s.errorCh:
			log.Err(errMsg).Msgf("something goes wrong with channel")
		}
	}
}

func (s *Server) Run() error {
	http.HandleFunc("/ws", s.wsHandler)
	log.Info().Msgf("server started on %s", s.cfg.Listen)

	go s.runPipe()

	return http.ListenAndServe(s.cfg.Listen, nil)
}
