package app

import (
	"fmt"
	achttp "github.com/zhughes3/go-accelerate/internal/pkg/http"
	"time"
)

type Option func(*options)

type options struct {
	insecure bool
	errz     []error

	port int

	pprofEnabled   bool
	versionEnabled bool
	stateEnabled   bool

	prometheusEnabled bool

	customURL    string
	contextRoot  string
	timeout      time.Duration
	stateTimeout time.Duration

	authMiddleware achttp.HandlerMiddleware

	requestHandlerSpecs []achttp.RequestHandlerSpec

	version *Version

	beforeHooks []ShutdownErrorHook
	afterHooks  []ShutdownErrorHook
}

func (o *options) validate() error {
	if len(o.errz) > 0 {
		return fmt.Errorf("configuration errors: %v", o.errz)
	}

	return nil
}

func WithPort(p int) Option {
	return func(o *options) {
		o.port = p
	}
}

func WithPProfEnabled() Option {
	return func(o *options) {
		o.pprofEnabled = true
	}
}

func WithPrometheusEnabled() Option {
	return func(o *options) {
		o.prometheusEnabled = true
	}
}

func WithInsecure() Option {
	return func(o *options) {
		o.insecure = true
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.timeout = d
	}
}

func WithStateTimeout(d time.Duration) Option {
	return func(o *options) {
		o.stateTimeout = d
	}
}

func WithAuthMiddleware(mw achttp.HandlerMiddleware) Option {
	return func(o *options) {
		o.authMiddleware = mw
	}
}

func WithRequestHandlerSpecs(rhs []achttp.RequestHandlerSpec) Option {
	return func(o *options) {
		o.requestHandlerSpecs = append(o.requestHandlerSpecs, rhs...)
	}
}

func WithBeforeShutdownHook(h ShutdownHook) Option {
	return WithBeforeShutdownErrorHook(shutdownHookAdapter(h))
}

func WithBeforeShutdownErrorHook(h ShutdownErrorHook) Option {
	return func(o *options) {
		o.beforeHooks = append(o.beforeHooks, h)
	}
}

func WithAfterShutdownHook(h ShutdownHook) Option {
	return WithAfterShutdownErrorHook(shutdownHookAdapter(h))
}

func WithAfterShutdownErrorHook(h ShutdownErrorHook) Option {
	return func(o *options) {
		o.afterHooks = append(o.afterHooks, h)
	}
}
