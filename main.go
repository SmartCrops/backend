package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/SmartCrops/backend/mqtt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/cors"
)

const (
	brokerURL  = "tcp://172.111.242.63:6666"
	username   = "roslina"
	password   = "smartcrops"
	waterTopic = "command/69"
)

type Meas struct {
	Temp float32
	Time string
}

func getData(db *sql.DB) ([]Meas, error) {
	results, err := db.Query("SELECT Temperature, ts FROM Measurements ORDER BY ts DESC LIMIT 10")
	if err != nil {
		return nil, err
	}
	var measurements []Meas
	for results.Next() {
		var meas Meas
		err = results.Scan(&meas.Temp, &meas.Time)
		if err != nil {
			return nil, err
		}
		measurements = append([]Meas{meas}, measurements...)
	}
	return measurements, nil
}

func run() error {
	mqttClient, err := mqtt.Connect(brokerURL, username, password)
	if err != nil {
		return fmt.Errorf("failed to connect to the mqtt broker: %w", err)
	}
	log.Println("Connected to the broker!")

	// Create the database handle, confirm driver is present
	db, _ := sql.Open("mysql", "root:smartcrops@/PBL5_meas")
	defer db.Close()

	// Connect and check the server version
	var version string
	db.QueryRow("SELECT VERSION()").Scan(&version)
	fmt.Println("Connected to:", version)

	// var measurements []float32
	// err = db.QueryRow("SELECT Temperature FROM Measurements").Scan(&measurements)
	// if err != nil {
	// 	return fmt.Errorf("failed to query the database: %w", err)
	// }
	// log.Println("Got some measurements", len(measurements))
	// log.Println(measurements)

	mux := http.NewServeMux()
	mux.HandleFunc("/water", func(w http.ResponseWriter, r *http.Request) {
		payload := `{"pumpGpio":0, "durationS":5}`
		err = mqttClient.Pub(waterTopic, 1, false, payload)
		if err != nil {
			log.Println(err)
		}
		_, _ = w.Write([]byte("Ok"))
	})
	mux.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		measurements, err := getData(db)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		b, _ := json.Marshal(measurements)
		w.Write(b)
	})
	handler := cors.Default().Handler(mux)
	return http.ListenAndServe(":8080", handler)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
