package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	port = 8999
)

func main() {

	signingSecret := os.Getenv("JWT_SIGNING_SECRET")
	if len(signingSecret) == 0 {
		log.Fatal("JWT_SIGNING_SECRET not found in environment")
	}

	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}

	redis, err := NewRediClient(dsn)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	s := &store{
		RWMutex: sync.RWMutex{},
		data:    make(map[string]User),
	}

	mux.HandleFunc("/login", login(s, signingSecret))

	auth := requireAuth(s, redis, signingSecret)

	mux.HandleFunc("/user/logout", auth(logoutHandler(redis)))

	mux.HandleFunc("/user/profile", auth(userProfile))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatal(err)
	}
}

type apiGenericResponse struct {
	Message   string `json:"message"`
	Status    bool   `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

func encode(w io.Writer, v interface{}) {
	_ = json.NewEncoder(w).Encode(v)
}

func login(s *store, signingSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var u User

		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			encode(w, apiGenericResponse{
				Message:   "Invalid request body ",
				Status:    false,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		if u.FullName == "" {
			w.WriteHeader(http.StatusBadRequest)
			encode(w, apiGenericResponse{
				Message:   "Please provide your name",
				Status:    false,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		if u.Email == "" {
			w.WriteHeader(http.StatusBadRequest)
			encode(w, apiGenericResponse{
				Message:   "Please provide your email",
				Status:    false,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		// no errors
		_ = s.Save(u)

		token, err := GenerateJWT(signingSecret, u)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			encode(w, apiGenericResponse{
				Message:   "Could not generate JWT",
				Status:    false,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		w.Header().Set("X-JWT-APP", token)
		encode(w, apiGenericResponse{
			Message:   "You have been logged in successfully",
			Status:    true,
			Timestamp: time.Now().Unix(),
		})
	}
}

func userProfile(w http.ResponseWriter, r *http.Request) {
	encode(w, r.Context().Value(userContextID))
}
