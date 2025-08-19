package config

import (
	"time"
)

const (
	TokenLength     = 64
	MaxSessions     = 500
	SessionMaxAge   = 2 * time.Hour
	SessionTokenKey = "session"
	CookieMaxAge    = 60 * 30
)

type contextKey string

const (
	CSRFTokenKey contextKey = "csrf_token"
	UserIDKey    contextKey = "user_id"
	AccountIDKey contextKey = "account_id"
)

type Session struct {
	UserID    int64
	AccountID int64
}
