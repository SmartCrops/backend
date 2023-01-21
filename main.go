package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/SmartCrops/backend/mqtt"
)

const (
	brokerURL  = "tcp://172.111.242.63:6666"
	username   = "roslina"
	password   = "smartcrops"
	waterTopic = "command/69"
)

func run() error {
	mqttClient, err := mqtt.Connect(brokerURL, username, password)
	if err != nil {
		return fmt.Errorf("failed to connect to the mqtt broker: %w", err)
	}
	log.Println("Connected to the broker!")

	http.HandleFunc("/water", func(w http.ResponseWriter, r *http.Request) {
		payload := `{"pumpGpio":0, "durationS":5}`
		err = mqttClient.Pub(waterTopic, 1, false, payload)
		if err != nil {
			log.Println(err)
		}
		_, _ = w.Write([]byte("Ok"))
	})
	return http.ListenAndServe(":8080", nil)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
