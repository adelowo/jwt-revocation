package main

import (
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type User struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type store struct {
	sync.RWMutex

	data map[string]User
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
		"foo":   "bar",
		"nbf":   time.Now().Add(-1 * time.Second),
		"jti":   uuid.New().String(),
		"email": u.Email,
		"iss":   "JWT-revocation-app",
		"exp":   time.Now().Add(time.Hour * 168),
	})

	return token.SignedString([]byte(signingSecret))
}
