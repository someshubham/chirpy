package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/someshubham/chirpy/internal/auth"
)

func (a *apiConfig) webhookPolka() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type param struct {
			Event string `json:"event"`
			Data  struct {
				UserId string `json:"user_id"`
			} `json:"data"`
		}

		key, err := auth.GetAPIKey(r.Header)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		if key != a.polkaKey {
			w.WriteHeader(401)
			return
		}

		decoder := json.NewDecoder(r.Body)
		prm := param{}
		err = decoder.Decode(&prm)
		if err != nil {
			writeError(w, "Something went wrong")
			return
		}

		if prm.Event != "user.upgraded" {
			w.WriteHeader(204)
			return
		}

		usrId, err := uuid.Parse(prm.Data.UserId)
		if err != nil {
			w.WriteHeader(404)
			return
		}

		_, err = a.db.UpgradeUserToChirpyRed(r.Context(), usrId)
		if err != nil {
			w.WriteHeader(404)
			return
		}

		w.WriteHeader(204)
	}
}
