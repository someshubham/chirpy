package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/someshubham/chirpy/internal/auth"
	"github.com/someshubham/chirpy/internal/database"
)

type apiConfig struct {
	platform              string
	tokenSecret           string
	accessTokenExpiration time.Duration
	fileServerHits        atomic.Int32
	db                    *database.Queries
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

func (a *apiConfig) handleRefreshToken() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			writeError(w, err.Error())
			return
		}

		refreshTokenData, err := a.db.GetUserIdFromRefreshToken(r.Context(), token)
		if err != nil || refreshTokenData.RevokedAt.Valid {
			w.WriteHeader(401)
			return
		}

		if refreshTokenData.ExpiresAt.Before(time.Now()) {
			w.WriteHeader(401)
			return
		}

		accessToken, err := auth.MakeJWT(refreshTokenData.UserID, a.tokenSecret, a.accessTokenExpiration)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Unable to create a auth token " + err.Error()))
			return
		}

		type returnVal struct {
			Token string `json:"token"`
		}

		val := returnVal{
			Token: accessToken,
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
