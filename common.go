//go:generate sqlc generate --file db/sqlc.yaml
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/lmittmann/tint"
	"maragu.dev/gomponents"
)

const (
	APP_NAME            = "go-server"
	STATIC_DIR_PATH     = "public"
	BIND_ADDRESS        = "0.0.0.0:3000"
	SESSION_COOKIE_NAME = APP_NAME + "_session"
	CSRF_COOKIE_NAME    = APP_NAME + "_csrf"
)

var (
	Query   *Queries
	router  = http.NewServeMux()
	session = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	CSRF    = csrf.TemplateField
)

func defaultMiddlewares() []func(http.Handler) http.Handler {
	crsfOpts := []csrf.Option{
		csrf.Path("/"),
		csrf.FieldName("csrf"),
		csrf.CookieName(CSRF_COOKIE_NAME),
	}

	sessionSecret := []byte(os.Getenv("SESSION_SECRET"))
	if len(sessionSecret) == 0 {
		sessionSecret = make([]byte, 128)
		rand.Read(sessionSecret)
	}

	middlewares := []func(http.Handler) http.Handler{
		methodOverrideMiddleware,
		csrf.Protect(sessionSecret, crsfOpts...),
		requestLoggerMiddleware,
	}

	return middlewares
}

// Some aliases to make it shorter to write handlers
type (
	Request  = *http.Request
	Response = http.HandlerFunc
)

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	Query = New(queryLogger{db})
	session.Options.HttpOnly = true
}

func Start() {
	router.HandleFunc("GET /"+STATIC_DIR_PATH+"/", staticDirectoryMiddleware())

	var handler http.Handler = router
	for _, v := range defaultMiddlewares() {
		handler = v(handler)
	}

	srv := &http.Server{
		Handler:      handler,
		Addr:         BIND_ADDRESS,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	slog.Info("Server listening", "address", BIND_ADDRESS)
	slog.Info("Server closing", "error", srv.ListenAndServe())
}

// LOGGING ===============================================

func LogDuration(msg string, args ...interface{}) func() {
	start := time.Now()

	return func() {
		slog.
			With("duration", time.Now().Sub(start)).
			With(args...).
			Debug(msg)
	}
}

// DATABASE LOGGER ===================================

type queryLogger struct {
	db *pgxpool.Pool
}

func (p queryLogger) Exec(ctx context.Context, q string, args ...interface{}) (pgconn.CommandTag, error) {
	defer LogDuration("DB Exec", "query", q, "args", args)()
	return p.db.Exec(ctx, q, args...)
}

func (p queryLogger) Query(ctx context.Context, q string, args ...interface{}) (pgx.Rows, error) {
	defer LogDuration("DB Query", "query", q, "args", args)()
	return p.db.Query(ctx, q, args...)
}

func (p queryLogger) QueryRow(ctx context.Context, q string, args ...interface{}) pgx.Row {
	defer LogDuration("DB Row", "query", q, "args", args)()
	return p.db.QueryRow(ctx, q, args...)
}

// Responses functions ==========================================

type HandlerFunc func(Request) Response

func handlerFuncToHttpHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(r)(w, r)
	}
}

func TextO(out gomponents.Node) Response {
	return func(w http.ResponseWriter, r *http.Request) {
		out.Render(w)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusNotFound)
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusUnauthorized)
}

func InternalServerError(err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Redirect(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

// ROUTES functions ==========================================

func Get(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.HandleFunc("GET "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

func Post(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.HandleFunc("POST "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

func Delete(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.HandleFunc("DELETE "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

// SESSION =================================

func Session(r *http.Request) *sessions.Session {
	s, _ := session.Get(r, SESSION_COOKIE_NAME)
	return s
}

// MIDDLEWARES ==============================

// First middleware gets executed first
func applyMiddlewares(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func staticDirectoryMiddleware() http.HandlerFunc {
	dir := http.Dir(STATIC_DIR_PATH)
	server := http.FileServer(dir)
	handler := http.StripPrefix("/"+STATIC_DIR_PATH, server)

	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		handler.ServeHTTP(w, r)
	}
}

// Derived from Gorilla middleware https://github.com/gorilla/handlers/blob/v1.5.1/handlers.go#L134
func methodOverrideMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			om := r.FormValue("_method")
			if om == "PUT" || om == "PATCH" || om == "DELETE" {
				r.Method = om
			}
		}
		h.ServeHTTP(w, r)
	})
}

func requestLoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer LogDuration(r.URL.Path, "method", r.Method)()
		h.ServeHTTP(w, r)
	})
}

// HELPERS FUNCTIONS ======================

var sha256cache = map[string]string{}

func Sha256(p string) string {
	if v, ok := sha256cache[p]; ok {
		return v
	}

	f, err := os.Open(p)
	if err != nil {
		return err.Error()
	}

	d, err := io.ReadAll(f)
	if err != nil {
		return err.Error()
	}

	sha256cache[p] = fmt.Sprintf("%x", sha256.Sum256(d))
	return sha256cache[p]
}
