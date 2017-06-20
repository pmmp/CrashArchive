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

const dbRetry = 5

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

	var retry int
loop:
	for {
		if retry == dbRetry {
			log.Println("could not connect to database")
			os.Exit(1)
		}

		context.Database, err = database.New(context.Config.Database)
		if err == nil {
			if err := context.Database.Ping(); err != nil {
				log.Println(err)
				os.Exit(1)
			}
			break loop
		}
		log.Printf("unable to connect to database: sleeping...\n")
		time.Sleep(5 * time.Second)
		retry++
	}

	r := router.New(context)
	log.Printf("listening on: %s\n", context.Config.ListenAddress)
	if err = http.ListenAndServe(context.Config.ListenAddress, r); err != nil {
		log.Fatal(err)
	}

}
