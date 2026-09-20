package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

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

	mux.Handle("/app", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	server.ListenAndServe()
}
