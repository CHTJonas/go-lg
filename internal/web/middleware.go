package web

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/handlers"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (serv *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{w, http.StatusOK}
		next.ServeHTTP(lrw, r)
		addr, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			addr = r.RemoteAddr
		}
		httpInfo := fmt.Sprintf("\"%s %s %s\"", r.Method, r.URL.Path, r.Proto)
		refererInfo := fmt.Sprintf("\"%s\"", r.Referer())
		if refererInfo == "\"\"" {
			refererInfo = "\"-\""
		}
		uaInfo := fmt.Sprintf("\"%s\"", r.UserAgent())
		if uaInfo == "\"\"" {
			uaInfo = "\"-\""
		}
		log.Println(addr, httpInfo, lrw.statusCode, refererInfo, uaInfo)
	})
}

func (serv *Server) rateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serv.rl.Take()
		next.ServeHTTP(w, r)
	})
}

func serverHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Powered-By", "https://github.com/CHTJonas/go-lg")
		next.ServeHTTP(w, r)
	})
}

func proxyMiddleware(next http.Handler) http.Handler {
	return handlers.ProxyHeaders(next)
}
