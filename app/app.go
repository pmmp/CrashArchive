package app

import (
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/webhook"
)

// App ...
type App struct {
	Config   *Config
	Database *database.DB
	Webhook  *webhook.Webhook
}