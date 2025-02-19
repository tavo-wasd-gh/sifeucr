package config

import (
	"time"
)

// Default cookie duration
func CookieTimeout() time.Time {
	return time.Now().Add(1*time.Hour)
}
