package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/SmartCrops/backend/mqtt"
)

/* -------------------------------------------------------------------------- */
/* --------------------------------- Server --------------------------------- */
/* -------------------------------------------------------------------------- */

type Server struct {
	DB              *sql.DB
	MQTT            mqtt.Client
	OptimalHumidity *float32 // this should use a mutex
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/data", wrapHand(s.handleGetData))
	mux.HandleFunc("/set-humi", wrapHandB(s.handleSetOptimalHumidity))
	mux.HandleFunc("/water", wrapHandB(s.handleWater))
	mux.HandleFunc("/echo", wrapHandB(s.handleEcho))

	return http.ListenAndServe(addr, cors(mux))
}

/* -------------------------------------------------------------------------- */
/* -------------------------------- Handlers -------------------------------- */
/* -------------------------------------------------------------------------- */

func (s *Server) handleEcho(msg string) (string, int, error) {
	return msg, http.StatusOK, nil
}

func (s *Server) handleGetData() ([]Meas, int, error) {
	data, err := getMeasurements(s.DB)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return data, http.StatusOK, err
}

func (s *Server) handleSetOptimalHumidity(value float32) (string, int, error) {
	*s.OptimalHumidity = value
	return "Ok", http.StatusOK, nil
}

func (s *Server) handleWater(seconds uint) (string, int, error) {
	err := sendWaterCommand(s.MQTT, seconds)
	if err != nil {
		err = fmt.Errorf("failed to send the watering command: %w", err)
		return "", http.StatusInternalServerError, err
	}
	return "Ok", http.StatusOK, nil
}

/* -------------------------------------------------------------------------- */
/* ----------------------------- CORS Middleware ---------------------------- */
/* -------------------------------------------------------------------------- */

func cors(hand http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		hand.ServeHTTP(w, r)
	})
}
