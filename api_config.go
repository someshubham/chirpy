package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/someshubham/chirpy/internal/database"
)

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
