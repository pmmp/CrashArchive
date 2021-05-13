package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/handler"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
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
		r.Use(user.CheckLoginCookieMiddleware)

		r.Get("/login", handler.LoginGet)
		r.Post("/login", handler.LoginPost(db))
		r.Get("/logout", handler.LogoutGet)

		r.Group(func(r chi.Router) {
			if !config.Public {
				r.Use(MustBeLogged)
			}

			r.NotFound(func(w http.ResponseWriter, r *http.Request) {
				template.ErrorTemplate(w, r, "", http.StatusNotFound)
			})

			r.Get("/", handler.HomeGet)
			r.Get("/list", handler.ListGet(db))
			r.Get("/view/{reportID}", handler.ViewIDGet(db))
			r.Get("/view/{reportID}/raw", handler.ViewIDRawGet(db))
			r.Get("/download/{reportID}", handler.DownloadGet(db))
			r.Get("/delete/{reportID}", handler.DeleteGet(db))
			r.Route("/search", func(r chi.Router) {
				r.Get("/", handler.SearchGet)
				r.Get("/id", handler.SearchIDGet)
				r.Get("/report", handler.SearchReportGet(db))
			})
		})

		r.Route("/submit", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				if !config.Public {
					r.Use(MustBeLogged)
				}

				r.Get("/", handler.SubmitGet)
			})

			r.Group(func(r chi.Router) {
				if !config.Public {
					r.Use(SubmitAllowed(config))
				}

				r.Post("/", handler.SubmitPost(db, wh, config))
				r.Post("/api", handler.SubmitPost(db, wh, config))
			})
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
