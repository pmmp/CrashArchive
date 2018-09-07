package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/handler"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/webhook"
)

func New(db *database.DB, wh *webhook.Webhook, config *app.Config) *chi.Mux {
	r := chi.NewRouter()

	staticDirs := []string{"/css", "/js", "/fonts"}
	workDir, _ := os.Getwd()
	for _, v := range staticDirs {
		dir := filepath.Join(workDir, "static", v[1:])
		FileServer(r, v, http.Dir(dir))
	}

	r.Route("/", func(r chi.Router) {
		r.Use(RealIP)
		r.Use(middleware.Logger)

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			template.ErrorTemplate(w, "", http.StatusNotFound)
		})

		r.Get("/", handler.HomeGet)
		r.Get("/list", handler.ListGet(db))
		r.Get("/view/{reportID}", handler.ViewIDGet)
		r.Get("/download/{reportID}", handler.DownloadGet)

		r.Route("/search", func(r chi.Router) {
			r.Get("/", handler.SearchGet)
			r.Get("/id", handler.SearchIDGet)
			r.Get("/plugin", handler.SearchPluginGet(db))
			r.Get("/build", handler.SearchBuildGet(db))
			r.Get("/report", handler.SearchReportGet(db))
			r.Get("/message", handler.SearchMessageGet(db))
		})

		r.Route("/submit", func(r chi.Router) {
			r.Get("/", handler.SubmitGet)
			r.Post("/", handler.SubmitPost(db, wh, config))
			r.Post("/api", handler.SubmitPost(db, wh, config))
		})
	})
	return r
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, ":*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
