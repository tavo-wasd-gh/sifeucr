package views

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

func Init(viewMap map[string]string) (map[string]*template.Template, error) {
	funcMap := template.FuncMap{
		"frac":     frac,
		"currency": currency,
		"eq":       eq,
		"gt":       gt,
		"lt":       lt,
		"sub":      sub,
		"sum":      sum,
		"datetime": datetime,
		"summary":  summary,
	}

	views := make(map[string]*template.Template)
	for name, path := range viewMap {
		tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(path)
		if err != nil {
			return nil, err
		}
		views[name] = tmpl
	}

	return views, nil
}

func frac(a, b float64) float64 {
	if b == 0 {
		return 0
	}

	frac := 100 * (a / b)
	if frac < 0 {
		return 0
	}

	return frac
}

func currency(amount float64) string {
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
}

func eq(a, b interface{}) bool {
	switch a := a.(type) {
	case time.Time:
		b, ok := b.(time.Time)
		if !ok {
			return false
		}
		return a.Equal(b)
	case float64:
		b, ok := b.(float64)
		if !ok {
			return false
		}
		return a == b
	case int:
		b, ok := b.(int)
		if !ok {
			return false
		}
		return a == b
	default:
		return false
	}
}

func gt(a, b interface{}) bool {
	switch a := a.(type) {
	case time.Time:
		b, ok := b.(time.Time)
		if !ok {
			return false
		}
		return a.After(b)
	case float64:
		b, ok := b.(float64)
		if !ok {
			return false
		}
		return a > b
	case int:
		b, ok := b.(int)
		if !ok {
			return false
		}
		return a > b
	default:
		return false
	}
}

func lt(a, b interface{}) bool {
	switch a := a.(type) {
	case time.Time:
		b, ok := b.(time.Time)
		if !ok {
			return false
		}
		return a.Before(b)
	case float64:
		b, ok := b.(float64)
		if !ok {
			return false
		}
		return a < b
	case int:
		b, ok := b.(int)
		if !ok {
			return false
		}
		return a < b
	default:
		return false
	}
}

func sub(a, b interface{}) float64 {
	switch a := a.(type) {
	case float64:
		b, ok := b.(float64)
		if !ok {
			return 0
		}
		return a - b
	case int:
		b, ok := b.(int)
		if !ok {
			return 0
		}
		return float64(a - b)
	default:
		return 0
	}
}

func sum(a, b interface{}) float64 {
	switch a := a.(type) {
	case float64:
		b, ok := b.(float64)
		if !ok {
			return 0
		}
		return a + b
	case int:
		b, ok := b.(int)
		if !ok {
			return 0
		}
		return float64(a + b)
	default:
		return 0
	}
}

func datetime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}

	months := []string{
		"", "enero", "febrero", "marzo", "abril", "mayo", "junio",
		"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
	}
	days := []string{
		"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado",
	}

	dayName := days[t.Weekday()]
	monthName := months[int(t.Month())]

	return fmt.Sprintf("%s, %02d de %s de %d, %02d:%02d",
		dayName,
		t.Day(),
		monthName,
		t.Year(),
		t.Hour(),
		t.Minute(),
	)
}

func summary(str string, length int) string {
	if len(str) > length {
		return str[:length] + "..."
	}
	return str
}
