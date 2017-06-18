package router

import (
	"net/http"
	"os"
	"path/filepath"

	"bitbucket.org/intyre/ca-pmmp/app"
	"bitbucket.org/intyre/ca-pmmp/app/handler"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

func New(app *app.App) *chi.Mux {
	r := chi.NewRouter()

	staticDirs := []string{"/css", "/js", "/fonts"}
	workDir, _ := os.Getwd()
	for _, v := range staticDirs {
		dir := filepath.Join(workDir, "static", v[1:])
		r.FileServer(v, http.Dir(dir))
	}

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(middleware.RealIP)

		r.Get("/", handler.HomeGet(app))
		r.Get("/list", handler.ListGet(app))
		r.Get("/view/:reportID", handler.ViewIDGet(app))

		r.Route("/search", func(r chi.Router) {
			r.Get("/", handler.SearchGet(app))
			r.Post("/id", handler.SearchIDPost(app))
			r.Post("/plugin", handler.SearchPluginPost(app))
			r.Post("/build", handler.SearchBuildPost(app))
			r.Post("/report", handler.SearchReportPost(app))
		})

		r.Route("/submit", func(r chi.Router) {
			r.Get("/", handler.SubmitGet(app))
			r.Post("/", handler.SubmitPost(app))
			r.Post("/api", handler.SubmitPost(app))
		})
	})
	return r
}
