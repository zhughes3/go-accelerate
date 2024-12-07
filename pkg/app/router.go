package app

import (
	"github.com/go-chi/chi/v5"
	achttp "github.com/zhughes3/go-accelerate/internal/pkg/http"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	"net/http"
)

type RouterBuilder struct {
	logger slog.Logger

	router *chi.Mux

	serviceName string

	handlerMiddlewares []achttp.HandlerMiddleware
	authMiddleware     *achttp.HandlerMiddleware

	requestHandlerSpecs []achttp.RequestHandlerSpec
}

func NewRouterBuilder(logger slog.Logger) *RouterBuilder {
	return &RouterBuilder{logger: logger}
}

func (b *RouterBuilder) WithRouter(r *chi.Mux) *RouterBuilder {
	b.router = r
	return b
}

func (b *RouterBuilder) WithAuthMiddleware(mw *achttp.HandlerMiddleware) *RouterBuilder {
	b.authMiddleware = mw
	return b
}

func (b *RouterBuilder) WithAuthMiddlewareFunc(mw func(http.Handler) http.Handler) *RouterBuilder {
	_mw := achttp.HandlerMiddleware(mw)
	return b.WithAuthMiddleware(&_mw)
}

func (b *RouterBuilder) WithHandlerMiddlewares(mws []achttp.HandlerMiddleware) *RouterBuilder {
	b.handlerMiddlewares = append(b.handlerMiddlewares, mws...)
	return b
}

func (b *RouterBuilder) WithHandleMiddlewareFuncs(mws []func(http.Handler) http.Handler) *RouterBuilder {
	for _, mw := range mws {
		b.handlerMiddlewares = append(b.handlerMiddlewares, mw)
	}

	return b
}

func (b *RouterBuilder) WithRequestHandlerSpecs(specs []achttp.RequestHandlerSpec) *RouterBuilder {
	b.requestHandlerSpecs = append(b.requestHandlerSpecs, specs...)
	return b
}

func (b *RouterBuilder) Build() *chi.Mux {
	if b.router == nil {
		b.router = chi.NewRouter()
	}

	if b.authMiddleware != nil {
		b.router.Use(*b.authMiddleware)
	}

	for _, mw := range b.handlerMiddlewares {
		b.router.Use(mw)
	}

	b.router.Use(achttp.NewLoggingMiddleware(b.logger))
	b.router.Use(achttp.NewRecoveryMiddleware(b.logger, achttp.DefaultErrorEncoder))

	for _, spec := range b.requestHandlerSpecs {
		ep := spec.Endpoint()
		enc := spec.Encoder()
		handler := achttp.NewHandler(ep, spec.Decoder(), enc)

		// TODO prometheus and trace instrumentation

		for _, method := range spec.Methods() {
			b.router.Method(method, spec.Path(), handler)
		}
	}

	return b.router
}
