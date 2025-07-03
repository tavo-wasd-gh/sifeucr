package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"log"

	"github.com/tavo-wasd-gh/sifeucr/config"
	"git.tavo.one/tavo/axiom/sessions"
	"git.tavo.one/tavo/axiom/storage"
)

type Handler struct {
	cfg    Config
}

type Logger struct {
	Enabled bool
}

type Config struct {
	Production bool
	Logger      *Logger
	Views      map[string]*template.Template
	DB         *sql.DB
	S3         *storage.Client
	Sessions   *sessions.SessionStore[config.Session]
}

func New(cfg Config) *Handler {
	return &Handler{
		cfg: cfg,
	}
}

func (h *Handler) Production() bool {
	return h.cfg.Production
}

func (h *Handler) Views() map[string]*template.Template {
	return h.cfg.Views
}

func (h *Handler) DB() *sql.DB {
	return h.cfg.DB
}

func (h *Handler) S3() *storage.Client {
	return h.cfg.S3
}

func (h *Handler) Sessions() *sessions.SessionStore[config.Session] {
	return h.cfg.Sessions
}

func (h *Handler) Log() *Logger {
	return h.cfg.Logger
}

func (l *Logger) Error(format string, args ...any) {
	if !l.Enabled {
		return
	}

	pc, _, _, ok := runtime.Caller(1)
	funcName := "unknown"
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}

	msg := fmt.Sprintf(format, args...)

	log.Printf("%s: %s", funcName, msg)
}

func getCSRFToken(r *http.Request) string {
	if ct, ok := r.Context().Value(config.CSRFTokenKey).(string); ok {
		return ct
	}
	return ""
}

func getUserID(r *http.Request) int64 {
	if v, ok := r.Context().Value(config.UserIDKey).(int64); ok {
		return v
	}
	return 0
}

func getAccountID(r *http.Request) int64 {
	if v, ok := r.Context().Value(config.AccountIDKey).(int64); ok {
		return v
	}
	return 0
}
