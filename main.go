package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"users/api"
	"users/usersdb"

	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/context"
)

func main() {
	host := flag.String("host", ":8000", "hosting address of app")
	dbpath := flag.String("db", "data.json", "path to json database file")
	flag.Parse()

	db, err := usersdb.NewDBJSON(*dbpath)
	defer db.Flush() // save data on exit
	if err != nil {
		log.Fatal(err)
	}
	app := api.NewUsersAPI(db)

	quitChan := make(chan os.Signal)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) // save data on exit by signal (ctrl+c and etc)
	go func() {
		<-quitChan
		db.Flush()
		app.Shutdown(context.Background())
	}()

	flushChan := make(chan os.Signal)
	signal.Notify(flushChan, syscall.SIGHUP) // flush data on sighup (kill -HUP pid)
	go func() {
		<-flushChan
		db.Flush()
	}()

	app.Use(middleware.Recover())
	app.Use(middleware.Logger())

	app.Logger.Print(app.Start(*host))
}
