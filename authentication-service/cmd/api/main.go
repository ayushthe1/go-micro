package main

import (
	"authentication/data"

	// Package sql provides a generic interface around SQL (or SQL-like) databases.The sql package must be used in conjunction with a database driver
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	//  "pgx" Go driver is a software component that acts as a bridge between your Go application and the PostgreSQL database server. It provides a set of tools and functions that make it easy for your Go code to send requests and receive responses from the database. It simplifies the process of integrating a database into your application, making it easier to store and retrieve data as needed.

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Broker is also listening on port 8084 and we can make authentication service also listen on that port as docker lets multiple containers listen on same port and treats them as indiviual servers

type Config struct {
	DB     *sql.DB
	Models data.Models
}

// authentication service will listen on port 80 inside of docker.
// you can have multiple services listening on the same port inside docker, just as if they were separate machines. So every single web/api service inside of docker can all listen on port 80.
const webPort = "80"

var counts int64

func main() {
	log.Println("Starting authentication service")

	// connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	// set up config
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for 2 sec ...")
		time.Sleep(2 * time.Second)
		continue
	}
}
