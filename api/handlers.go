package api

import (
	"errors"
	"net/http"
	"users/usersdb"

	"github.com/labstack/echo/v4"
)

// UsersAPI represents api
type UsersAPI struct {
	*echo.Echo
	db usersdb.DB
}

// NewUsersAPI returns api around db
func NewUsersAPI(db usersdb.DB) *UsersAPI {
	api := &UsersAPI{db: db}
	api.Echo = echo.New()
	api.GET("/users/:id", api.getUser)
	api.PUT("/users/:id", api.updateUser)
	api.DELETE("/users/:id", api.deleteUser)
	api.POST("/users", api.createUser)
	api.GET("/users", api.getUserList)

	oldHandler := api.HTTPErrorHandler

	api.HTTPErrorHandler = func(err error, c echo.Context) {
		if errors.Is(err, usersdb.ErrUserNotExists) {
			c.JSON(http.StatusNotFound, NewMessage(
				err.Error(), nil,
			))
			return
		}
		oldHandler(err, c)
	}

	return api
}

// because of good db api all handlers look almost the same
// bind gets id from url param
func (api *UsersAPI) getUser(c echo.Context) error {
	user := &usersdb.User{}
	if err := c.Bind(user); err != nil {
		return err
	}
	if err := api.db.Get(user); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMessage(
		"OK", user,
	))
}

func (api *UsersAPI) createUser(c echo.Context) error {
	user := &usersdb.User{}
	if err := c.Bind(user); err != nil {
		return err
	}
	if err := api.db.Create(user); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMessage(
		"OK", user,
	))
}

func (api *UsersAPI) updateUser(c echo.Context) error {
	user := &usersdb.User{}
	if err := c.Bind(user); err != nil {
		return err
	}
	if err := (&echo.DefaultBinder{}).BindPathParams(c, user); err != nil {
		return err
	}
	if err := api.db.Update(user); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMessage(
		"OK", user,
	))
}

func (api *UsersAPI) deleteUser(c echo.Context) error {
	user := &usersdb.User{}
	if err := c.Bind(user); err != nil {
		return err
	}
	if err := api.db.Delete(user); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMessage(
		"OK", user,
	))
}

func (api *UsersAPI) getUserList(c echo.Context) error {
	users := []usersdb.User{}
	if err := api.db.GetList(&users); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMessage(
		"OK", users,
	))
}
