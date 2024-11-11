package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	var (
		port = os.Getenv("PORT")
	)

	if port == "" {
		fatal("Error cargando credenciales", nil)
	}

	http.HandleFunc("/api/docs/", docsHandler)
	http.HandleFunc("/api/dash/", docsHandler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		msg("Servidor iniciado en el puerto " + port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fatal("Error iniciando servidor", err)
		}
	}()

	<-stop

	msg("Servidor detenido")
}

func msg(msg string) {
	log.Println(msg)
}

func fatal(notice string, err error) {
	log.Fatalf("%s: %v", notice, err)
}

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
}

func docsHandler(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Lógica
}
