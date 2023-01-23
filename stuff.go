package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SmartCrops/backend/mqtt"
)

/* -------------------------------------------------------------------------- */
/* ------------------------------ Measurements ------------------------------ */
/* -------------------------------------------------------------------------- */

type Meas struct {
	Light uint
	Preas float32
	Humi  float32 // Percents
	Temp  float32
	Time  string
}

func getMeasurements(db *sql.DB) ([]Meas, error) {
	results, err := db.Query(
		"SELECT Light_Intensity, Pressure, Soil_Humidity, Temperature, ts FROM Measurements ORDER BY ts DESC LIMIT 10",
	)
	if err != nil {
		return nil, err
	}
	var measurements []Meas
	for results.Next() {
		var meas Meas
		err = results.Scan(&meas.Light, &meas.Preas, &meas.Humi, &meas.Temp, &meas.Time)
		if err != nil {
			return nil, err
		}
		measurements = append([]Meas{meas}, measurements...)
	}
	return measurements, nil
}

/* -------------------------------------------------------------------------- */
/* ----------------------------- Humidity Stuff ----------------------------- */
/* -------------------------------------------------------------------------- */

func sendWaterCommand(client mqtt.Client, seconds uint) error {
	topic := "command/69"
	payload := fmt.Sprintf(`{"pumpGpio":0, "durationS":%d}`, seconds)
	return client.Pub(topic, 1, false, payload)
}

func adjustHumidity(db *sql.DB, client mqtt.Client, optimHumi float32) {
	measurements, err := getMeasurements(db)
	if err != nil {
		log.Println(err)
		return
	}
	currHumi := measurements[len(measurements)-1].Humi
	if currHumi >= optimHumi {
		return
	}

	secs := uint(optimHumi - currHumi)
	err = sendWaterCommand(client, secs)
	if err != nil {
		log.Println(err)
	}
}
