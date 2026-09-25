package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/someshubham/chirpy/internal/database"
)

type apiConfig struct {
	platform       string
	tokenSecret    string
	fileServerHits atomic.Int32
	db             *database.Queries
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
