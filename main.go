package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"bitbucket.org/intyre/ca-pmmp/app"
	"bitbucket.org/intyre/ca-pmmp/app/database"
	"bitbucket.org/intyre/ca-pmmp/app/router"
)

func main() {
	log.SetFlags(log.Lshortfile)

	configPath := flag.String("c", "./config/config.json", "path to `config.json`")
	flag.Parse()

	var err error
	context := &app.App{}
	context.Config, err = app.LoadConfig(*configPath)
	if err != nil {
		log.Printf("unable to load config: %v", err)
		os.Exit(1)
	}

	time.Sleep(5 * time.Second)
	context.Database, err = database.New(context.Config.Database)
	if err != nil {
		log.Printf("unable to connect to database: %v\n", err)
		os.Exit(1)
	}
loop:
	for i := 0; i <= 5; i++ {
		if i == 5 {
			log.Fatal("unable to ping database")
		}
		if err := context.Database.Ping(); err == nil {
			break loop
		}
		time.Sleep(2 * time.Second)
	}
	r := router.New(context)
	log.Printf("listening on: %s\n", context.Config.ListenAddress)
	if err = http.ListenAndServe(context.Config.ListenAddress, r); err != nil {
		log.Fatal(err)
	}

}
