package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
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

	serveMux.HandleFunc("GET /api/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(200)

		body := "OK"
		writer.Write([]byte(body))
	})

	serveMux.HandleFunc("POST /api/validate_chirp", func(writer http.ResponseWriter, request *http.Request) {
		//? JULIUS: note that what's shown in is kind of the long way
		// Recall that we had json.NewDecoder() vs. Unmarshal.
		// json.NewEncoder() can be used when preparing JSON for sending as a response

		type reqJSON struct {
			Body string `json:"body"`
		}
		type resErr struct {
			Error string `json:"error"`
		}
		type resSuccess struct {
			Valid bool `json:"valid"`
		}
		type resCleanedBody struct {
			CleanedBody string `json:"cleaned_body"`
		}

		somethingError := resErr{
			Error: "Something went wrong",
		}
		somethingErrorData, sometingErrDetails := json.Marshal(somethingError)
		if sometingErrDetails != nil {
			log.Printf("Error marshalling JSON: %s", sometingErrDetails)
		}

		exceedLengthError := resErr{
			Error: "Chirp is too long",
		}
		exceedLengthErrorData, exceedLengthErrDetails := json.Marshal(exceedLengthError)
		if exceedLengthErrDetails != nil {
			log.Printf("Error marshalling JSON: %s", exceedLengthErrDetails)
		}

		// returnSuccess := resSuccess{
		// 	Valid: true,
		// }
		// returnSuccessData, returnSuccessErr := json.Marshal(returnSuccess)
		// if returnSuccessErr != nil {
		// 	log.Printf("Error marshalling JSON: %s", returnSuccessErr)
		// }

		decoder := json.NewDecoder(request.Body)
		defer request.Body.Close()

		reqJsonDecoded := reqJSON{}
		err := decoder.Decode(&reqJsonDecoded)
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(500)
			writer.Write(somethingErrorData)
			return
		}

		if len(reqJsonDecoded.Body) > 140 {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(400)
			writer.Write(exceedLengthErrorData)
			return
		}

		bannedWords := []string{
			"kerfuffle",
			"sharbert",
			"fornax",
		}
		words := strings.Split(reqJsonDecoded.Body, " ")
		cleanedWordsSlice := []string{}
		for _, word := range words {
			fmt.Println("word check: ", word)
			if slices.Contains(bannedWords, strings.ToLower(word)) {
				cleanedWordsSlice = append(cleanedWordsSlice, "****")
			} else {
				cleanedWordsSlice = append(cleanedWordsSlice, word)

			}
			// for _, bannedWord := range bannedWords {
			// 	if strings.ToLower(word) == bannedWord {
			// 		cleanedWordsSlice = append(cleanedWordsSlice, "****")
			// 	}
			// }
		}

		cleanedWords := strings.Join(cleanedWordsSlice, " ")
		// cleanedWordsMarshall, cleanedWordsMarshalErr := json.Marshal(cleanedWords)
		// if cleanedWordsMarshalErr != nil {
		// 	log.Printf("Error marshalling JSON: %s", cleanedWordsMarshalErr)
		// }

		// returnSuccess := resSuccess{
		// 	Valid: true,
		// }
		// returnSuccessData, returnSuccessErr := json.Marshal(returnSuccess)
		// if exceedLengthErrDetails != nil {
		// 	log.Printf("Error marshalling JSON: %s", returnSuccessErr)
		// }

		returnCleanedBody := resCleanedBody{
			CleanedBody: cleanedWords,
		}
		returnCleanedBodyData, returnCleanedBodyErr := json.Marshal(returnCleanedBody)
		if returnCleanedBodyErr != nil {
			log.Printf("Error marshalling JSON: %s", returnCleanedBodyErr)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		writer.Write(returnCleanedBodyData)
	})

	serveMux.HandleFunc("GET /admin/metrics", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(200)

		body := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`,
			apiCfg.fileserverHits.Load())
		writer.Write([]byte(body))
	})

	serveMux.HandleFunc("POST /admin/reset", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		apiCfg.fileserverHits.Store(0)
	})

	log.Fatal(serverStruct.ListenAndServe())
}
