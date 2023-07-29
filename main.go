package main

import (
	"fmt"
	"net/http"
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

func healthHandler(writer http.ResponseWriter, req *http.Request) {
	//header := writer.Header()
	//header["content-type"] = []string{"text/plain; charset=utf-8"}
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	corsMux := middlewareCors(mux)
	server := http.Server{
		Addr:    "localhost:8080",
		Handler: corsMux,
	}
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", handler)
	mux.HandleFunc("/healthz", healthHandler)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(handler)

}
