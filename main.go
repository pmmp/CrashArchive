package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/router"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/webhook"
)

const dbRetry = 5

func main() {
	log.SetFlags(log.Lshortfile)

	configPath := flag.String("c", "./config/config.json", "path to `config.json`")
	flag.Parse()

	var err error
	config, err := app.LoadConfig(*configPath)
	if err != nil {
		log.Printf("unable to load config: %v", err)
		os.Exit(1)
	}

	if err := template.Preload(config.Template); err != nil {
		log.Fatal(err)
	}

	var wh *webhook.Webhook = nil
	if config.SlackURLs != nil {
		wh = webhook.New(config.Domain, config.SlackURLs, config.SlackHookInterval)
	}

	var retry int
	var db *database.DB = nil
loop:
	for {
		if retry == dbRetry {
			log.Println("could not connect to database")
			os.Exit(1)
		}

		db, err = database.New(config.Database)
		if err == nil {
			if err := db.Ping(); err != nil {
				log.Println(err)
				os.Exit(1)
			}
			break loop
		} else {
			log.Println(err)
		}
		log.Printf("unable to connect to database: sleeping...\n")
		time.Sleep(5 * time.Second)
		retry++
	}

	r := router.New(db, wh, config)
	log.Printf("listening on: %s\n", config.ListenAddress)
	if err = http.ListenAndServe(config.ListenAddress, r); err != nil {
		log.Fatal(err)
	}

}
