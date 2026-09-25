package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/someshubham/chirpy/data"
	"github.com/someshubham/chirpy/internal/auth"
	"github.com/someshubham/chirpy/internal/database"
)

func (a *apiConfig) handlePostChirp() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type param struct {
			Body   string `json:"body"`
			UserId string `json:"user_id"`
		}

		decoder := json.NewDecoder(r.Body)
		prm := param{}
		err := decoder.Decode(&prm)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			writeError(w, "Something went wrong")
			return
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			writeError(w, err.Error())
			return
		}

		userUuid, err := auth.ValidateJWT(token, a.tokenSecret)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		if len(prm.Body) > 140 {
			writeError(w, "Chirp is too long")
			return
		}

		cleanedBody := cleanBody(prm.Body)

		dbChirp, err := a.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   cleanedBody,
			UserID: userUuid,
		})

		if err != nil {
			writeError(w, "Unable to create Chirp")
			return
		}

		chirp := data.NewChirpFromDB(dbChirp)

		dat, err := json.Marshal(chirp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(201)
		w.Write(dat)
	}
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
