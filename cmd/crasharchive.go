package main

import (
	"crypto/rand"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/crashreport"
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

	if config.ErrorCleanPatterns != nil {
		crashreport.PrepareErrorCleanPatterns(config.ErrorCleanPatterns)
	}

	githubAppClientId := ""
	if config.GitHubAuth != nil && config.GitHubAuth.Enabled {
		githubAppClientId = config.GitHubAuth.ClientId
		log.Printf("GitHub Auth enabled")
	} else {
		log.Printf("GitHub Auth disabled. Use bin/crasharchive-adduser.go to add new admin users to login with username & password.")
	}

	if err := template.Preload(config.Template, githubAppClientId); err != nil {
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
		time.Sleep(time.Second)
		retry++
	}
	db.UpdateTables()

	csrfKey, err := ioutil.ReadFile("./config/csrf-key.bin")
	if err != nil || len(csrfKey) != 32 {
		log.Println("Unable to read CSRF key, generating a new one!")
		csrfKey = make([]byte, 32)
		rand.Read(csrfKey)
		ioutil.WriteFile("./config/csrf-key.bin", csrfKey, 0644)
		log.Println("Successfully generated new CSRF key")
	} else {
		log.Println("Reusing existing CSRF key")
	}
	if config.CsrfInsecureCookies {
		log.Println("WARNING: Secure CSRF cookies are disabled. Set CsrfInsecureCookies to true in config.json to enable.")
	} else {
		log.Println("Secure CSRF cookies are enabled. If you're not using HTTPS and get CSRF errors, set CsrfInsecureCookies to false in config.json to disable.")
	}

	r := router.New(db, wh, config, csrfKey)
	log.Printf("listening on: %s\n", config.ListenAddress)
	if err = http.ListenAndServe(config.ListenAddress, r); err != nil {
		log.Fatal(err)
	}

}
