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
	"strconv"
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

const (
	DEFAULT_DATA_DIR     = "/var/lib/sifeucr/data"  // SF_DATA_DIR
	DEFAULT_DB_FILE      = "/var/lib/sifeucr/db.db" // SF_DB_FILE
	DEFAULT_MAX_OBJ_SIZE = 4 << 20                  // SF_MAX_OBJ_SIZE (4MB default)
)

func main() {
	godotenv.Load()

	// Views
	err := views.Init(viewFS, config.ViewMap, config.ViewFormatters)
	if err != nil {
		log.Fatalf("failed to initialize templates: %v", err)
	}

	// DB
	dbFile := os.Getenv("SF_DB_FILE")
	if dbFile == "" {
		dbFile = DEFAULT_DB_FILE
	}
	db, isFirstTimeSetup, err := config.InitDB(dbFile)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Storage
	dataDir := os.Getenv("SF_DATA_DIR")
	if dataDir == "" {
		dataDir = DEFAULT_DATA_DIR
	}
	var maxObjectSize int64 = 0
	maxObjectSizeStr := os.Getenv("SF_MAX_OBJ_SIZE")
	if maxObjectSizeStr == "" {
		maxObjectSize = DEFAULT_MAX_OBJ_SIZE
	} else {
		maxObjectSize, err = strconv.ParseInt(maxObjectSizeStr, 10, 64)
		if err != nil {
			log.Fatalf("failed to set max object size: %v", err)
		}
	}
	s3, err := storage.New(dataDir, maxObjectSize)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	// Production
	isProduction := os.Getenv("SF_PROD") == "1"

	// Sessions
	sessionStore := sessions.NewStore[config.Session](config.TokenLength, config.MaxSessions)

	// Router
	handlerConfig := handlers.Config{
		IsFirstTimeSetup: isFirstTimeSetup,
		Production:       isProduction,
		Logger: &handlers.Logger{
			Enabled: os.Getenv("SF_DEBUG") == "1",
		},
		Views:    nil,
		DB:       db,
		S3:       s3,
		Sessions: sessionStore,
	}
	handler := handlers.New(handlerConfig)
	router := routes(handler)

	// Serve static files
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

	// Port
	port := os.Getenv("SF_PORT")
	if port == "" {
		port = "8080"
	}

	// goroutine
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
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
