package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	acerrors "github.com/zhughes3/go-accelerate/pkg/errors"
	acstrings "github.com/zhughes3/go-accelerate/pkg/strings"
	"io"
	"net/http"
	"strings"
)

func RequiredParamValue(r *http.Request, paramName string) (string, error) {
	v, found := ParamValue(r, paramName)
	if !found {
		return v, acerrors.NewInvalidInputErrorf("Missing required request parameter: '%s'", paramName)
	}

	return v, nil
}

func ParamValue(r *http.Request, paramName string) (string, bool) {
	if v, found := pathParamValue(r, paramName); found {
		return v, found
	}

	return urlParamValue(r, paramName)
}

func paramValues(r *http.Request, paramName string) ([]string, bool) {
	if v, found := pathParamValue(r, paramName); found {
		return []string{v}, found
	}

	return urlParamValues(r, paramName)
}

func pathParamValue(r *http.Request, paramName string) (string, bool) {
	v := chi.URLParam(r, paramName)

	return v, acstrings.IsNotBlank(v)
}

func urlParamValues(r *http.Request, paramName string) ([]string, bool) {
	values, found := r.URL.Query()[paramName]

	return values, found
}

func urlParamValue(r *http.Request, paramName string) (string, bool) {
	v := r.URL.Query().Get(paramName)

	return v, acstrings.IsNotBlank(v)
}

func DecodeJSONRequestBody(r *http.Request, v any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		if err == io.EOF {
			return acerrors.NewInvalidInputError("JSON must be provided in the request body")
		}

		return acerrors.NewInvalidInputErrorf("Could not parse JSON in the request body: %s", err.Error())
	}

	return nil
}

func ClientIP(r *http.Request) string {
	header := r.Header.Get("X-Forwarded-For")
	ipAddList := strings.Split(header, ",")
	if len(ipAddList) == 1 && len(ipAddList[0]) == 0 {
		return r.RemoteAddr
	}

	return ipAddList[0]
}
