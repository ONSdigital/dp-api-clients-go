package upload

import (
	"encoding/json"
	"net/http"
)

type jsonError struct {
	Code        string `json:"errorCode"`
	Description string `json:"description"`
}

type jsonErrors struct {
	Errors []jsonError `json:"errors"`
}

func writeError(w http.ResponseWriter, errs jsonErrors, httpCode int) {
	encoder := json.NewEncoder(w)
	w.WriteHeader(httpCode)
	err := encoder.Encode(&errs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func buildErrors(err error, code string) jsonErrors {
	return jsonErrors{Errors: []jsonError{{Description: err.Error(), Code: code}}}
}
