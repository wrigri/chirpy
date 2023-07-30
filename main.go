package main

import (
	"encoding/json"
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

func healthHandler(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func writeErrorResponse(w http.ResponseWriter, code int, errStr string) {
	type errResponse struct {
		Error string `json:"error"`
	}
	errResp := errResponse{Error: errStr}
	resp, err := json.Marshal(errResp)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write([]byte(resp))
}

func writeValidResponse(w http.ResponseWriter, code int, valid bool) {
	type validResponse struct {
		Valid bool `json:"valid"`
	}
	valResp := validResponse{Valid: valid}
	resp, err := json.Marshal(valResp)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write([]byte(resp))
}

func validateHandler(writer http.ResponseWriter, req *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(req.Body)
	chp := chirp{}
	err := decoder.Decode(&chp)
	if err != nil {
		writeErrorResponse(writer, 400, err.Error())
		return
	}
	if len(chp.Body) == 0 {
		writeErrorResponse(writer, 400, "Empty Chirp")
		return
	}

	if len(chp.Body) > 140 {
		writeErrorResponse(writer, 400, "Chirp is too long")
		return
	}
	writeValidResponse(writer, 200, true)
}

func main() {
	r := chi.NewRouter()
	api := chi.NewRouter()
	admin := chi.NewRouter()
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
	api.Post("/validate_chirp", validateHandler)
	admin.Get("/metrics", apiCfg.metricsHandler)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

}
