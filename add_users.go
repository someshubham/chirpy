package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/someshubham/chirpy/data"
	"github.com/someshubham/chirpy/internal/auth"
	"github.com/someshubham/chirpy/internal/database"
)

func (a *apiConfig) addUsers() func(w http.ResponseWriter, r *http.Request) {
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

		hash, err := auth.HashPassword(prm.Password)
		if err != nil {
			writeError(w, "Unable to process the password")
			return
		}

		usr, err := a.db.CreateUser(r.Context(), database.CreateUserParams{
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
		w.WriteHeader(201)
		w.Write(dat)
	}
}
