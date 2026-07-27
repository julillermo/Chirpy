package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiHandler struct{}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

//? JULIUS: Was discussed but not required
// func middlwareLog(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("%s %s", r.Method, r.URL.Path)
// 		next.ServeHTTP(w, r)
// 	})
// }

type apiConfig struct {
	fileserverHits atomic.Int32 // allows modifying and incrementing across goroutines
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, req *http.Request) {
			cfg.fileserverHits.Add(1)

			writer.Header().Set("Cache-Control", "no-cache")
			writer.WriteHeader(200)

			next.ServeHTTP(writer, req)
		},
	)
}

func main() {
	serveMux := http.NewServeMux()
	serverStruct := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	apiCfg := &apiConfig{}

	// serveMux.Handle(
	// 	"/app/",
	// 	http.StripPrefix("/app/",
	// 		http.FileServer(http.Dir("./"))),
	// )

	serveMux.Handle(
		"/app/",
		apiCfg.middlewareMetricsInc(
			http.StripPrefix("/app/",
				http.FileServer(http.Dir("./"))),
		),
	)

	serveMux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(200)

		body := "OK"
		writer.Write([]byte(body))
	})

	serveMux.HandleFunc("GET /metrics", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(200)

		body := fmt.Sprintf("Hits: %v", apiCfg.fileserverHits.Load())
		writer.Write([]byte(body))
	})

	serveMux.HandleFunc("POST /reset", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		apiCfg.fileserverHits.Store(0)
	})

	log.Fatal(serverStruct.ListenAndServe())
}
