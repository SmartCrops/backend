package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/SmartCrops/backend/mqtt"
	_ "github.com/go-sql-driver/mysql"
)

func run() error {
	var optimHumi float32 = 50.0 // Optimal humidity
	var lastMeasId string = ""

	/* ----------------------------- Conenct to MQTT ---------------------------- */
	mqttClient, err := mqtt.Connect("tcp://172.111.242.63:6666", "roslina", "smartcrops")
	if err != nil {
		return fmt.Errorf("failed to connect to the mqtt broker: %w", err)
	}
	log.Println("Connected to the broker!")

	/* ------------------------------ Conenct to DB ----------------------------- */
	db, err := sql.Open("mysql", "root:smartcrops@tcp(172.111.242.63:3306)/PBL5_meas")
	if err != nil {
		return fmt.Errorf("failed to open a db connection: %w", err)
	}
	defer db.Close()

	/* ----------------------- Start checking the humidity ---------------------- */
	go func() {
		for {
			time.Sleep(time.Second)

			lastMeasId = adjustHumidity(lastMeasId, db, mqttClient, optimHumi)
			log.Println(lastMeasId)
		}
	}()

	/* -------------------------- Start the http server ------------------------- */
	myServer := Server{
		DB:              db,
		MQTT:            mqttClient,
		OptimalHumidity: &optimHumi,
	}
	return myServer.Start(":8080")
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
