package main

import (
	"chirpy/internal/database"
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("ADMIN_KEY environment variable is not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("POLKA_KEY environment variable is not set")
	}

	dbQueries := database.New(dbConn)

	apicfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		jwtSecret:      jwtSecret,
		polkaKey:       polkaKey,
		platform:       platform,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apicfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/polka/webhooks", apicfg.handlerWebhook)

	mux.HandleFunc("POST /api/login", apicfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apicfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apicfg.handlerRevoke)

	mux.HandleFunc("POST /api/users", apicfg.handlerUsersCreate)
	mux.HandleFunc("PUT /api/users", apicfg.handlerUsersUpdate)

	mux.HandleFunc("POST /api/chirps", apicfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apicfg.handlerChirpsRetrieve)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apicfg.handlerChirpsGet)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apicfg.handlerChirpsDelete)

	mux.HandleFunc("GET /admin/metrics", apicfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apicfg.handlerReset)

	srv := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving on: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
