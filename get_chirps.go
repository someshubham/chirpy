package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/someshubham/chirpy/data"
	"github.com/someshubham/chirpy/internal/database"
)

func (a *apiConfig) handleGetAllChirps() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authorId := r.URL.Query().Get("author_id")
		dbChirps, err := getChirps(r.Context(), a.db, authorId)

		if err != nil {
			writeError(w, "Unable to fetch Chirps")
			return
		}

		var chirps []data.Chirp
		for _, dbChirp := range dbChirps {
			chirps = append(chirps, data.NewChirpFromDB(dbChirp))
		}

		dat, err := json.Marshal(chirps)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(dat)
	}
}

func (a *apiConfig) handleGetChirpById() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		chirpID, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			writeError(w, "Unable to parse UUID")
			return
		}

		dbChirp, err := a.db.GetChirpByID(r.Context(), chirpID)
		if err != nil {
			w.WriteHeader(404)
			return
		}

		chirp := data.NewChirpFromDB(dbChirp)

		dat, err := json.Marshal(chirp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(dat)
	}
}

func getChirps(context context.Context, db *database.Queries, authorId string) ([]database.Chirp, error) {
	var dbChirps []database.Chirp
	if authorId != "" {
		usrId, err := uuid.Parse(authorId)
		if err != nil {
			return dbChirps, fmt.Errorf("incorrect input")
		}
		return db.GetAllChirpsByAuthor(context, usrId)
	}
	return db.GetAllChirps(context)

}
