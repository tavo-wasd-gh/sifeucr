package config

import (
	"strings"
)

var ViewMap = map[string]string{
	"dashboard": "templates/dashboard.html",
	"login":     "templates/login.html",
}

var FuncMap = map[string]interface{}{
	"uppercase": func(s string) string { return strings.ToUpper(s) },
}
