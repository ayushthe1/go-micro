package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

// Receiver that we use for the application
type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting broker service on port %s \n", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backoff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready
	for {
		// rabbitmq matches our service in docker compose
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = c
			break
		}

		// If we didn't connect :

		// if we didn't connect after 5 tries
		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		// if we haven't tried atleast 5 times ,then call backoff
		backoff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off....")
		// suspends the execution of the loop for the calculated duration.
		time.Sleep(backoff)
		continue
	}

	return connection, nil
}
