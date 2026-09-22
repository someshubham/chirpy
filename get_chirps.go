package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/someshubham/chirpy/data"
)

func (a *apiConfig) handleGetAllChirps() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dbChirps, err := a.db.GetAllChirps(r.Context())
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
