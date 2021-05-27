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
	db.Delete(&User{ID: 1})

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
	expected := &User{
		Name: "John",
		Age:  23,
	}
	actual := &User{
		Name: expected.Name,
		Age:  expected.Age,
	}

	t.Require().Nil(t.db.Create(actual), "operation should succed")
	expected.ID = actual.ID
	t.Require().Equal(expected, actual)
}

func (t *DBJSONTestSuite) TestGet() {
	user := &User{ID: 2}
	t.Require().Nil(t.db.Get(user), "operation should succed")
	t.Require().Equal(data.Users[2], *user)
}

func (t *DBJSONTestSuite) TestGetList() {
	actual := []User{}
	err := t.db.GetList(&actual)
	t.Require().Nil(err, "operation should succed")

	expected := []User{}
	for _, u := range data.Users {
		expected = append(expected, u)
	}

	t.Require().ElementsMatch(expected, actual)
}

func (t *DBJSONTestSuite) TestCount() {
	t.Require().Equal(len(data.Users), t.db.Count())
}

func (t *DBJSONTestSuite) TestDelete() {
	//delete
	t.Require().Nil(t.db.Delete(&User{ID: 2}), "operation should succed")
	//check it not exist
	t.Require().NotNil(t.db.Get(&User{ID: 2}), "operation should fail")
}

func (t *DBJSONTestSuite) TestUpdatePartial() {
	expected := data.Users[1]
	expected.Age++
	actual := &User{ID: 1, Age: expected.Age}
	err := t.db.Update(actual)
	t.Require().Nil(err, "operation should succed")

	t.Require().Equal(expected, *actual, "partial update should work")
}

func (t *DBJSONTestSuite) TestUpdateFull() {
	expected := data.Users[1]
	expected.Age++
	expected.Name += " UPD"
	actual := &User{ID: 1, Age: expected.Age, Name: expected.Name}
	err := t.db.Update(actual)
	t.Require().Nil(err, "operation should succed")

	t.Require().Equal(expected, *actual, "partial update should work")
}

func (t *DBJSONTestSuite) TestErrors() {
	err := t.db.Get(&User{ID: 55})
	t.Require().True(errors.Is(err, ErrUserNotExists))
	err = t.db.Delete(&User{ID: 55})
	t.Require().True(errors.Is(err, ErrUserNotExists))
}
