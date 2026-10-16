package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/someshubham/chirpy/internal/auth"
	"github.com/someshubham/chirpy/internal/database"
)

func (a *apiConfig) handleDeleteChirpById() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		chirpID, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			writeError(w, "Unable to parse UUID")
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

		chirp, err := a.db.GetChirpByID(r.Context(), chirpID)
		if err != nil {
			w.WriteHeader(404)
			return
		}

		if chirp.UserID != userUuid {
			w.WriteHeader(403)
			return
		}

		err = a.db.DeleteChirpById(r.Context(), database.DeleteChirpByIdParams{
			ID:     chirp.ID,
			UserID: chirp.UserID,
		})

		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(204)
	}
}
