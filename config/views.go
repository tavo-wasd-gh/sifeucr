package config

import (
	"encoding/json"
	"fmt"
	"html/template"
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
		"views/panel/period-modal.html",
		"views/panel/period-update-form.html",
		"views/panel/distributions.html",
		"views/panel/distribution.html",
		"views/panel/dist-modal.html",
		"views/panel/dist-update-form.html",
		"views/panel/suppliers.html",
		"views/panel/supplier.html",
		"views/panel/supplier-update-form.html",
		"views/panel/catalogs.html",
		"views/panel/catalog.html",
		"views/panel/items.html",
		"views/panel/item.html",
		"views/panel/item-modal.html",
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
		"views/panel/period-modal.html",
	},
	"period-update-form": {
		"views/panel/period-update-form.html",
	},
	"distribution": {
		"views/panel/distribution.html",
		"views/panel/dist-update-form.html",
		"views/panel/dist-modal.html",
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
		"views/panel/item-modal.html",
		"views/panel/item-update-form.html",
	},
	"item-update-form": {
		"views/panel/item-update-form.html",
	},
	"forms-purchase-registered": {
		"views/forms/purchase-registered.html",
	},
	"doc-justification": {
		"views/docs/justification.html",
	},
	"doc-quotation": {
		"views/docs/quotation.html",
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
	"opMult": func(a, b float64) float64 {
		return a * b
	},
	"opSum": func(nums ...float64) float64 {
		var total float64
		for _, n := range nums {
			total += n
		}
		return total
	},
	"pathToSVG": func(s string) template.HTML {
		sig, err := SignatureJSONToSVGText(s, 400, 300, 2, "#000")
		if err != nil {
			return ""
		}
		return template.HTML(sig)
	},
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
	"currency":      FormatAsCurrency,
	"unixDateToStr": unixDateToStr,
	"unixDateLong":  UnixDateLong,
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

func FormatAsCurrency(amount float64) string {
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

func UnixDateLong(timestamp int64) string {
	loc, _ := time.LoadLocation("America/Costa_Rica")
	t := time.Unix(timestamp, 0).In(loc)

	// Spanish names
	days := []string{"Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"}
	months := []string{
		"enero", "febrero", "marzo", "abril", "mayo", "junio",
		"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
	}

	dayName := days[t.Weekday()]
	monthName := months[t.Month()-1]

	// Hour in 12h format
	hour := t.Hour()
	ampm := "AM"
	if hour >= 12 {
		ampm = "PM"
	}
	hour12 := hour % 12
	if hour12 == 0 {
		hour12 = 12
	}

	return fmt.Sprintf("%d:%02d%s %s %d de %s, %d",
		hour12, t.Minute(), ampm,
		dayName, t.Day(), monthName, t.Year())
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

// SignatureJSONToSVGText converts the JSON-encoded normalized paths into a plain SVG string.
func SignatureJSONToSVGText(jsonStr string, width, height int, stroke float64, color string) (string, error) {
	var paths [][][]float64
	if err := json.Unmarshal([]byte(jsonStr), &paths); err != nil {
		return "", fmt.Errorf("parse signature json: %w", err)
	}
	return pathsToSVG(paths, width, height, stroke, color), nil
}

func pathsToSVG(paths [][][]float64, width, height int, stroke float64, color string) string {
	if color == "" {
		color = "black"
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`, width, height, width, height)
	for _, path := range paths {
		if len(path) == 0 {
			continue
		}
		b.WriteString(`<path fill="none"`)
		fmt.Fprintf(&b, ` stroke="%s" stroke-width="%g" stroke-linecap="round" stroke-linejoin="round"`, color, stroke)
		b.WriteString(` vector-effect="non-scaling-stroke" d="`)
		for i, pt := range path {
			if len(pt) != 2 {
				continue
			}
			x := pt[0] * float64(width)
			y := pt[1] * float64(height)
			if i == 0 {
				fmt.Fprintf(&b, "M%.3f %.3f", x, y)
			} else {
				fmt.Fprintf(&b, " L%.3f %.3f", x, y)
			}
		}
		b.WriteString(`"/>`)
	}
	b.WriteString(`</svg>`)
	return b.String()
}
