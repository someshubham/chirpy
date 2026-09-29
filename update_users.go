package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/someshubham/chirpy/data"
	"github.com/someshubham/chirpy/internal/auth"
	"github.com/someshubham/chirpy/internal/database"
)

func (a *apiConfig) handleUpdateUsers() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type param struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		prm := param{}
		err := decoder.Decode(&prm)
		if err != nil {
			writeError(w, "Something went wrong")
			return
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		userUuid, err := auth.ValidateJWT(token, a.tokenSecret)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		hash, err := auth.HashPassword(prm.Password)
		if err != nil {
			writeError(w, "Unable to process the password")
			return
		}

		usr, err := a.db.UpdateUserEmailPassword(r.Context(), database.UpdateUserEmailPasswordParams{
			ID:             userUuid,
			Email:          prm.Email,
			HashedPassword: hash,
		})
		if err != nil {
			writeError(w, "Unable to create a user")
			return
		}

		val := data.NewUserFromDB(usr)

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
