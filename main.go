package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type errorRes struct {
	Error string `json:"error"`
}

type apiConfig struct {
	fileServerHits atomic.Int32
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
	}
}

func main() {

	apiCfg := apiConfig{}

	mux := http.NewServeMux()

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	mux.HandleFunc("GET /api/healthz", func(resWriter http.ResponseWriter, req *http.Request) {
		resWriter.Header().Add("Content-Type", "text/plain; charset=utf-8")
		resWriter.WriteHeader(200)
		resWriter.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricHandler())
	mux.HandleFunc("POST /admin/reset", apiCfg.metricReset())

	// FE facing
	mux.Handle("/app", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	// validate_chirp
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, req *http.Request) {
		type param struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(req.Body)
		prm := param{}
		err := decoder.Decode(&prm)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			writeError(w, "Something went wrong")
			return
		}

		if len(prm.Body) > 140 {
			writeError(w, "Chirp is too long")
			return
		}

		type returnVal struct {
			Valid bool `json:"valid"`
		}

		resBody := returnVal{
			Valid: true,
		}

		dat, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(dat)

	})

	server.ListenAndServe()
}

func writeError(w http.ResponseWriter, msg string) {
	errResp := errorRes{
		Error: msg,
	}
	dat, err := json.Marshal(errResp)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(400)
	w.Write(dat)
}
