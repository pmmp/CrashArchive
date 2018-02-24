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
)

func New(app *app.App) *chi.Mux {
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
			w.WriteHeader(http.StatusNotFound)
			template.ErrorTemplate(w, "Page not found")
		})

		r.Get("/", handler.HomeGet)
		r.Get("/list", handler.ListGet(app))
		r.Get("/view/{reportID}", handler.ViewIDGet)
		r.Get("/download/{reportID}", handler.DownloadGet)

		r.Route("/search", func(r chi.Router) {
			r.Get("/", handler.SearchGet)
			r.Get("/id", handler.SearchIDGet)
			r.Get("/plugin", handler.SearchPluginGet(app))
			r.Get("/build", handler.SearchBuildGet(app))
			r.Get("/report", handler.SearchReportGet(app))
		})

		r.Route("/submit", func(r chi.Router) {
			r.Get("/", handler.SubmitGet)
			r.Post("/", handler.SubmitPost(app))
			r.Post("/api", handler.SubmitPost(app))
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
