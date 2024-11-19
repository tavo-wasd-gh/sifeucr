package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
	"strings"

	"github.com/tavo-wasd-gh/gocors"
)

func main() {
	http.HandleFunc("/hello", HandleHelloWorld)
	http.HandleFunc("/dash/", HandleDashboard)
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.ListenAndServe(":8080", nil)
}

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	time.Sleep(1 * time.Second)

	id := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/dash/"), "/", 2)[0]

	if id != "0" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	htmlTemplate, err := os.ReadFile("server/views/dashboard.html")
	if err != nil {
		http.Error(w, "Failed to read template file", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		ID    string
	}{
		Title: "Dashboard",
		ID:    id,
	}

	filledTemplate, err := fill(string(htmlTemplate), data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(filledTemplate)
}

func HandleHelloWorld(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	time.Sleep(1 * time.Second)

	w.Header().Set("Content-Type", "text/html")

	fmt.Fprintln(w, `
	<h1>Hello, World!</h1>
	<p>Hello, World!</p>
	`)
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
