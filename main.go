package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"users/api"
	"users/usersdb"

	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/context"
)

func main() {
	host := flag.String("host", ":8000", "hosting address of app")
	dbpath := flag.String("db", "data.json", "path to json database file")
	flushTimeout := flag.Int("flush", 60, "-1 = off autoflush\n"+
		"0 = enable autoflush after every operation\n"+
		"any other = flush timeout in seconds;")
	flag.Parse()

	db, err := usersdb.NewDBJSON(*dbpath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Flush() // save data on exit

	app := api.NewUsersAPI(db)
	if *flushTimeout == 0 {
		db.SetAutoflush(true)
		app.Logger.Print("autoflush enabled")
	} else if *flushTimeout > 0 {
		go func() {
			app.Logger.Printf("flush timeout: %vs", *flushTimeout)
			ch := time.After(time.Duration(*flushTimeout) * time.Second)
			for {
				<-ch
				app.Logger.Printf("flushing...")
				db.Flush()
				ch = time.After(time.Duration(*flushTimeout) * time.Second)
			}
		}()
	} else {
		app.Logger.Print("no autoflush")
	}

	quitChan := make(chan os.Signal)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) // save data on exit by signal (ctrl+c and etc)
	go func() {
		<-quitChan
		app.Logger.Printf("flushing...")
		db.Flush()
		app.Shutdown(context.Background())
	}()

	flushChan := make(chan os.Signal)
	signal.Notify(flushChan, syscall.SIGHUP) // flush data on sighup (kill -HUP <pid>)
	go func() {
		for range flushChan {
			app.Logger.Printf("flushing...")
			db.Flush()
		}
	}()

	app.Use(middleware.Recover())
	app.Use(middleware.Logger())

	app.Logger.Print(app.Start(*host))
}
