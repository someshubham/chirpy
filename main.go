package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/someshubham/chirpy/internal/database"
)

type errorRes struct {
	Error string `json:"error"`
}

type apiConfig struct {
	platform       string
	fileServerHits atomic.Int32
	db             *database.Queries
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileServerHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) metricHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-Type", "text/html")
		w.Write(fmt.Appendf(nil,
			`
	<html>
  		<body>
    		<h1>Welcome, Chirpy Admin</h1>
    		<p>Chirpy has been visited %d times!</p>
  		</body>
	</html>
`,
			a.fileServerHits.Load()))
	}
}

func (a *apiConfig) metricReset() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		a.fileServerHits.Store(0)
		if strings.Compare(a.platform, "dev") != 0 {
			w.WriteHeader(403)
			return
		}
		err := a.db.DeleteAllUsers(r.Context())
		if err != nil {
			fmt.Println("Unable to delete records")
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	}
}

func (a *apiConfig) addUsers() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type param struct {
			Email string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		prm := param{}
		err := decoder.Decode(&prm)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			writeError(w, "Something went wrong")
			return
		}

		usr, err := a.db.CreateUser(r.Context(), prm.Email)
		if err != nil {
			writeError(w, "Unable to create a user")
			return
		}

		type returnVal struct {
			ID        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Email     string `json:"email"`
		}

		val := returnVal{
			ID:        usr.ID.String(),
			CreatedAt: usr.CreatedAt.String(),
			UpdatedAt: usr.UpdatedAt.String(),
			Email:     usr.Email,
		}

		dat, err := json.Marshal(val)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(201)
		w.Write(dat)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("unable to load .env file")
		return
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Unable to open db\n%s\n", err.Error())
		return
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		db:       dbQueries,
		platform: platform,
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

	// FE facing
	mux.Handle("/app", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	// validate_chirp
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, req *http.Request) {
		type param struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(req.Body)
		prm := param{}
		err := decoder.Decode(&prm)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			writeError(w, "Something went wrong")
			return
		}

		if len(prm.Body) > 140 {
			writeError(w, "Chirp is too long")
			return
		}

		cleanedBody := cleanBody(prm.Body)

		type returnVal struct {
			CleanedBody string `json:"cleaned_body"`
		}

		resBody := returnVal{
			CleanedBody: cleanedBody,
		}

		dat, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(dat)
	})

	server.ListenAndServe()
}

func writeError(w http.ResponseWriter, msg string) {
	errResp := errorRes{
		Error: msg,
	}
	dat, err := json.Marshal(errResp)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(400)
	w.Write(dat)
}

func cleanBody(body string) string {
	profaneWords := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	words := strings.Split(body, " ")
	cleanBody := make([]string, 0)
	for _, word := range words {
		isProfane := false
		for _, profane := range profaneWords {
			if strings.Compare(strings.ToLower(word), profane) == 0 {
				isProfane = true
				break
			}
		}

		if isProfane {
			cleanBody = append(cleanBody, "****")
		} else {
			cleanBody = append(cleanBody, word)
		}
	}
	return strings.Join(cleanBody, " ")
}
