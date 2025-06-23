package config

import (
	"strings"
)

var ViewMap = map[string][]string{
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

	"login": {
		"templates/login.html",
	},

	"dashboard-page": {
		"templates/baseof.html",
		"templates/_partials/head.html",
		"templates/_partials/header.html",
		"templates/dashboard.html",
		"templates/dashboard-page.html",
		"templates/_partials/footer.html",
		// Modules
		"templates/resumen-cuentas.html",
	},

	"dashboard": {
		"templates/dashboard.html",
		// Modules
		"templates/resumen-cuentas.html",
	},

	"suppliers-page": {
		"templates/baseof.html",
		"templates/_partials/head.html",
		"templates/_partials/header.html",
		"templates/suppliers.html",
		"templates/suppliers-page.html",
		"templates/_partials/footer.html",
	},

	"fse-page": {
		"templates/baseof.html",
		"templates/_partials/head.html",
		"templates/_partials/header.html",
		"templates/fse.html",
		"templates/fse-page.html",
		"templates/_partials/footer.html",
	},

	"resumen-cuentas": {
		"templates/resumen-cuentas.html",
	},
}

var FuncMap = map[string]interface{}{
	"uppercase": func(s string) string { return strings.ToUpper(s) },
	"firstWord": func(s string) string {
		words := strings.Fields(s)
		if len(words) > 0 {
			return words[0]
		}
		return ""
	},
}
