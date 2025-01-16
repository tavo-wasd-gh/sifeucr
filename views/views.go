package views

import (
	"fmt"
	"database/sql"
	"html/template"
	"strings"
)

func Init(templatePaths map[string]string) (map[string]*template.Template, error) {
	funcMap := template.FuncMap{
		"frac": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			frac := 100 * (a / b)
			if frac < 0 {
				return 0
			}
			return frac
		},
		"currency": func(amount float64) string {
			formatted := fmt.Sprintf("%.2f", amount)
			parts := strings.Split(formatted, ".")
			intPart := parts[0]
			decPart := parts[1]

			isNegative := strings.HasPrefix(intPart, "-")
			if isNegative {
				intPart = intPart[1:]
			}

			var result strings.Builder
			length := len(intPart)
			for i, digit := range intPart {
				if i > 0 && (length-i)%3 == 0 {
					result.WriteString(".")
				}
				result.WriteRune(digit)
			}

			if isNegative {
				return "-" + result.String() + "," + decPart
			}
			return "₡" + result.String() + "," + decPart
		},
		"calcularPeriodo": func(a, b sql.NullTime) bool {
			if !a.Valid || !b.Valid {
				return false
			}
			return a.Time.Before(b.Time)
		},
		"eq": func(a, b string) bool {
			return a == b
		},
		"sub": func(a, b float64) float64 {
			return a - b
		},
		"sum": func(a, b float64) float64 {
			return a + b
		},
	}

	views := make(map[string]*template.Template)
	for name, path := range templatePaths {
		tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(path)
		if err != nil {
			return nil, err
		}
		views[name] = tmpl
	}

	return views, nil
}
