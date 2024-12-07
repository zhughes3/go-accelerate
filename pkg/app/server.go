package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zhughes3/go-accelerate/pkg/app/state"
	acerrors "github.com/zhughes3/go-accelerate/pkg/errors"
	"github.com/zhughes3/go-accelerate/pkg/slices"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	acstrings "github.com/zhughes3/go-accelerate/pkg/strings"
	acurl "github.com/zhughes3/go-accelerate/pkg/url"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

const defaultOpContextRoot = "/app"

const (
	pathMetrics      = "/metrics"
	pathPprof        = "/pprof"
	pathPprofCmdLine = "/cmdline"
	pathPprofProfile = "/profile"
	pathPprofSymbol  = "/symbol"
	pathPprofTrace   = "/trace"
	pathState        = "/state"
	pathVersion      = "/version"
)

// A Server is a generic HTTP server where you can mound arbitrary HTTP handlers.
type Server struct {
	httpServer *http.Server
	logger     slog.Logger

	rwMu               sync.RWMutex
	state              state.State
	nextStateTimestamp time.Time     // time until calculating state again
	stateTimeout       time.Duration // time in seconds to wait to calculate state again

	beforeHooks []ShutdownErrorHook
	afterHooks  []ShutdownErrorHook

	prometheusEnabled bool
}

func NewServer(logger slog.Logger, opts ...Option) (*Server, error) {
	cfg := options{}
	for _, o := range opts {
		o(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, acerrors.Wrap(err, "invalid configuration")
	}

	port := ":6161"
	if cfg.port > 0 {
		port = fmt.Sprintf(":%d", cfg.port)
	}

	if cfg.version == nil {
		cfg.version = DefaultVersion
	}
	logger.WithFields(cfg.version.Map()).InfoContext(context.Background(), "Version")

	router := chi.NewRouter()

	contextRoot := acurl.CreateFullPath("/", cfg.contextRoot)

	if len(cfg.requestHandlerSpecs) > 0 {
		appRouter := NewRouterBuilder(logger).WithAuthMiddleware(&cfg.authMiddleware).
			WithRequestHandlerSpecs(cfg.requestHandlerSpecs).Build()

		logger.InfoContextf(context.Background(), "Adding app-specific HTTP handlers to a chi router mounted at %s", contextRoot)
		router.Mount(contextRoot, appRouter)
	}

	timeout := 16 * time.Second
	if cfg.timeout > 0 {
		timeout = cfg.timeout
	}

	server := &Server{
		httpServer: &http.Server{
			Addr:           port,
			Handler:        router,
			ReadTimeout:    timeout,
			WriteTimeout:   timeout,
			MaxHeaderBytes: 1 << 20,
		},
		logger:             logger,
		rwMu:               sync.RWMutex{},
		state:              state.Starting,
		nextStateTimestamp: time.Now(),
		beforeHooks:        cfg.beforeHooks,
		afterHooks:         cfg.afterHooks,
		prometheusEnabled:  cfg.prometheusEnabled,
	}

	registerOperationalHandlers(logger, cfg, router, server)

	logRoutes(logger, router)

	return server, nil
}

func logRoutes(logger slog.Logger, router *chi.Mux) {
	_ = chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.With("method", method).With("path", route).DebugContextf(context.Background(), "Route: %s %s with %d middleware", method, route, len(middlewares))
		return nil
	})
}

func (s *Server) Run(ctx context.Context) error {
	logger := s.logger.With("Address", s.httpServer.Addr)
	logger.InfoContext(ctx, "HTTP App Server starting")

	go s.handleSignals(ctx)

	err := s.httpServer.ListenAndServe()

	// ListenAndServe will always return a non-nil error
	if !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).ErrorContext(ctx, "Unexpected error from HTTP server")
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.WarnContext(ctx, "Starting shutdown")
	defer s.logger.WarnContext(ctx, "Shutdown complete")

	var errs shutdownErrors
	if hookErrs := runHooks(ctx, s.getBeforeHooks()); len(hookErrs) > 0 {
		errs = append(errs, hookErrs...)
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	if hookErrs := runHooks(ctx, s.getAfterHooks()); len(hookErrs) > 0 {
		errs = append(errs, hookErrs...)
	}

	if errs.hasErrors() {
		return fmt.Errorf("problem occurred while shutting down server: %w", errs)
	}

	return nil
}

func (s *Server) getAfterHooks() (hooks []ShutdownErrorHook) {
	s.rwMu.RLock()
	hooks = s.afterHooks
	s.rwMu.RUnlock()
	return
}

func (s *Server) getBeforeHooks() (hooks []ShutdownErrorHook) {
	s.rwMu.RLock()
	hooks = s.beforeHooks
	s.rwMu.RUnlock()
	return
}

func (s *Server) handleSignals(ctx context.Context) {
	s.logger.InfoContext(ctx, "Setting up server SigTerm listener")

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTSTP)

	// block until signal is received
	x := <-ch
	s.logger.WithFields(map[string]any{
		"Signal":              x,
		"NumActiveGoroutines": runtime.NumGoroutine(),
	}).WarnContext(ctx, "Received OS Signal")

	if err := s.Shutdown(ctx); err != nil {
		s.logger.WithError(err).ErrorContext(ctx, "Error occurred while shutting down server")
	}
}

func (s *Server) RegisterBeforeShutdownErrorHook(h ShutdownErrorHook) {
	s.rwMu.Lock()
	s.beforeHooks = append(s.beforeHooks, h)
	s.rwMu.Unlock()
}

func (s *Server) getStateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := checkGetMethodAndSetContentHeader(w, r); err != nil {
			s.logger.WithError(err).ErrorContext(ctx, "problem checking method and setting header")
			return
		}
		resp := s.getStateResponse(ctx)
		w.WriteHeader(stateToHTTPStatus(resp.State))
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.WithError(err).ErrorContext(ctx, "problem encoding state handler response")
			handlerErrorResponse(w, r, http.StatusInternalServerError, "Error processing state")
		}
	})
}

func stateToHTTPStatus(newState state.State) int {
	if newState != state.Ready {
		return http.StatusServiceUnavailable
	}

	return http.StatusOK
}

type stateResponse struct {
	State state.State `json:"state"`
}

func (s *Server) getStateResponse(ctx context.Context) stateResponse {
	return stateResponse{State: s.GetState(ctx)}
}

func (s *Server) GetState(ctx context.Context) state.State {
	s.rwMu.Lock()
	if s.nextStateTimestamp.After(time.Now()) {
		s.logger.WarnContext(ctx, "High amount of traffic, using stored state")
		return s.state
	}

	s.state = state.Ready
	s.nextStateTimestamp = time.Now().Add(s.stateTimeout)
	s.rwMu.Unlock()
	return s.state
}

func registerOperationalHandlers(logger slog.Logger, cfg options, router *chi.Mux, s *Server) {
	ctx := context.Background()
	opContextRoot := defaultOpContextRoot

	if acstrings.IsNotBlank(cfg.customURL) {
		opContextRoot = acurl.CreateFullPath("/", cfg.customURL)
	}

	// Set Version endpoint
	logger.InfoContextf(ctx, "Using version endpoint at '%s'", opContextRoot+pathVersion)
	router.Method(http.MethodGet, opContextRoot+pathVersion, cfg.version.GetVersionHandler(logger))

	// Set State endpoint
	logger.InfoContextf(ctx, "Using state endpoint at '%s'", opContextRoot+pathState)
	s.stateTimeout = 5 * time.Second
	if cfg.stateTimeout != 0 {
		s.stateTimeout = cfg.stateTimeout
	}
	router.Method(http.MethodGet, opContextRoot+pathState, s.getStateHandler())

	if cfg.pprofEnabled {
		pprofContextRoot := opContextRoot + pathPprof
		logger.InfoContextf(context.Background(), "Using pprof endpoint at %s", pprofContextRoot)
		router.Get(pprofContextRoot+"*", pprofIndexOverride(pprofContextRoot))
		router.Get(pprofContextRoot+pathPprofCmdLine, pprof.Cmdline)
		router.Get(pprofContextRoot+pathPprofProfile, pprof.Profile)
		router.Get(pprofContextRoot+pathPprofSymbol, pprof.Symbol)
		router.Get(pprofContextRoot+pathPprofTrace, pprof.Trace)
	}

	// TODO set up prometheus, tracing
}

func pprofIndexOverride(path string) http.HandlerFunc {
	pathWithTrailingSlash := strings.TrimRight(path, "/") + "/"
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, pathWithTrailingSlash) {
			name := strings.TrimPrefix(r.URL.Path, pathWithTrailingSlash)
			if acstrings.IsNotBlank(name) {
				pprof.Handler(name).ServeHTTP(w, r)
				return
			}
			pprof.Index(w, r)
		}
	}
}

func runHooks(ctx context.Context, hooks []ShutdownErrorHook) []error {
	var errs []error
	for _, hook := range hooks {
		if err := hook(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

type shutdownErrors []error

func (e shutdownErrors) hasErrors() bool {
	return len(e) > 0
}

func (e shutdownErrors) Error() string {
	msgs := slices.Map(e, func(e error) string {
		return e.Error()
	})
	return strings.Join(msgs, "; ")
}
