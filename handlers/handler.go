package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"runtime"

	"git.tavo.one/tavo/axiom/sessions"
	"git.tavo.one/tavo/axiom/storage"

	"sifeucr/config"
	"sifeucr/internal/db"
)

type Handler struct {
	cfg Config
}

type Logger struct {
	Enabled bool
}

type Config struct {
	Production bool
	Logger     *Logger
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

func getCSRFTokenFromContext(ctx context.Context) string {
	if ct, ok := ctx.Value(config.CSRFTokenKey).(string); ok {
		return ct
	}
	return ""
}

func getUserIDFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(config.UserIDKey).(int64); ok {
		return v
	}
	return 0
}

func getAccountIDFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(config.AccountIDKey).(int64); ok {
		return v
	}
	return 0
}

func (h *Handler) checkPermissionFromContext(ctx context.Context, required int64) error {
	userID := getUserIDFromContext(ctx)
	accountID := getAccountIDFromContext(ctx)
	if userID == 0 || accountID == 0 {
		return fmt.Errorf("cannot addBudgetEntry: invalid data")
	}

	queries := db.New(h.DB())
	perm, err := queries.PermissionByUserIDAndAccountID(ctx, db.PermissionByUserIDAndAccountIDParams{
		UserID:    userID,
		AccountID: accountID,
	})
	if err != nil {
		return fmt.Errorf("failed to query permissions: %v")
	}

	if !config.HasPermission(perm.PermissionInteger, required) {
		return fmt.Errorf("incorrect permissions, got:%d want:%d", perm.PermissionInteger, required)
	}

	return nil
}
