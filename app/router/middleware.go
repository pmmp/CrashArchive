package router

import (
	"fmt"
	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/user"
	"log"
	"net"
	"net/http"
	"strings"
)

var cfConnectingIP = http.CanonicalHeaderKey("Cf-Connecting-Ip")
var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

func RealIP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if rip := realIP(r); rip != "" {
			r.RemoteAddr = rip
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func MustBeLogged(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		userInfo := user.GetUserInfo(r)
		if userInfo.Name == "anonymous" {
			_, _ = fmt.Fprintf(w, "Unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func SubmitAllowed(c *app.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				log.Printf("Can't parse remote addr: %v\n", err)
				_, _ = fmt.Fprintf(w, err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if c.SubmitAllowedIpsMap[ip] == "" {
				log.Println("A request came from the stranger. IP=" + r.RemoteAddr)
				_, _ = fmt.Fprintf(w, "Unauthorized")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func realIP(r *http.Request) string {
	var ip string
	if cfcip := r.Header.Get(cfConnectingIP); cfcip != "" {
		ip = cfcip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip

	}
	return ip
}
