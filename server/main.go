package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tavo-wasd-gh/gocors"
	// "github.com/tavo-wasd-gh/gosmtp"
)

func main() {
	var (
		port   = os.Getenv("PORT")
		// db_uri = os.Getenv("DB_URI")
	)

	if port == "" {
		log.Fatalf("Fatal: Missing env variables")
	}

	if err := initializeDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/api/dashboard", handleDashboard)
	http.Handle("/", http.FileServer(http.Dir("public")))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Log: Running on :" + port + "...")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Fatal: Failed to start on port %s: %v", port, err)
		}
	}()

	<-stop

	log.Println("Log: Shutting down...")
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	var id_cuenta string
	var err error

	time.Sleep(100 * time.Millisecond)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		correo := r.FormValue("correo")
		// passwd := r.FormValue("passwd")
		cuenta_pedida := r.FormValue("id_cuenta")

		cuentas, err := cuentasAcreditadas(correo)
		if err != nil {
			http.Error(w, "Failed to validate user", http.StatusUnauthorized)
			return
		}

		// s := smtp.Client("smtp.ucr.ac.cr", "587", passwd)
		// if err := s.Validate(correo); err != nil {
		// 	http.Error(w, "Failed to validate email", http.StatusUnauthorized)
		// 	return
		// }

		if cuenta_pedida != "" {
			for _, cuenta := range cuentas {
				if cuenta.IDCuenta == id_cuenta {
					id_cuenta = cuenta_pedida
					break
				}
			}
		} else {
			if len(cuentas) == 1 {
				id_cuenta = cuentas[0].IDCuenta
			} else if len(cuentas) > 1 {
				// TODO: Login con las cuentas acreaditadas
				view(w, "views/login.html", nil)
				return
			} else {
				http.Error(w, "Failed to validate user", http.StatusUnauthorized)
				return
			}
		}

		if err := jwtSet(w, "jwt_token", id_cuenta, time.Now().Add(15*time.Minute)); err != nil {
			http.Error(w, "Failed to set JWT", http.StatusInternalServerError)
			return
		}
	}

	if r.Method == http.MethodGet {
		id_cuenta, err = jwtValidate(r, "jwt_token")
		if err != nil {
			view(w, "views/login.html", nil)
			return
		}
	}

	data := &Data{}
	if err := fillData(data, id_cuenta); err != nil {
		log.Println(err)
		view(w, "views/login.html", nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	view(w, "views/dashboard.html", data)
	return
}

func view(w http.ResponseWriter, path string, data *Data) error {
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
		"calcularEmitido": func(tipo, periodo string) float64 {
			restante, err := calcularEmitido(data, tipo, periodo)
			if err != nil {
				return 0
			}
			return restante
		},
		"calcularPeriodo": func(a, b sql.NullTime) bool {
			return isBefore(a, b)
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

	file, err := os.ReadFile(path)
	if err != nil {
		log.Println("Error: Failed to read template file:", err)
		http.Error(w, "Failed to read template file", http.StatusInternalServerError)
		return err
	}

	tmpl, err := template.New("template").Funcs(funcMap).Parse(string(file))
	if err != nil {
		log.Println("Error: Failed to parse template:", err)
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return err
	}

	var filled bytes.Buffer
	if err := tmpl.Execute(&filled, data); err != nil {
		log.Println("Error: Failed to execute template:", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return err
	}

	_, err = w.Write(filled.Bytes())
	if err != nil {
		log.Println("Error: Failed to write response:", err)
		return err
	}

	return nil
}
