package config

import (
	"strings"
)

var base = []string{
	"min/views/baseof.html",
	"min/views/_partials/head.html",
	"min/views/_partials/head/fetch-and-swap.html",
	"min/views/_partials/header.html",
	"min/views/_partials/footer.html",
}

var ViewMap = map[string][]string{
	"login-page": append(
		base,
		"min/views/login.html",
		"min/views/login-page.html",
	),
	"login": {
		"min/views/login.html",
	},

	"dashboard-page": append(
		base,
		"min/views/dashboard.html",
		"min/views/dashboard-page.html",
		"min/views/control-madre.html",
	),
	"dashboard": {
		"min/views/dashboard.html",
		"min/views/control-madre.html",
	},

	"panel-page": append(
		base,
		"min/views/panel.html",
		"min/views/panel-page.html",
		"min/views/user.html",
		"min/views/users.html",
	),

	"index-page": append(
		base,
		"min/views/index.html",
		"min/views/index-page.html",
	),

	"suppliers-page": append(
		base,
		"min/views/suppliers.html",
		"min/views/suppliers-page.html",
	),

	"fse-page": append(
		base,
		"min/views/fse.html",
		"min/views/fse-page.html",
	),
}

var ViewFormatters = map[string]any{
	"uppercase": func(s string) string { return strings.ToUpper(s) },
	"firstWord": func(s string) string {
		words := strings.Fields(s)
		if len(words) > 0 {
			return words[0]
		}
		return ""
	},
}
