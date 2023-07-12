package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// Register the RPC Server
	err = rpc.Register(new(RPCServer))

	go app.rpcListen()

	// Listen for gRPC connection
	go app.gRPCListen()

	// start web server
	log.Println("Starting service on port", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic()
	}

}

// func (app *Config) serve() {
// 	srv := &http.Server{
// 		Addr: fmt.Sprintf(":%s", webPort),
// 		Handler: app.routes(),
// 	}

// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		log.Panic()
// 	}
// }

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// connect
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}

	log.Println("Connected to mongo!")

	return c, nil
}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port ", rpcPort)

	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}

	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeConn(rpcConn)
	}
}

// //  Think of a context as a container that holds important information for a particular operation. It carries things like request-scoped values ( authentication tokens, request IDs,etc), deadlines (time limits) and cancellation signals. It helps manage and control how long an operation should take and allows for stopping or canceling it if needed.

// Imagine you have a program with multiple tasks running at the same time. Each task might need to know how much time it has left to complete or if it should stop because something went wrong. That's where the context package comes in.

// Here are some key things to understand about the context package:

// Context: A context is like a box that holds important information related to a specific task or operation. It includes details like deadlines (time limits) and cancellation signals.

// Cancelling: Sometimes, an operation may take too long, or you might want to stop it for some reason. With the context package, you can create a cancellation signal and pass it along with the context. When the signal is triggered, all the related tasks can gracefully stop what they're doing.

// Timeouts: You can set deadlines or time limits for tasks using the context package. If a task doesn't finish within the specified time, it can be automatically canceled. This helps ensure that your program doesn't get stuck waiting forever for a task to complete.

// Passing Values: The context package also allows you to pass values between different parts of your program. For example, you might want to pass an authentication token or a request ID across different function calls or goroutines. The context can hold and carry these values for you.

// By using the context package, you can better manage and control concurrent operations, timeouts, cancellations, and the exchange of values in your Go programs. It provides a structured way to handle these common challenges and helps make your programs more reliable and responsive.

// The context package in Go allows you to store and access the request scoped values throughout the execution flow of the request.

// Request-scoped values are data or information that is associated with a specific HTTP request in a web application. They are values that are relevant and specific to a particular request and should be accessible across different parts of the code that handle the request.

// Incoming requests to a server should create a Context, and outgoing calls to servers should accept a Context
