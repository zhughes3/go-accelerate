package http

import "net/http"

type HandlerMiddleware func(handler http.Handler) http.Handler
