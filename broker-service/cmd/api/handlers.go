package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
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
		app.logItem(w, requestPayload.Log)
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
