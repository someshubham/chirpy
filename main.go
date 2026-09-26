package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/someshubham/chirpy/internal/database"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("unable to load .env file")
		return
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Unable to open db\n%s\n", err.Error())
		return
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		db:                    dbQueries,
		platform:              platform,
		tokenSecret:           tokenSecret,
		accessTokenExpiration: time.Duration(time.Minute * 60),
	}

	mux := http.NewServeMux()

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	mux.HandleFunc("GET /api/healthz", func(resWriter http.ResponseWriter, req *http.Request) {
		resWriter.Header().Add("Content-Type", "text/plain; charset=utf-8")
		resWriter.WriteHeader(200)
		resWriter.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricHandler())
	mux.HandleFunc("POST /admin/reset", apiCfg.metricReset())
	mux.HandleFunc("POST /api/users", apiCfg.addUsers())
	mux.HandleFunc("POST /api/chirps", apiCfg.handlePostChirp())
	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetAllChirps())
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirpById())
	mux.HandleFunc("POST /api/login", apiCfg.handleUserLogin())

	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshToken())

	// FE facing
	mux.Handle("/app", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	server.ListenAndServe()
}
