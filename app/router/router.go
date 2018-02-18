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

		r.Get("/", handler.HomeGet(app))
		r.Get("/list", handler.ListGet(app))
		r.Get("/list/{pageID}", handler.ListGet(app))
		r.Get("/view/{reportID}", handler.ViewIDGet(app))
		r.Get("/download/{reportID}", handler.DownloadGet(app))

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
