package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrInternalServer    = errors.New("internal error")
	ErrBadRequest        = errors.New("bad request")
	ErrNotFound          = errors.New("not found")
	ErrForbidden         = errors.New("forbidden")
	ErrConflict          = errors.New("conflict")
	ErrEmailAlreadyExist = errors.New("email already exist")
)

type errorBlock struct {
	Name string
	Code int
	Msg  string
}

type errorEnvelope struct {
	Error errorBlock
}

func HandleError(w http.ResponseWriter, statusCode int, message string, name error) {

	response := errorEnvelope{
		Error: errorBlock{
			Name: name.Error(),
			Code: statusCode,
			Msg:  message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
