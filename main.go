package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/julillermo/Chirpy.git/internal/auth"
	"github.com/julillermo/Chirpy.git/internal/database"
	_ "github.com/lib/pq"
)

/* Regarding Flow to reach database */
// 1. Application receives query
// 2. database/sql (common built-in API and connection management)
// 3. database driver (database specific protocol imlementation like lib/pq)
// 4. database server

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
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, req *http.Request) {
			cfg.fileserverHits.Add(1)

			writer.Header().Set("Cache-Control", "no-cache")
			writer.WriteHeader(http.StatusOK)

			next.ServeHTTP(writer, req)
		},
	)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, psqlConnectionErr := sql.Open("postgres", dbURL)
	if psqlConnectionErr != nil {
		log.Printf("Error marshalling JSON: %s", psqlConnectionErr)
		os.Exit(1)
	}
	defer db.Close()

	dbQueries := database.New(db)

	serveMux := http.NewServeMux()
	serverStruct := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	apiCfg := &apiConfig{
		dbQueries: dbQueries,
	}

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
		writer.WriteHeader(http.StatusOK)

		body := "OK"
		writer.Write([]byte(body))
	})

	serveMux.HandleFunc("POST /api/validate_chirp", func(writer http.ResponseWriter, request *http.Request) {
		//? JULIUS: note that what's shown here is kind of the long way
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
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write(somethingErrorData)
			return
		}

		if len(reqJsonDecoded.Body) > 140 {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)
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
		writer.WriteHeader(http.StatusOK)
		writer.Write(returnCleanedBodyData)
	})

	serveMux.HandleFunc("GET /api/chirps/{id}", func(writer http.ResponseWriter, request *http.Request) {
		type resErr struct {
			Error string `json:"error"`
		}
		type resJSON struct {
			Id        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Body      string `json:"body"`
			UserId    string `json:"user_id"`
		}

		idString := request.PathValue("id")
		uuid, err := uuid.Parse(idString)
		if err != nil {
			http.Error(writer, "invalid chirp ID", http.StatusBadRequest)
			return
		}

		somethingError := resErr{
			Error: "Something went wrong",
		}
		somethingErrorData, sometingErrDetails := json.Marshal(somethingError)
		if sometingErrDetails != nil {
			log.Printf("Error marshalling JSON: %s", sometingErrDetails)
		}

		chirpRes, chirpResErr := apiCfg.dbQueries.GetChirp(request.Context(), uuid)
		if errors.Is(chirpResErr, sql.ErrNoRows) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		data, err := json.Marshal(resJSON{
			Id:        chirpRes.ID.String(),
			CreatedAt: chirpRes.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: chirpRes.UpdatedAt.Time.Format(time.RFC3339),
			Body:      chirpRes.Body,
			UserId:    chirpRes.UserID.UUID.String(),
		})
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write(somethingErrorData)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		writer.Write(data)
	})

	serveMux.HandleFunc("GET /api/chirps", func(writer http.ResponseWriter, request *http.Request) {
		type resErr struct {
			Error string `json:"error"`
		}
		type resJSON struct {
			Id        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Body      string `json:"body"`
			UserId    string `json:"user_id"`
		}

		chirpsRes, chirpsResErr := apiCfg.dbQueries.GetAllChirps(request.Context())
		if chirpsResErr != nil {
			log.Printf("Error with database query: %s", chirpsResErr)
			http.Error(writer, "could not create chirp", http.StatusInternalServerError)
			return
		}

		response := make([]resJSON, 0, len(chirpsRes))
		for _, chirp := range chirpsRes {
			response = append(response, resJSON{
				Id:        chirp.ID.String(),
				CreatedAt: chirp.CreatedAt.Time.Format(time.RFC3339),
				UpdatedAt: chirp.UpdatedAt.Time.Format(time.RFC3339),
				Body:      chirp.Body,
				UserId:    chirp.UserID.UUID.String(),
			})
		}

		data, err := json.Marshal(response)
		if err != nil {
			http.Error(writer, "could not encode chirps", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		writer.Write(data)
	})

	serveMux.HandleFunc("POST /api/chirps", func(writer http.ResponseWriter, request *http.Request) {
		type reqJSON struct {
			Body   string `json:"body"`
			UserId string `json:"user_id"`
		}
		type resErr struct {
			Error string `json:"error"`
		}
		type resJSON struct {
			Id        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Body      string `json:"body"`
			UserId    string `json:"user_id"`
		}

		somethingError := resErr{
			Error: "Something went wrong",
		}
		somethingErrorData, sometingErrDetails := json.Marshal(somethingError)
		if sometingErrDetails != nil {
			log.Printf("Error marshalling JSON: %s", sometingErrDetails)
		}

		decoder := json.NewDecoder(request.Body)
		defer request.Body.Close()

		reqJsonDecoded := reqJSON{}
		if err := decoder.Decode(&reqJsonDecoded); err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write(somethingErrorData)
			return
		}

		userID, err := uuid.Parse(reqJsonDecoded.UserId)
		if err != nil {
			http.Error(writer, "invalid user ID", http.StatusBadRequest)
			return
		}

		chirpRes, chirpResErr := apiCfg.dbQueries.CreateChirp(
			request.Context(),
			database.CreateChirpParams{
				Body: reqJsonDecoded.Body,
				UserID: uuid.NullUUID{
					UUID:  userID,
					Valid: true,
				},
			})
		if chirpResErr != nil {
			log.Printf("Error with database query: %s", chirpResErr)
			http.Error(writer, "could not create chirp", http.StatusInternalServerError)
			return
		}

		chirResValueMarshal, chirResValueMarshalErr := json.Marshal(resJSON{
			Id:        chirpRes.ID.String(),
			CreatedAt: chirpRes.CreatedAt.Time.String(),
			UpdatedAt: chirpRes.UpdatedAt.Time.String(),
			Body:      chirpRes.Body,
			UserId:    chirpRes.UserID.UUID.String(),
		})
		if chirResValueMarshalErr != nil {
			log.Printf("Error marshalling JSON: %s", chirResValueMarshalErr)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		writer.Write(chirResValueMarshal)
	})

	serveMux.HandleFunc("POST /api/users", func(writer http.ResponseWriter, request *http.Request) {
		type reqJSON struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		type resErr struct {
			Error string `json:"error"`
		}
		type resJSON struct {
			Id        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Email     string `json:"email"`
		}

		somethingError := resErr{
			Error: "Something went wrong",
		}
		somethingErrorData, sometingErrDetails := json.Marshal(somethingError)
		if sometingErrDetails != nil {
			log.Printf("Error marshalling JSON: %s", sometingErrDetails)
		}

		decoder := json.NewDecoder(request.Body)
		defer request.Body.Close()

		reqJsonDecoded := reqJSON{}
		err := decoder.Decode(&reqJsonDecoded)
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write(somethingErrorData)
			return
		}

		hashedPW, hashedPWErr := auth.HashPassword(reqJsonDecoded.Password)
		if hashedPWErr != nil {
			log.Printf("Error with password hasing : %s", hashedPWErr)
		}

		userRes, userResErr := apiCfg.dbQueries.CreateUser(
			request.Context(),
			database.CreateUserParams{
				Email:          reqJsonDecoded.Email,
				HashedPassword: hashedPW,
			},
		)
		if userResErr != nil {
			log.Printf("Error with database query: %s", userResErr)
		}

		userResValue := resJSON{
			Id:        userRes.ID.String(),
			CreatedAt: userRes.CreatedAt.Time.String(),
			UpdatedAt: userRes.UpdatedAt.Time.String(),
			Email:     userRes.Email,
		}

		userResValueMarshal, userResValueMarshalErr := json.Marshal(userResValue)
		if userResValueMarshalErr != nil {
			log.Printf("Error marshalling JSON: %s", userResValueMarshalErr)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		writer.Write(userResValueMarshal)
	})

	serveMux.HandleFunc("POST /api/login", func(writer http.ResponseWriter, request *http.Request) {
		type reqJSON struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		type resErr struct {
			Error string `json:"error"`
		}
		type resJSON struct {
			Id        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Email     string `json:"email"`
		}

		somethingError := resErr{
			Error: "Something went wrong",
		}
		somethingErrorData, sometingErrDetails := json.Marshal(somethingError)
		if sometingErrDetails != nil {
			log.Printf("Error marshalling JSON: %s", sometingErrDetails)
		}

		decoder := json.NewDecoder(request.Body)
		defer request.Body.Close()

		reqJsonDecoded := reqJSON{}
		err := decoder.Decode(&reqJsonDecoded)
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write(somethingErrorData)
			return
		}

		userRes, userResErr := apiCfg.dbQueries.GetUserByEmail(request.Context(), reqJsonDecoded.Email)
		if userResErr != nil {
			log.Printf("Error querying User: %s", userResErr)
		}

		validPassword, validPasswordErr := auth.CheckPasswordHash(reqJsonDecoded.Password, userRes.HashedPassword)
		if validPasswordErr != nil {
			log.Printf("Error validating password hash: %s", validPasswordErr)
		}

		if !validPassword {
			incorrectEmailPass, incorrectEmailPassErr := json.Marshal(resErr{
				Error: "Incorrect email or password",
			})
			if incorrectEmailPassErr != nil {
				log.Printf("Error marshalling JSON: %s", incorrectEmailPassErr)
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write(incorrectEmailPass)
		} else {
			userResValueMarshal, userResValueMarshalErr := json.Marshal(resJSON{
				Id:        userRes.ID.String(),
				CreatedAt: userRes.CreatedAt.Time.String(),
				UpdatedAt: userRes.UpdatedAt.Time.String(),
				Email:     userRes.Email,
			})
			if userResValueMarshalErr != nil {
				log.Printf("Error marshalling JSON: %s", userResValueMarshalErr)
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			writer.Write(userResValueMarshal)
		}

	})

	serveMux.HandleFunc("GET /admin/metrics", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)

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
		writer.WriteHeader(http.StatusOK)
		apiCfg.fileserverHits.Store(0)
	})

	log.Fatal(serverStruct.ListenAndServe())
}
