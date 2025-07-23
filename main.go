package main

import (
	"compress/gzip"
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"git.tavo.one/tavo/axiom/sessions"
	"git.tavo.one/tavo/axiom/storage"
	"git.tavo.one/tavo/axiom/views"
	"github.com/joho/godotenv"

	"sifeucr/config"
	"sifeucr/handlers"
)

//go:embed static/*
var publicFS embed.FS

//go:embed views/*
var viewFS embed.FS

func init() {
	godotenv.Load()

	needed := []string{
		"PRODUCTION",
		"PORT",
		"APP_DATA_DIR",
		"DB_CONNDVR",
		"DB_CONNSTR",
	}

	for _, v := range needed {
		if os.Getenv(v) == "" {
			log.Fatalf("missing environment varialbe: %s", v)
		}
	}
}

func main() {
	err := views.Init(viewFS, config.ViewMap, config.ViewFormatters)
	if err != nil {
		log.Fatalf("failed to initialize templates: %v", err)
	}

	db, isFirstTimeSetup, err := config.InitDB(os.Getenv("DB_CONNDVR"), os.Getenv("DB_CONNSTR"))
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	s3, err := storage.New(os.Getenv("APP_DATA_DIR"), 4<<20)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	isProduction := os.Getenv("PRODUCTION") == "1"
	sessionStore := sessions.NewStore[config.Session](config.TokenLength, config.MaxSessions)

	handlerConfig := handlers.Config{
		IsFirstTimeSetup: isFirstTimeSetup,
		Production:       isProduction,
		Logger: &handlers.Logger{
			Enabled: os.Getenv("DEBUG") == "1",
		},
		Views:    nil,
		DB:       db,
		S3:       s3,
		Sessions: sessionStore,
	}
	handler := handlers.New(handlerConfig)

	router := routes(handler)

	staticFiles, err := fs.Sub(publicFS, "static")
	if err != nil {
		log.Fatalf("failed to create static files filesystem: %v", err)
	}

	router.Handle(
		"GET /s/",
		http.StripPrefix("/s/",
			gzipHandler(http.FileServer(http.FS(staticFiles))),
		),
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	port := os.Getenv("PORT")

	go func() {
		log.Printf("starting on :%s...", port)

		if err := http.ListenAndServe(":"+port, router); err != nil {
			log.Fatalf("fatal: failed to start on port %s: %v", port, err)
		}
	}()

	<-stop

	log.Printf("shutting down...")
}

func gzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		h.ServeHTTP(gzw, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
