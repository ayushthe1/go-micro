// We'll write a function called writeLog() that will allow us to communicate over gRPC

package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"logger-service/logs"
	"net"

	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer             // This field is going to be required for every service we ever write over gRPC (to ensure backward compatibility)
	Models                             data.Models // have access to necessary methods to write to Mongo

}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry() // gets the input
	log.Println("This is the input we get : ", input)

	// write the log
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Result: "failed response coming from WriteLog func"}
		return res, err
	}

	// return response if nothing went wrong
	res := &logs.LogResponse{Result: "logged! response coming from WriteLog func"}
	return res, nil

}

// function to listen for gRPC connections (will start the gRPC listener)
func (app *Config) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC %v", err)
	}

	s := grpc.NewServer()

	// register the service
	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	log.Printf("gRPC Server started on port %s", gRpcPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC %v", err)
	}
}
