package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type jsonResponse struct {
	Error   bool   `json:"errorrr"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// function to read json
func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // 1 MB

	// limiting the size of incoming request bodies
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	// decode a JSON value into an empty struct{}.
	err = dec.Decode(&struct{}{})
	// io.EOF is an error value indicating the end of the input source has been reached.
	// If err is not equal to io.EOF, it means that the decoding process did not reach the end of the input source, and there are additional JSON values remaining.
	if err != io.EOF {
		return errors.New("body must have only a single json value")
	}

	return nil
}

// function to write json
//  The ellipsis (...) before http.Header indicates that it can accept a variable number of arguments, allowing you to pass or store multiple http.Header values.
func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// function to write error message as json
func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}
