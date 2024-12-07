package http

import (
	"context"
	acendpoint "github.com/zhughes3/go-accelerate/internal/pkg/endpoint"
	"net/http"
)

type ErrorHandler interface {
	Handle(ctx context.Context, err error)
}

type ErrorHandlerFunc func(context.Context, error)

func (f ErrorHandlerFunc) Handle(ctx context.Context, err error) {
	f(ctx, err)
}

type RequestHandlerSpec interface {
	// Name provides an identifier for the spec.
	Name() string
	// Decoder is a func that transforms an [http.Request] into the input to the [Endpoint].
	Decoder() DecodeRequestFunc
	// Endpoint is a func that encapsulates the main operation of the request handler.
	Endpoint() acendpoint.Endpoint
	// Encoder is a func that writes the result of the [endpoint.Endpoint] to an [http.ResponseWriter].
	Encoder() EncodeResponseFunc
	// Path is the full path that the router should use to route requests to this request handler.
	Path() string
	// Methods are the HTTP methods to which this request handler should respond.
	Methods() []string
}

// A DecodeRequestFunc is a func responsible for decoding the parameters and/or body
// of the given [http.Request] to a custom request type.
type DecodeRequestFunc func(ctx context.Context, r *http.Request) (any, error)

type handlerSpec struct {
	name     string
	decoder  DecodeRequestFunc
	endpoint acendpoint.Endpoint
	encoder  EncodeResponseFunc
	path     string
	methods  []string
}

func NewRequestHandlerSpec(name string, dec DecodeRequestFunc, e acendpoint.Endpoint,
	enc EncodeResponseFunc, path string, methods []string) RequestHandlerSpec {
	return handlerSpec{
		name:     name,
		decoder:  dec,
		endpoint: e,
		encoder:  enc,
		path:     path,
		methods:  methods,
	}
}

func (h handlerSpec) Name() string {
	return h.name
}

func (h handlerSpec) Decoder() DecodeRequestFunc {
	return h.decoder
}

func (h handlerSpec) Endpoint() acendpoint.Endpoint {
	return h.endpoint
}

func (h handlerSpec) Encoder() EncodeResponseFunc {
	return h.encoder
}

func (h handlerSpec) Path() string {
	return h.path
}

func (h handlerSpec) Methods() []string {
	return h.methods
}

type requestHandler struct {
	endpoint     acendpoint.Endpoint
	decoder      DecodeRequestFunc
	encoder      EncodeResponseFunc
	errorEncoder EncodeErrorFunc
	errorHandler ErrorHandler
}

func NewHandler(ep acendpoint.Endpoint, decoder DecodeRequestFunc, encoder EncodeResponseFunc) http.Handler {
	return requestHandler{
		endpoint:     ep,
		decoder:      decoder,
		encoder:      encoder,
		errorEncoder: DefaultErrorEncoder,
		// TODO add logging error handler
		errorHandler: ErrorHandlerFunc(func(_ context.Context, _ error) {}),
	}
}

func (h requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := h.decoder(ctx, r)
	if err != nil {
		h.errorHandler.Handle(ctx, err)
		h.errorEncoder(ctx, err, w)
		return
	}

	resp, err := h.endpoint(ctx, request)
	if err != nil {
		h.errorHandler.Handle(ctx, err)
		h.errorEncoder(ctx, err, w)
		return
	}

	if err := h.encoder(ctx, w, resp); err != nil {
		h.errorHandler.Handle(ctx, err)
		h.errorEncoder(ctx, err, w)
	}
}
