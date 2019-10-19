package handler

import (
	"log"
	"net/http"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
)

func isAlreadyLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	userInfo := user.GetUserInfo(r)
	if userInfo.Permission != user.View {
		log.Printf("user %s (%s) is already logged in", userInfo.Name, r.RemoteAddr)
		template.ErrorTemplate(w, r, "You're already logged in", http.StatusBadRequest)
		return true
	}
	return false
}
func LoginGet(w http.ResponseWriter, r *http.Request) {
	if(!isAlreadyLoggedIn(w, r)){
		template.ExecuteTemplate(w, r, "login")
	}
}

func LoginPost(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("bad login post from %s: %v", r.RemoteAddr, err)
		template.ErrorTemplate(w, r, "", http.StatusBadRequest)
		return
	}
	if isAlreadyLoggedIn(w, r) {
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	userInfo, err := db.AuthenticateUser(username, []byte(password))
	//TODO: check the type of error (unknown user, wrong password, etc)
	if err != nil {
		log.Printf("%v", err)
		template.ErrorTemplate(w, r, "Failed to login", http.StatusUnauthorized)
		return
	}
	cookie, err2 := user.CreateCookie(userInfo)
	if err2 != nil {
		log.Printf("error logging in %s: %v", r.RemoteAddr, err2)
		template.ErrorTemplate(w, r, "", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
}

func LogoutGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("logging out user on %s", r.RemoteAddr)
	http.SetCookie(w, user.DeleteCookie())
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
