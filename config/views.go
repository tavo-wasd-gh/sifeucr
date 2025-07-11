package config

import (
	"fmt"
	"strings"
	"time"
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
		"views/dashboard/dashboard.html",
		"views/dashboard/dashboard-page.html",
		"views/dashboard/main-report.html",
	),
	"dashboard": {
		"views/dashboard/dashboard.html",
		"views/dashboard/main-report.html",
	},

	"panel-page": append(
		base,
		"views/panel/panel.html",
		"views/panel/panel-page.html",
		"views/panel/budget.html",
		"views/panel/budgets.html",
		"views/panel/user.html",
		"views/panel/users.html",
		"views/panel/account.html",
		"views/panel/accounts.html",
		"views/panel/periods.html",
		"views/panel/period.html",
		"views/panel/period-update-form.html",
		"views/panel/distributions.html",
		"views/panel/distribution.html",
		"views/panel/dist-update-form.html",
	),
	"budget": {
		"views/panel/budget.html",
	},
	"user": {
		"views/panel/user.html",
	},
	"account": {
		"views/panel/account.html",
	},
	"period": {
		"views/panel/period.html",
	},
	"period-update-form": {
		"views/panel/period-update-form.html",
	},
	"distribution": {
		"views/panel/distribution.html",
	},
	"dist-update-form": {
		"views/panel/dist-update-form.html",
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
	"currency":      formatAsCurrency,
	"unixDateToStr": unixDateToStr,
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

// Show dates as UTC, because shifting n hours backwards will result in the date
// changing to a day before. Convert to local time when checking validity.
func unixDateToStr(timestamp int64) string {
	t := time.Unix(timestamp, 0).UTC()
	return t.Format("2006-01-02")
}
