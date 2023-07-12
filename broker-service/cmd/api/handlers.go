package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

// HandleSubmission is the main point of entry into the broker. It accepts a JSON
// payload and performs an action based on the value of "action" in that JSON.
func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		// app.logItem(w, requestPayload.Log)
		// app.logEventViaRabbit(w, requestPayload.Log)
		app.LogItemViaRPC(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)

}

// authenticate calls the authentication microservice and sends back the appropriate response
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the service (This is done form inside the docker container)
	log.Println("Sending POST request to authentication-service from broker")

	// Since we don't specify a port, it calls on default port 80 (inside docker)
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	// create a variable we'll read response.Body into
	var jsonFromService jsonResponse

	// decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticateddd!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.Marshal(msg)

	// call the mail service
	mailServiceURL := "http://mailer-service/send"

	// post to mail service
	// buffer stores data as a sequence of bytes. The bytes.Buffer type in the standard library treats the data as a byte slice ([]byte).when you read data from a buffer, you receive a byte slice ([]byte). Buffer is like a  holding area that holds a certain amount of data until it is processed or transferred elsewhere.
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the right status code
	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}

	// send back json
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + msg.To

	app.writeJSON(w, http.StatusAccepted, payload)

}

// Different function to handle logging an item and we'll do so by emitting an event to Rabbit MQ
func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	app.writeJSON(w, http.StatusAccepted, payload)
}

// we'll use this function every time we need to push something to the queue
func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.Marshal(&payload)
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

// log event via rpc
func (app *Config) LogItemViaRPC(w http.ResponseWriter, l LogPayload) {

	// get rpc client ;logger-service is our service name in docker compose
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// After we have the client, we need to create some kind of payload. We need to create a type that exactly matches the one that the remote RPCServer expects to get

	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	// Any method that i want to expose to rpc on the serverend must be exported. So it has to start with a capital letter
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// write some json back to the user
	payload := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	log.Println("THIS IS THE REQUESTpayload :", requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close() // defer closing it or we'll have connections left open all over the places and it's a resource leak

	// create a client
	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)
}

// package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"errors"
// 	"net/http"
// )

// type RequestPayload struct {
// 	Action string `json:"action"`
// 	// we'll create a different type for each of the possible action. (eg: auth ,mail ,log)
// 	Auth AuthPayload `json:"auth, omitempty"`
// }

// type AuthPayload struct {
// 	// fields we need to authenticate
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }

// // define a handler and add it to the routes file
// // Note that function Broker starts with capital B as it's required to use this function in other files.
// func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
// 	payload := jsonResponse{
// 		Error:   false,
// 		Message: "Hit the broker babyy",
// 	}

// 	// // Write this data (payload) out
// 	// // json.Marshal is used to convert a Go data structure (such as a struct, map, or slice) into its corresponding JSON representation.It takes a Go value as input and returns a byte slice containing the JSON-encoded data.
// 	// out, _ := json.Marshal(payload)
// 	// w.Header().Set("Content-Type", "application/json")
// 	// w.WriteHeader(http.StatusAccepted)

// 	// // w.Write() function is used to write the response body(w). out is a byte slice containing the JSON-encoded data that will be sent as the response body.
// 	// w.Write(out)

// 	// OR

// 	_ = app.writeJSON(w, http.StatusOK, payload)

// }

// func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {

// 	// read into variable of type request payload
// 	var requestPayload RequestPayload

// 	err := app.readJSON(w, r, &requestPayload)
// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}

// 	switch requestPayload.Action {
// 	case "auth":
// 		app.authenticate(w, requestPayload.Auth)
// 	default:
// 		app.errorJSON(w, errors.New("unknown action"))
// 	}
// }

// func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
// 	// create some json that we'll send to the auth microservice
// 	jsonData, _ := json.MarshalIndent(a, "", "\t")

// 	// call the service
// 	request, err := http.NewRequest("POST", "http://localhost:8081/authenticate", bytes.NewBuffer(jsonData))

// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}

// 	client := &http.Client{}
// 	response, err := client.Do(request)
// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}
// 	defer response.Body.Close()

// 	// make sure we get back the correct status code
// 	if response.StatusCode == http.StatusUnauthorized {
// 		app.errorJSON(w, errors.New("Invalid credentials (in handlers.go broker)"))
// 		return
// 	} else if response.StatusCode != http.StatusAccepted {
// 		app.errorJSON(w, errors.New("error calling auth service (in handlers.go broker)"))
// 		return
// 	}

// 	// create a variable we'll read response.Body into (this response is coming from authentication service)
// 	var jsonFromService jsonResponse

// 	// decode the json from the auth service
// 	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}

// 	if jsonFromService.Error {
// 		app.errorJSON(w, err, http.StatusUnauthorized)
// 		return
// 	}

// 	// what we want to send back to the user
// 	var payload jsonResponse
// 	payload.Error = false
// 	payload.Message = "Authenticated oooribabaaa "
// 	payload.Data = jsonFromService.Data

// 	app.writeJSON(w, http.StatusAccepted, payload)
// }
