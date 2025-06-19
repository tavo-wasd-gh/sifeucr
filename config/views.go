package config

import (
	"strings"
)

var ViewMap = map[string][]string{
	// -----
	// PAGES
	// -----

	"index-page": {
		"templates/baseof.html",
		"templates/_partials/head.html",
		"templates/_partials/header.html",
		"templates/index.html",
		"templates/index-page.html",
		"templates/_partials/footer.html",
	},

	"login-page": {
		"templates/baseof.html",
		"templates/_partials/head.html",
		"templates/_partials/header.html",
		"templates/login.html",
		"templates/login-page.html",
		"templates/_partials/footer.html",
	},

	"dashboard-page": {
		"templates/baseof.html",
		"templates/_partials/head.html",
		"templates/_partials/header.html",
		"templates/dashboard.html",
		"templates/dashboard-page.html",
		"templates/_partials/footer.html",
	},

	// ---------
	// FRAGMENTS
	// ---------

	"login": {
		"templates/login.html",
	},

	"dashboard": {
		"templates/dashboard.html",
	},
}

var FuncMap = map[string]interface{}{
	"uppercase": func(s string) string { return strings.ToUpper(s) },
}
