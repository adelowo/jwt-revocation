package main

import (
	"errors"
	"sync"
)

type User struct {
	Email string `json:"email"`
	FullName string `json:"full_name"`
}

type store struct {
	sync.RWMutex

	data map[string]User
}

func (s *store) Get(email string) (User, error) {
	s.RLock()
	defer s.RUnlock()

	user,ok := s.data[email]
	if !ok {
		return User{}, errors.New("user not found")
	}

	return user,nil
}

func (s *store) Save(u User) error {
	s.Lock()
	defer s.Unlock()

	s.data[u.Email] = u
	return nil
}