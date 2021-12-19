package web

import (
	"log"
	"net/http"
	"strings"

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
		ip := strings.Split(r.RemoteAddr, ":")[0]
		method := r.Method
		uri := r.RequestURI
		proto := r.Proto
		log.Printf("%s \"%s %s %s\" %d\n", ip, method, uri, proto, lrw.statusCode)
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
