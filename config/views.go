package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"sifeucr/internal/db"
)

var base = []string{
	"views/baseof.html",
	"views/_partials/head.html",
	"views/_partials/header.html",
	"views/_partials/footer.html",
}

var ViewMap = map[string][]string{
	"setup-page": append(
		base,
		"views/panel/setup.html",
		"views/panel/setup-page.html",
	),
	"setup-result": {
		"views/panel/setup-result.html",
	},
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
		"views/dashboard/purchases.html",
	),
	"dashboard": {
		"views/dashboard/dashboard.html",
		"views/dashboard/main-report.html",
		"views/dashboard/purchases.html",
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
		"views/panel/user-permissions.html",
		"views/panel/user-modal.html",
		"views/panel/permission.html",
		"views/panel/periods.html",
		"views/panel/period.html",
		"views/panel/period-update-form.html",
		"views/panel/distributions.html",
		"views/panel/distribution.html",
		"views/panel/dist-update-form.html",
		"views/panel/suppliers.html",
		"views/panel/supplier.html",
		"views/panel/supplier-update-form.html",
		"views/panel/catalogs.html",
		"views/panel/catalog.html",
		"views/panel/items.html",
		"views/panel/item.html",
		"views/panel/item-update-form.html",
	),
	"budget": {
		"views/panel/budget.html",
	},
	"user": {
		"views/panel/user.html",
		"views/panel/user-modal.html",
		"views/panel/user-permissions.html",
		"views/panel/permission.html",
	},
	"account": {
		"views/panel/account.html",
	},
	"permission": {
		"views/panel/permission.html",
	},
	"period": {
		"views/panel/period.html",
		"views/panel/period-update-form.html",
	},
	"period-update-form": {
		"views/panel/period-update-form.html",
	},
	"distribution": {
		"views/panel/distribution.html",
		"views/panel/dist-update-form.html",
	},
	"dist-update-form": {
		"views/panel/dist-update-form.html",
	},
	"supplier": {
		"views/panel/supplier.html",
	},
	"supplier-update-form": {
		"views/panel/supplier-update-form.html",
	},
	"catalog": {
		"views/panel/catalog.html",
	},
	"item": {
		"views/panel/item.html",
		"views/panel/item-update-form.html",
	},
	"item-update-form": {
		"views/panel/item-update-form.html",
	},

	"forms-purchase-form-page": append(
		base,
		"views/forms/purchase-form.html",
		"views/forms/purchase-form-page.html",
		"views/forms/purchase-form-generic.html",
		"views/forms/purchase-form-catering.html",
	),

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
	"summary": func(s string, max int) string {
		if max >= len(s) {
			return s
		}
		cut := s[:max]
		lastSpace := strings.LastIndex(cut, " ")
		if lastSpace == -1 {
			return cut + "..."
		}
		return s[:lastSpace] + "..."
	},
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
	"eq": func(a, b any) bool {
		switch va := a.(type) {
		case int:
			vb, ok := b.(int)
			return ok && va == vb
		case int64:
			vb, ok := b.(int64)
			return ok && va == vb
		case float64:
			vb, ok := b.(float64)
			return ok && va == vb
		case string:
			vb, ok := b.(string)
			return ok && va == vb
		default:
			return false
		}
	},
	"filterPermissionsByUser": func(perms []db.AllPermissionsRow, userID int64) []db.AllPermissionsRow {
		var out []db.AllPermissionsRow
		for _, p := range perms {
			if p.PermissionUser == userID {
				out = append(out, p)
			}
		}
		return out
	},
	"hasPermission": HasPermission,
	"dict": func(values ...any) (map[string]any, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call: uneven number of arguments")
		}

		dict := make(map[string]any, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings, got %T", values[i])
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
	"phone": func(countryCode, phone int64) string {
		cc := strconv.FormatInt(countryCode, 10)
		pn := fmt.Sprintf("%d", phone)

		switch cc {
		case "1": // US/Canada
			if len(pn) == 10 {
				return fmt.Sprintf("+1 (%s) %s-%s", pn[0:3], pn[3:6], pn[6:])
			}
		case "506": // Costa Rica
			return fmt.Sprintf("+506 %s-%s", pn[0:4], pn[4:])
		default:
			return fmt.Sprintf("+%s %s", cc, pn)
		}

		return pn
	},
	"formatID": formatID,
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

// Format as Costa Rican ID
func formatID(id int64) string {
	idStr := strconv.FormatInt(id, 10)

	switch len(idStr) {
	case 9:
		// Física: 0#-####-####
		return fmt.Sprintf("%s-%s-%s", string(idStr[0]), idStr[1:5], idStr[5:])
	case 10:
		// Jurídica/Gobierno Central/Inst. Autónoma: #-###-######
		return fmt.Sprintf("%s-%s-%s", idStr[0:1], idStr[1:4], idStr[4:])
	default:
		// Any other: return raw number
		return idStr
	}
}
