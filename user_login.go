package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/someshubham/chirpy/internal/auth"
)

func (a *apiConfig) handleUserLogin() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type param struct {
			Email              string `json:"email"`
			Password           string `json:"password"`
			ExpiringTimeSecond string `json:"expires_in_seconds"`
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

		expireTime := time.Duration(time.Minute * 60)

		if prm.ExpiringTimeSecond != "" {
			newExpireTime, err := time.ParseDuration(prm.ExpiringTimeSecond + "h")
			if err == nil {
				expireTime = newExpireTime
			}
		}

		tokenString, err := auth.MakeJWT(usr.ID, a.tokenSecret, expireTime)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Unable to create a auth token " + err.Error()))
			return
		}

		type returnVal struct {
			ID        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Email     string `json:"email"`
			Token     string `json:"token"`
		}

		val := returnVal{
			ID:        usr.ID.String(),
			CreatedAt: usr.CreatedAt.String(),
			UpdatedAt: usr.UpdatedAt.String(),
			Email:     usr.Email,
			Token:     tokenString,
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
