package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/gorilla/csrf"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/handler"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
	"github.com/pmmp/CrashArchive/app/webhook"
)

func New(db *database.DB, wh *webhook.Webhook, config *app.Config, csrfKey []byte) *chi.Mux {
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

		//CSRF-protected routes
		r.Route("/", func(r chi.Router) {
			r.Use(csrf.Protect(
				csrfKey,
				csrf.Secure(!config.CsrfInsecureCookies),
			))

			r.NotFound(func(w http.ResponseWriter, r *http.Request) {
				template.ErrorTemplate(w, r, "", http.StatusNotFound)
			})

			r.Get("/", handler.HomeGet)

			if config.GitHubAuth != nil && config.GitHubAuth.Enabled {
				r.Get("/github_callback", handler.LoginGetGithubCallback(config.GitHubAuth))
			} else {
				r.Get("/login", handler.LoginGetUserPassword)
				r.Post("/login", handler.LoginPostUserPassword(db))
			}
			r.Get("/logout", handler.LogoutGet)
			r.Get("/list", handler.ListGet(db))
			r.Get("/view/{reportID}", handler.ViewIDGet(db, config))
			r.Get("/view/{reportID}/raw", handler.ViewIDRawGet(db, config.ViewReportRequiresAuth))
			r.Get("/download/{reportID}", handler.DownloadGet(db))
			r.Post("/delete/{reportID}", handler.DeletePost(db))
			r.Get("/duplicates/{reportID}", handler.DuplicatesGet(db))
		})

		//these APIs don't check CSRF and won't generate tokens
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
