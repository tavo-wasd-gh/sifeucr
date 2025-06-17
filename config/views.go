package config

import (
	"strings"
)

var ViewMap = map[string][]string{
	"index": {
		"templates/baseof.html",
		"templates/_partials/head.html",
		"templates/_partials/header.html",
		"templates/index.html",
		"templates/_partials/footer.html",
	},

	"dashboard": {
		"templates/baseof.html",
		"templates/_partials/head.html",
		"templates/_partials/header.html",
		"templates/dashboard.html",
		"templates/_partials/footer.html",
	},
}

var FuncMap = map[string]interface{}{
	"uppercase": func(s string) string { return strings.ToUpper(s) },
}
