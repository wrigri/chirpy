package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type apiConfig struct {
	fileserverHits int
}

func (ac *apiConfig) collectMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func healthHandler(writer http.ResponseWriter, req *http.Request) {
	//header := writer.Header()
	//header["content-type"] = []string{"text/plain; charset=utf-8"}
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (ac *apiConfig) metricsHandler(writer http.ResponseWriter, req *http.Request) {
	template := `<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>`
	writer.WriteHeader(200)
	metrics := fmt.Sprintf(template, ac.fileserverHits)
	writer.Write([]byte(metrics))
}

func main() {
	r := chi.NewRouter()
	api := chi.NewRouter()
	admin := chi.NewRouter()
	//mux := http.NewServeMux()
	apiCfg := apiConfig{}
	corsMux := middlewareCors(r)
	server := http.Server{
		Addr:    "localhost:8080",
		Handler: corsMux,
	}
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	r.Handle("/app", apiCfg.collectMetrics(handler))
	r.Handle("/app/*", apiCfg.collectMetrics(handler))
	r.Mount("/api/", api)
	r.Mount("/admin/", admin)
	api.Get("/healthz", healthHandler)
	admin.Get("/metrics", apiCfg.metricsHandler)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

}
