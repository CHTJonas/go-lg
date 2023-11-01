package web

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

var aboutURL string = "https://github.com/CHTJonas/go-lg"
var runtimeVer string = strings.TrimPrefix(runtime.Version(), "go")
var id = uuid.New()

func requestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Request().Header.Set("X-Request-Id", id.String())
			return next(c)
		}
	}
}

func loggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			err := next(c)
			stop := time.Now()
			if err != nil {
				c.Error(err)
			}

			dur := stop.Sub(start)
			id := req.Header.Get("X-Request-Id")
			addr := c.RealIP()
			path := req.URL.Path
			if path == "" {
				path = "/"
			}
			httpInfo := fmt.Sprintf("\"%s %s %s\"", req.Method, path, req.Proto)
			refererInfo := fmt.Sprintf("\"%s\"", req.Referer())
			if refererInfo == "\"\"" {
				refererInfo = "\"-\""
			}
			uaInfo := fmt.Sprintf("\"%s\"", req.UserAgent())
			if uaInfo == "\"\"" {
				uaInfo = "\"-\""
			}
			log.Println(id, addr, httpInfo, res.Status, res.Size, dur, refererInfo, uaInfo)

			return nil
		}
	}
}

func recoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					if r != http.ErrAbortHandler {
						debug.PrintStack()
					}
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%s", r)
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}

func serverHeaderMiddleware(version string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			pwrBy := fmt.Sprintf("go-lg/%s Go/%s (+%s)", version, runtimeVer, aboutURL)
			c.Response().Header().Set("X-Powered-By", pwrBy)
			c.Response().Header().Set("X-Robots-Tag", "noindex, nofollow")
			return next(c)
		}
	}
}

func clientRateLimitingMiddleware() echo.MiddlewareFunc {
	extractor := func(ctx echo.Context) (string, error) {
		return ctx.RealIP(), nil
	}
	return rateLimitingMiddleware(2, 4, 3*time.Minute, extractor)
}

func serverRateLimitingMiddleware() echo.MiddlewareFunc {
	extractor := func(_ echo.Context) (string, error) {
		return "localhost", nil
	}
	return rateLimitingMiddleware(10, 30, 3*time.Minute, extractor)
}

func rateLimitingMiddleware(rate rate.Limit, burst int, expiry time.Duration, extractor middleware.Extractor) echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(
		middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{
					Rate:      rate,
					Burst:     burst,
					ExpiresIn: expiry},
			),
			IdentifierExtractor: extractor,
		})
}
