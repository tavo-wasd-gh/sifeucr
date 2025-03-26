package config

import (
	"strings"
)

var ViewMap = map[string]string{
	"dashboard": "templates/dashboard.html",
	"login":     "templates/login.html",
	"setup":     "templates/setup.html",
}

var FuncMap = map[string]interface{}{
	"uppercase": func(s string) string { return strings.ToUpper(s) },
}
