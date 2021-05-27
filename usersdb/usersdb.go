package usersdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

// ErrUserNotExists will be wrapped as cause when user not exists
var ErrUserNotExists = errors.New("user not exist")

type accessError struct {
	id int
	error
}

// User model with minimum of data
type User struct {
	ID   int
	Name string
	Age  int
}

// DB is users database interface
// we are using inteface, because json db is not very effective
// so later we can swap it without code rewrite
// implementation should be threadsafe, and should return copies and store copies
type DB interface {
	GetUser(id int) (*User, error)
	UpdateUser(id int, patch *User) error // patch is new data for user, id field ignored
	DeleteUser(id int) error
	CreateUser(*User) (id int, err error)
	GetUserList() ([]User, error)
	Count() int
	// Flush data to save changes
	// we not use autoflush after every change because it slow and requires syncronization
	Flush() error
}

type dbJSON struct {
	// use map for fast access by id, it should be syncronized
	// it should be threadsafe, because db will be used in multithreaded environment
	// (every requests works in new goroutine)
	Users map[int]User
	// Counter is used for generate ids
	Counter int
	mu      sync.RWMutex // use rwmutex for decrease locks count(reading will be more effective)
	path    string
}

// NewDBJSON is a constructor of json db, requires path to db file
// if file not exists, it will be created
func NewDBJSON(path string) (DB, error) {
	db := &dbJSON{path: path, Users: map[int]User{}}
	file, err := os.Open(path)

	// it is ok
	if os.IsNotExist(err) {
		return db, nil
	}
	if err != nil {
		return nil, fmt.Errorf("when initializing db with file: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(db); err != nil {
		return nil, fmt.Errorf("when loading users from %v: %w", path, err)
	}

	return db, nil
}

func (db *dbJSON) GetUser(id int) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, ok := db.Users[id]
	if !ok {
		return nil, fmt.Errorf("when access user with id=%v: %w", id, ErrUserNotExists)
	}

	return &user, nil
}

func (db *dbJSON) DeleteUser(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.Users[id]; !ok {
		return fmt.Errorf("when access user with id=%v: %w", id, ErrUserNotExists)
	}

	delete(db.Users, id)
	return nil
}

func (db *dbJSON) UpdateUser(id int, patch *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.Users[id]; !ok {
		return fmt.Errorf("when access user with id=%v: %w", id, ErrUserNotExists)
	}

	user := db.Users[id]
	// update field only if it not omitted
	// default value == omitted
	if patch.Age > 0 {
		user.Age = patch.Age
	}
	if patch.Name != "" {
		user.Name = patch.Name
	}

	db.Users[id] = user

	return nil
}

func (db *dbJSON) CreateUser(user *User) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.Counter++
	u := *user
	u.ID = db.Counter
	db.Users[u.ID] = u

	return u.ID, nil
}

func (db *dbJSON) GetUserList() ([]User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	res := []User{}
	for _, user := range db.Users {
		res = append(res, user)
	}

	return res, nil
}

func (db *dbJSON) Count() int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return len(db.Users)
}

func (db *dbJSON) Flush() error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	file, err := os.Create(db.path)
	if err != nil {
		return fmt.Errorf("when flushing: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(db); err != nil {
		return fmt.Errorf("when flushing: %w", err)
	}

	return nil
}
