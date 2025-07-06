package config

import (
	"fmt"
	"strings"
)

var base = []string{
	"views/baseof.html",
	"views/_partials/head.html",
	"views/_partials/header.html",
	"views/_partials/footer.html",
}

var ViewMap = map[string][]string{
	"login-page": append(
		base,
		"views/login.html",
		"views/login-page.html",
	),
	"login": {
		"views/login.html",
	},

	"dashboard-page": append(
		base,
		"views/dashboard.html",
		"views/dashboard-page.html",
		"views/control-madre.html",
	),
	"dashboard": {
		"views/dashboard.html",
		"views/control-madre.html",
	},

	"panel-page": append(
		base,
		"views/panel.html",
		"views/panel-page.html",
		"views/budget.html",
		"views/budgets.html",
		"views/user.html",
		"views/users.html",
	),
	"budget": {
		"views/budget.html",
	},
	"user": {
		"views/user.html",
	},

	"index-page": append(
		base,
		"views/index.html",
		"views/index-page.html",
	),

	"suppliers-page": append(
		base,
		"views/suppliers.html",
		"views/suppliers-page.html",
	),

	"fse-page": append(
		base,
		"views/fse.html",
		"views/fse-page.html",
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
	"currency": formatAsCurrency,
}

func formatAsCurrency(amount float64) string {
    sign := ""
    if amount < 0 {
        sign = "-"
        amount = -amount
    }

    formatted := fmt.Sprintf("%.2f", amount)
    parts := strings.Split(formatted, ".")

    integerPart := parts[0]
    decimalPart := parts[1]

    var result []byte
    for i, digit := range integerPart {
        if (len(integerPart)-i)%3 == 0 && i != 0 {
            result = append(result, ',')
        }
        result = append(result, byte(digit))
    }

    return fmt.Sprintf("₡%s%s.%s", sign, string(result), decimalPart)
}
