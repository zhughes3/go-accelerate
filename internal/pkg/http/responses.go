package http

import (
	"context"
	"encoding/json"
	"errors"
	acerrors "github.com/zhughes3/go-accelerate/pkg/errors"
	"net/http"
)

const (
	HeaderContentType = "Content-Type"
)

const (
	MediaTypeAppJSON = "application/json; charset=UTF-8"
)

type Non2XxResponse struct {
	StatusCode int    `json"-"`
	Message    string `json:"message,omitempty"`
}

type EncodeErrorFunc func(context.Context, error, http.ResponseWriter)

func DefaultErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	statusCode := http.StatusInternalServerError

	var statusCoder acerrors.StatusCoder
	if errors.As(err, &statusCoder) {
		statusCode = statusCoder.StatusCode()
	}

	EncodeNon2XxResponse(w, statusCode, err.Error())
}

func EncodeNon2XxResponse(w http.ResponseWriter, code int, message string) {
	_ = encodeJSONResponse(w, code, Non2XxResponse{Message: message})
}

// EncodeResponseFunc is responsible for encoding generic responses to the [http.ResponseWriter]
type EncodeResponseFunc func(context.Context, http.ResponseWriter, any) error

func EncodeJSONResponse(_ context.Context, w http.ResponseWriter, resp any) error {
	return encodeJSONResponse(w, http.StatusOK, resp)
}

func EncodeJSONCreatedResponse(_ context.Context, w http.ResponseWriter, resp any) error {
	return encodeJSONResponse(w, http.StatusCreated, resp)
}

func EncodeNoContentResponse(_ context.Context, w http.ResponseWriter, _ any) error {
	w.WriteHeader(http.StatusNoContent)

	return nil
}

func encodeJSONResponse(w http.ResponseWriter, status int, resp any) error {
	w.Header().Set(HeaderContentType, MediaTypeAppJSON)
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(resp)
}
