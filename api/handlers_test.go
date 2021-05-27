package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"users/usersdb"

	"github.com/stretchr/testify/suite"
)

// copypaste here
// maybe its better to add feature to load data from custom reader?
// but it will enlarge inteface

const dataFile = "testdata/data.json"

type testData struct {
	Counter int
	Users   map[int]usersdb.User
}

// prepared data
var data = testData{
	Counter: 3,
	Users: map[int]usersdb.User{
		1: {ID: 1, Name: "Petya", Age: 20},
		2: {ID: 2, Name: "Alyosha", Age: 30},
		3: {ID: 3, Name: "Vasya", Age: 35},
	},
}

func createTestData() {
	file, err := os.Create(dataFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		panic(err)
	}
}

type UsersAPITestSuite struct {
	suite.Suite
	db  usersdb.DB
	api *UsersAPI
}

func TestUsersAPITestSuite(t *testing.T) {
	suite.Run(t, &UsersAPITestSuite{})
}

func (t *UsersAPITestSuite) SetupTest() {
	createTestData()
	db, err := usersdb.NewDBJSON(dataFile)
	if err != nil {
		panic(err)
	}
	t.db = db
	t.api = NewUsersAPI(db)
}

func (t *UsersAPITestSuite) TestGet() {
	req := httptest.NewRequest("GET", "/users/1", nil)
	rw := httptest.NewRecorder()
	t.api.ServeHTTP(rw, req)
	resp := rw.Result()

	msg := &responseWithUser{}

	t.Require().NoError(json.NewDecoder(resp.Body).Decode(&msg))
	t.Require().Equal(200, resp.StatusCode, msg.Message)

	user := msg.User

	t.Require().Equal(data.Users[1], user)
}

type responseWithUser struct {
	Message string
	User    usersdb.User `json:"data"`
}

func (t *UsersAPITestSuite) TestCreate() {
	oldSize := t.db.Count()
	reqBody := bytes.NewBuffer(nil)
	expected := usersdb.User{
		Name: "Test",
		Age:  32,
	}
	json.NewEncoder(reqBody).Encode(expected)

	req := httptest.NewRequest("POST", "/users", reqBody)
	req.Header.Set("content-type", "application/json")
	rw := httptest.NewRecorder()
	t.api.ServeHTTP(rw, req)
	resp := rw.Result()

	t.Require().Equal(200, resp.StatusCode)

	msg := &responseWithUser{}
	t.Require().NoError(json.NewDecoder(resp.Body).Decode(&msg))

	user := msg.User
	expected.ID = user.ID

	t.Require().Equal(expected, user)
	t.Require().Equal(oldSize+1, t.db.Count())
}

func (t *UsersAPITestSuite) TestUpdate() {
	reqBody := bytes.NewBuffer(nil)
	patch := usersdb.User{Name: "Test"}
	json.NewEncoder(reqBody).Encode(patch)
	expected := data.Users[2]
	expected.Name = patch.Name

	req := httptest.NewRequest("PUT", "/users/2", reqBody)
	req.Header.Set("content-type", "application/json")
	rw := httptest.NewRecorder()
	t.api.ServeHTTP(rw, req)
	resp := rw.Result()

	t.Require().Equal(200, resp.StatusCode)
	msg := &responseWithUser{}
	t.Require().NoError(json.NewDecoder(resp.Body).Decode(&msg))

	user := msg.User

	t.Require().Equal(expected, user)
}

func (t *UsersAPITestSuite) TestDelete() {
	oldSize := t.db.Count()
	expectedUser := data.Users[3]
	req := httptest.NewRequest("DELETE", "/users/3", nil)
	rw := httptest.NewRecorder()
	t.api.ServeHTTP(rw, req)
	resp := rw.Result()

	t.Require().Equal(200, resp.StatusCode)
	msg := &responseWithUser{}
	t.Require().NoError(json.NewDecoder(resp.Body).Decode(&msg))

	user := msg.User

	t.Require().Equal(expectedUser, user)
	t.Require().Equal(oldSize-1, t.db.Count())
}

type responseWithUserList struct {
	Message string
	Users   []usersdb.User `json:"data"`
}

func (t *UsersAPITestSuite) TestGetList() {
	req := httptest.NewRequest("GET", "/users", nil)
	rw := httptest.NewRecorder()
	t.api.ServeHTTP(rw, req)
	resp := rw.Result()

	t.Require().Equal(200, resp.StatusCode)
	msg := &responseWithUserList{}
	t.Require().NoError(json.NewDecoder(resp.Body).Decode(&msg))
	actual := msg.Users

	expected := []usersdb.User{}
	for _, u := range data.Users {
		expected = append(expected, u)
	}

	t.Require().ElementsMatch(expected, actual)
}

func (t *UsersAPITestSuite) TestNotFound() {
	req := httptest.NewRequest("GET", "/users/55", nil)
	rw := httptest.NewRecorder()
	t.api.ServeHTTP(rw, req)
	resp := rw.Result()

	t.Require().Equal(http.StatusNotFound, resp.StatusCode)

	msg := &responseWithUser{}
	t.Require().NoError(json.NewDecoder(resp.Body).Decode(&msg))
}
