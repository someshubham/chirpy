package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/someshubham/chirpy/internal/auth"
)

func (a *apiConfig) handleUserLogin() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type param struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		prm := param{}
		err := decoder.Decode(&prm)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			writeError(w, "Something went wrong")
			return
		}

		usr, err := a.db.GetUserByEmail(r.Context(), prm.Email)
		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte("Incorrect email or password"))
			return
		}

		ok, err := auth.CheckPasswordHash(prm.Password, usr.HashedPassword)
		if err != nil || !ok {
			w.WriteHeader(401)
			w.Write([]byte("Incorrect email or password"))
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
		w.WriteHeader(200)
		w.Write(dat)

	}
}
