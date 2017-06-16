package app

import "bitbucket.org/intyre/ca-pmmp/app/database"

type App struct {
	Config   *Config
	Database *database.DB
}
