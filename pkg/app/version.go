package app

import (
	"fmt"
	achttp "github.com/zhughes3/go-accelerate/internal/pkg/http"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	"net/http"
)

var (
	buildtime      = "N/A"
	compiler       = "N/A"
	commit         = "N/A"
	version        = "N/A"
	DefaultVersion *Version
)

type Version struct {
	// BuildTime is the time of the build
	BuildTime string
	// Compiler is the version of the Go runtime
	Compiler string
	// Commit is the commit sha of the build
	Commit string
	// Version is the semantic version of the build
	Version string
}

func (v Version) Map() map[string]any {
	return map[string]any{
		"BuildTime": v.BuildTime,
		"Compiler":  v.Compiler,
		"Commit":    v.Commit,
		"Version":   v.Version,
	}
}

func (v Version) GetVersionHandler(logger slog.Logger) http.Handler {
	return versionHandler{
		logger:  logger,
		version: &v,
	}
}

func init() {
	DefaultVersion = &Version{
		BuildTime: buildtime,
		Compiler:  compiler,
		Commit:    commit,
		Version:   version,
	}
}

type versionHandler struct {
	logger  slog.Logger
	version *Version
}

func (h versionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := checkGetMethodAndSetContentHeader(w, r); err != nil {
		h.logger.WithError(err).ErrorContext(r.Context(), "Problem getting version")
		return
	}

	_ = achttp.EncodeJSONResponse(r.Context(), w, h.version)
}

func checkGetMethodAndSetContentHeader(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		handlerErrorResponse(w, r, http.StatusMethodNotAllowed, "Method Not Allowed")
		return fmt.Errorf("method %s not allowed", r.Method)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return nil
}

func handlerErrorResponse(w http.ResponseWriter, _ *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
