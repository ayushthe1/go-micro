package main

import (
	"context"
	"log"
	"logger-service/data"
	"time"
)

type RPCServer struct{}

// define the kind of payload that we're going to receive from RPC.
// kind of data we're going to receive for any methods that are tied to RPCServer
type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Println("error writing to mongo", err)
		return err
	}

	*resp = "Processed payload via RPC:" + payload.Name
	return nil

}
