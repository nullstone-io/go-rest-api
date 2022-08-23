package rest

import (
	"fmt"
	"sync"
)

type User struct {
	Name string `json:"name"`
}

var _ DataAccess[string, User] = &Users{}

type Users struct {
	entities    map[string]*User
	initializer sync.Once
}

func (u *Users) init() {
	u.entities = map[string]*User{}
}

func (u *Users) ParseKey(val string) (string, error) {
	return val, nil
}

func (u *Users) Read(key string) (*User, error) {
	u.initializer.Do(u.init)
	user, ok := u.entities[key]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (u *Users) Exists(user User) (bool, error) {
	u.initializer.Do(u.init)
	_, ok := u.entities[user.Name]
	return ok, nil
}

func (u *Users) Create(user User) (*User, error) {
	u.initializer.Do(u.init)
	_, ok := u.entities[user.Name]
	if ok {
		return nil, fmt.Errorf("user already exists")
	}
	u.entities[user.Name] = &user
	return &user, nil
}

func (u *Users) Ensure(user User) (*User, error) {
	u.initializer.Do(u.init)
	if existing, ok := u.entities[user.Name]; ok {
		return existing, nil
	}
	u.entities[user.Name] = &user
	return &user, nil
}

func (u *Users) Update(key string, user User) (*User, error) {
	u.initializer.Do(u.init)
	u.entities[key] = &user
	return &user, nil
}

func (u *Users) Drop(key string) (bool, error) {
	u.initializer.Do(u.init)
	_, ok := u.entities[key]
	delete(u.entities, key)
	return ok, nil
}
