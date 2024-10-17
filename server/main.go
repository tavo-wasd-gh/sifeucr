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
		fatal(nil, "Error cargando credenciales")
	}

	http.HandleFunc("/api/documentos/", documentosHandler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		msg("Servidor iniciado en el puerto " + port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fatal(err, "Error iniciando servidor")
		}
	}()

	<-stop

	msg("Servidor detenido")
}

func msg(notice string) {
	log.Println(notice)
}

func fatal(err error, notice string) {
	log.Fatalf("%s: %v", notice, err)
}

func documentosHandler(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Lógica
}

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
}
