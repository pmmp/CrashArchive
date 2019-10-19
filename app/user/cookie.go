package user

import (
	"context"
	"log"
	"net/http"

    "github.com/gorilla/securecookie"
)

const cookieName string = "userinfo"
var cookieEncoder *securecookie.SecureCookie

func init() {
	cookieEncoder = securecookie.New(securecookie.GenerateRandomKey(32), nil)
}

func CreateCookie(user UserInfo) (*http.Cookie, error) {
	encoded, err := cookieEncoder.Encode(cookieName, user)
	if err == nil {
		return &http.Cookie{
			Name: cookieName,
			Value: encoded,
			//Secure: true,
			HttpOnly: true,
			Path: "/",
		}, nil
	}
	return nil, err
}

func DeleteCookie() *http.Cookie {
	return &http.Cookie{
		Name: cookieName,
		Value: "",
		//Secure: true,
		HttpOnly: true,
		Path: "/",
		MaxAge: -1,
	}
}

func parseCookie(r *http.Request) (UserInfo, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return DefaultUserInfo(), err
	}
	userInfo := UserInfo{}
	if err = cookieEncoder.Decode(cookieName, cookie.Value, &userInfo); err != nil {
		return DefaultUserInfo(), err
	}
	return userInfo, nil
}

type userContextKeyType string
const userContextKey userContextKeyType = "handler_user_info"

func CheckLoginCookieMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo, err := parseCookie(r)
		if err != nil {
			log.Printf("user %s (%s) is not logged in: %v", userInfo.Name, r.RemoteAddr, err)
			http.SetCookie(w, DeleteCookie())
		} else {
			log.Printf("user %s (%s) is logged in by cookie", userInfo.Name, r.RemoteAddr)
		}
		ctx := context.WithValue(r.Context(), userContextKey, userInfo)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserInfo(r *http.Request) UserInfo {
	return r.Context().Value(userContextKey).(UserInfo)
}