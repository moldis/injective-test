package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"code.injective.org/service/pricefetcher/internal/config"
	"code.injective.org/service/pricefetcher/internal/model"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Server struct {
	receiverCh chan *model.CurrentPrice
	errorCh    chan error
	cfg        *config.Config
	wsUpgrader websocket.Upgrader
}

func NewServer(
	receiverCh chan *model.CurrentPrice,
	errorCh chan error,
	cfg *config.Config) (Server, error) {

	wsUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return Server{receiverCh: receiverCh, errorCh: errorCh, cfg: cfg, wsUpgrader: wsUpgrader}, nil
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

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

	for {
		select {
		case rate := <-s.receiverCh:
			log.Info().Msgf("received rate %v", rate)

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
				return
			}
		case errMsg := <-s.errorCh:
			log.Err(errMsg).Msgf("something goes wrong with channel")
		}
	}
}

func (s *Server) Run() error {
	http.HandleFunc("/ws", s.wsHandler)
	log.Info().Msgf("server started on %s", s.cfg.Listen)

	return http.ListenAndServe(s.cfg.Listen, nil)
}
