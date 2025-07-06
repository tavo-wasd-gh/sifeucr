package config

import (
	"time"
)

const (
	TokenLength     = 64
	MaxSessions     = 1000
	SessionMaxAge   = 5 * time.Hour
	SessionTokenKey = "session"
	CookieMaxAge    = 40 * 60
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
