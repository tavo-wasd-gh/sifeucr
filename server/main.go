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
	// "strings"
	"syscall"
	"time"

	"github.com/tavo-wasd-gh/gocors"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func main() {
	var (
		port = os.Getenv("PORT")
		db_uri = os.Getenv("DB_URI")
	)

	if port == "" || db_uri == "" {
		log.Fatalf("Fatal: Missing env variables")
	}

	db, err := sql.Open("postgres", db_uri)
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
		if err := r.ParseForm() ; err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		correo := r.FormValue("correo")
		passwd := r.FormValue("passwd")

		// check correo and passwd...

		// set new jwt

		w.Write([]byte("Hello world! "+correo+":"+passwd))
		return
	}

	if r.Method == http.MethodGet {
		cookie, err := r.Cookie("jwt_token")
		if err != nil { // No JWT cookie found or other error
			loginTmpl, err := os.ReadFile("views/login.html")
			if err != nil {
				log.Println("Error: Failed to read template file:", err)
				http.Error(w, "Failed to read template file", http.StatusInternalServerError)
				return
			}

			login, err := fill(string(loginTmpl), "")
			if err != nil {
				log.Println("Error: Failed to render template:", err)
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(login)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid { // JWT is not valid
			// also show login view
		}

		// JWT is valid
		// show dashboard view
		w.WriteHeader(http.StatusOK)
		w.Write(dashboard)
		return
	}

	/*
	id := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/api/dashboard/"), "/", 2)[0]

	dashboardData := Dashboard{
		Cuenta: cuenta,
		Periodo: time.Now().Year(),
		Servicios: servicios,
		Suministros: suministros,
		Bienes: bienes,
	}

	dashboardTmpl, err := os.ReadFile("views/dashboard.html")
	if err != nil {
		log.Println("Error: Failed to read template file:", err)
		http.Error(w, "Failed to read template file", http.StatusInternalServerError)
		return
	}

	dashboard, err := fill(string(dashboardTmpl), data)
	if err != nil {
		log.Println("Error: Failed to render template:", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	*/
}

func fill(htmlTemplate string, data interface{}) ([]byte, error) {
	template, err := template.New("").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var filled bytes.Buffer
	if err := template.Execute(&filled, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return filled.Bytes(), nil
}
