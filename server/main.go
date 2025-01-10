package main

import (
	"bytes"
	"database/sql"
	// "encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/tavo-wasd-gh/gocors"
)

func main() {
	var err error

	var (
		port   = os.Getenv("PORT")
		db_uri = os.Getenv("DB_URI")
	)

	if port == "" || db_uri == "" {
		log.Fatalf("Fatal: Missing env variables")
	}

	db, err = sql.Open("postgres", db_uri)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Established database connection")

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

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

	if db != nil {
		log.Println("Closing db connection...")
		if err := db.Close(); err != nil {
			log.Fatalf("Fatal: Failed to close db connection: %v", err)
		}
	}

	log.Println("Log: Shutting down...")
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	time.Sleep(1 * time.Second)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// correo := r.FormValue("correo")
		// passwd := r.FormValue("passwd")
		// Validar credenciales, usarlas para determinar id_cuenta
		id_cuenta := "C001"

		if err := jwtSet(w, "jwt_token", id_cuenta, time.Now().Add(15*time.Minute)); err != nil {
			http.Error(w, "Failed to set JWT", http.StatusInternalServerError)
			view(w, "views/login.html", nil)
			return
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

	if r.Method == http.MethodGet {
		if err := jwtValidate(r, "jwt_token"); err != nil {
			view(w, "views/login.html", nil)
			return
		}

		id_cuenta := "C001"

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
}

func view(w http.ResponseWriter, path string, data *Data) error {
	funcMap := template.FuncMap{
		"frac": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return 100 * (a / b)
		},
		"currency": func(amount float64) string {
			formatted := fmt.Sprintf("%.2f", amount)
			parts := strings.Split(formatted, ".")
			intPart := parts[0]
			decPart := parts[1]

			var result strings.Builder
			length := len(intPart)
			for i, digit := range intPart {
				if i > 0 && (length-i)%3 == 0 {
					result.WriteString(".")
				}
				result.WriteRune(digit)
			}

			return result.String() + "," + decPart
		},
		"calcularEmitido": func(tipo string) float64 {
			total, err := calcularEmitido(data, tipo)
			if err != nil {
				return 0
			}
			return total
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
