package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	port = 8999
)

func main() {

	signingSecret := os.Getenv("JWT_SIGNING_SECRET")
	if len(signingSecret) == 0 {
		log.Fatal("JWT_SIGNING_SECRET not found in environment")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/register", nil)
	mux.HandleFunc("/logout", nil)

	mux.HandleFunc("/user/profile", nil)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatal(err)
	}
}

func register(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var u User

		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {

		}

	}
}
