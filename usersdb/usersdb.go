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
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// DB is users database interface
// we are using inteface, because json db is not very effective
// so later we can swap it without code rewrite
// implementation should be threadsafe, and should return copies and store copies
// ids should be > 0
// UPD: new api
// arg - criteria AND destination of data
// example:
// u := &User{ID: 1}
// db.Get(u) get User with ID=1
// u contains that user
// it looks cool: db.Delete(&User{ID: 1})
type DB interface {
	Get(*User) error    // arg should contain id
	Update(*User) error // same
	Delete(*User) error // same
	Create(*User) error // arg should containg data
	GetList(*[]User) error
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

func (db *dbJSON) Get(u *User) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, ok := db.Users[u.ID]
	if !ok {
		return fmt.Errorf("when access user with id=%v: %w", u.ID, ErrUserNotExists)
	}

	*u = user
	return nil
}

func (db *dbJSON) Delete(u *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, ok := db.Users[u.ID]
	if !ok {
		return fmt.Errorf("when access user with id=%v: %w", u.ID, ErrUserNotExists)
	}

	delete(db.Users, u.ID)
	*u = user
	return nil
}

func (db *dbJSON) Update(u *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, ok := db.Users[u.ID]
	if !ok {
		return fmt.Errorf("when access user with id=%v: %w", u.ID, ErrUserNotExists)
	}

	// update field only if it not omitted
	// default value == omitted
	if u.Age > 0 {
		user.Age = u.Age
	}
	if u.Name != "" {
		user.Name = u.Name
	}

	db.Users[u.ID] = user
	*u = user
	return nil
}

func (db *dbJSON) Create(u *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.Counter++
	u.ID = db.Counter
	db.Users[u.ID] = *u

	return nil
}

func (db *dbJSON) GetList(ul *[]User) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	*ul = []User{}
	for _, user := range db.Users {
		*ul = append(*ul, user)
	}

	return nil
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
