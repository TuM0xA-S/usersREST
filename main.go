package main

import (
	"log"
	"users/api"
	"users/usersdb"
)

func main() {
	db, err := usersdb.NewDBJSON("data.json")
	defer db.Flush() // save data on exit
	if err != nil {
		log.Fatal(err)
	}
	app := api.NewUsersAPI(db)
	app.Logger.Fatal(app.Start(":8000"))

}
