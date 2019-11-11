package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type contextKey string

const (
	jtiContextID  = "jtiContextID"
	userContextID = "userContext"
)

type User struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type store struct {
	sync.RWMutex

	data map[string]User
}

func (s *store) Get(email string) (User, error) {
	s.RLock()
	defer s.RUnlock()

	user, ok := s.data[email]
	if !ok {
		return User{}, errors.New("user not found")
	}

	return user, nil
}

func (s *store) Save(u User) error {
	s.RLock()

	_, ok := s.data[u.Email]
	s.RUnlock()
	if ok {
		return nil
	}

	s.Lock()
	s.data[u.Email] = u
	s.Unlock()
	return nil
}

func GenerateJWT(signingSecret string, u User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":   time.Now().Add(-1 * time.Second),
		"jti":   uuid.New().String(),
		"email": u.Email,
		"iss":   "JWT-revocation-app",
		"exp":   time.Now().Add(time.Hour * 168),
	})

	return token.SignedString([]byte(signingSecret))
}

func getToken(r *http.Request) string {
	return strings.Trim(strings.TrimLeft(r.Header.Get("Authorization"), "Bearer"), " ")
}

func requireAuth(store *store, redis *Client, signingSecret string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			token, err := jwt.Parse(getToken(r), func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(signingSecret), nil
			})

			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				encode(w, apiGenericResponse{
					Message:   "Invalid token provided",
					Status:    false,
					Timestamp: time.Now().Unix(),
				})
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				jti := claims["jti"].(string)

				if err := redis.IsBlacklisted(jti); err != nil {

					w.WriteHeader(http.StatusUnauthorized)
					encode(w, apiGenericResponse{
						Message:   "Authorization denied",
						Status:    false,
						Timestamp: time.Now().Unix(),
					})
					return
				}

				user, err := store.Get(claims["email"].(string))
				if err != nil {
					encode(w, apiGenericResponse{
						Message:   "Could not complete request",
						Status:    false,
						Timestamp: time.Now().Unix(),
					})
					return
				}

				ctx := context.WithValue(r.Context(), userContextID, user)
				ctx = context.WithValue(ctx, jtiContextID, jti)

				next(w, r.WithContext(ctx))
				return
			}

			w.WriteHeader(http.StatusUnauthorized)
			encode(w, apiGenericResponse{
				Message:   "Token is invalid",
				Status:    false,
				Timestamp: time.Now().Unix(),
			})
		}
	}
}
