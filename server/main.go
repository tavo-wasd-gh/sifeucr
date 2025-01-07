package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"time"

	"github.com/tavo-wasd-gh/gocors"
)

type account struct {
	id string
	name string
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatalf("Fatal: Missing env variables")
        }

	http.HandleFunc("/api/dashboard", handleDashboard)
	http.Handle("/", http.FileServer(http.Dir("public")))

	stop := make(chan os.Signal, 1)
        signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

        go func() {
		log.Println("Log: Running on :" + port + "...")
                if err := http.ListenAndServe(":" + port, nil); err != nil {
			log.Fatalf("Fatal: Failed to start on port %s: %v", port, err)
                }
        }()

        <-stop

	log.Println("Log: Shutting down...")
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	time.Sleep(200 * time.Millisecond)

	id := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/api/dashboard/"), "/", 2)[0]

	htmlTemplate, err := os.ReadFile("views/dashboard.html")
	if err != nil {
		http.Error(w, "Failed to read template file", http.StatusInternalServerError)
		return
	}

	a := account{
		id: id,
		name: "Mi Asocia",
	}

	filledTemplate, err := fill(string(htmlTemplate), a)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(filledTemplate)
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
