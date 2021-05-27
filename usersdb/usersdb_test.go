package usersdb

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// some basic tests

const dataFile = "testdata/data.json"

type testData struct {
	Counter int
	Users   map[int]User
}

// prepared data
var data = testData{
	Counter: 3,
	Users: map[int]User{
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

func TestLoad(t *testing.T) {
	req := require.New(t)

	createTestData()
	db, err := NewDBJSON(dataFile)
	req.Nil(err, "file should open")
	req.Equal(len(data.Users), db.Count(), "should load all data")
}

func TestFlush(t *testing.T) {
	req := require.New(t)

	createTestData()
	db, err := NewDBJSON(dataFile)
	req.Nil(err, "file should open")
	db.DeleteUser(1)

	req.Nil(db.Flush())
	anotherdb, err := NewDBJSON(dataFile)
	req.Equal(len(data.Users)-1, anotherdb.Count(), "file should contain changed data")
}

type DBJSONTestSuite struct {
	suite.Suite
	db DB
}

func TestDBJSONTestSuite(t *testing.T) {
	suite.Run(t, &DBJSONTestSuite{})
}

func (t *DBJSONTestSuite) SetupTest() {
	createTestData()
	db, err := NewDBJSON(dataFile)
	if err != nil {
		panic(err)
	}
	t.db = db
}

func (t *DBJSONTestSuite) TestCreate() {
	user := &User{
		Name: "John",
		Age:  23,
	}
	id, err := t.db.CreateUser(user)
	t.Require().Nil(err, "operation should succed")
	user.ID = id
	actual, err := t.db.GetUser(id)
	t.Require().Nil(err, "operation should succed")
	t.Require().Equal(user, actual)
}

func (t *DBJSONTestSuite) TestGet() {
	user, err := t.db.GetUser(2)
	t.Require().Nil(err, "operation should succed")
	t.Require().Equal(data.Users[2], *user)
}

func (t *DBJSONTestSuite) TestGetList() {
	users, err := t.db.GetUserList()
	t.Require().Nil(err, "operation should succed")

	expected := []User{}
	for _, u := range data.Users {
		expected = append(expected, u)
	}

	t.Require().Equal(expected, users)
}

func (t *DBJSONTestSuite) TestCount() {
	t.Require().Equal(len(data.Users), t.db.Count())
}

func (t *DBJSONTestSuite) TestDelete() {
	err := t.db.DeleteUser(2)
	t.Require().Nil(err, "operation should succed")

	_, err = t.db.GetUser(2)
	t.Require().NotNil(err, "operation should fail")
}

func (t *DBJSONTestSuite) TestUpdatePartial() {
	expected := data.Users[1]
	expected.Age++
	err := t.db.UpdateUser(1, &User{Age: expected.Age})
	t.Require().Nil(err, "operation should succed")

	actual, err := t.db.GetUser(1)
	t.Require().Nil(err, "operation should succed")
	t.Require().Equal(expected, *actual, "partial update should work")
}

func (t *DBJSONTestSuite) TestUpdateFull() {
	expected := data.Users[1]
	expected.Age++
	expected.Name += " UPD"
	err := t.db.UpdateUser(1, &expected)
	t.Require().Nil(err, "operation should succed")

	actual, err := t.db.GetUser(1)
	t.Require().Nil(err, "operation should succed")
	t.Require().Equal(expected, *actual, "full update should work")
}

func (t *DBJSONTestSuite) TestErrors() {
	_, err := t.db.GetUser(55)
	t.Require().True(errors.Is(err, ErrUserNotExists))
	err = t.db.DeleteUser(55)
	t.Require().True(errors.Is(err, ErrUserNotExists))
}
