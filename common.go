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
	"path"
	"strings"
	"time"

	"github.com/emad-elsaid/go-server/db"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	_ "github.com/lib/pq"
	"github.com/lmittmann/tint"
	"maragu.dev/gomponents"
)

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	wd, err := os.Getwd()
	if err != nil {
		wd = "/app"
	}

	NewApp(path.Base(wd), "0.0.0.0:3000")
}

var DefaultApp *App

type App struct {
	Name       string
	Address    string
	PublicPath string
	Mux        *http.ServeMux
	DB         *db.Queries
	Session    *sessions.CookieStore
}

func NewApp(name, address string) *App {
	DefaultApp = &App{
		Name:       name,
		Address:    address,
		PublicPath: "public",
		Mux:        http.NewServeMux(),
		Session:    sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET"))),
	}

	DefaultApp.Session.Options.HttpOnly = true

	return DefaultApp
}

func defaultMiddlewares() []func(http.Handler) http.Handler {
	csrfCookieName := DefaultApp.Name + "_csrf"

	crsfOpts := []csrf.Option{
		csrf.Path("/"),
		csrf.FieldName("csrf"),
		csrf.CookieName(csrfCookieName),
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

func Start() {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	config.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   (*db.Logger)(slog.Default()),
		LogLevel: tracelog.LogLevelInfo,
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	DefaultApp.DB = db.New(pool)
	DefaultApp.Mux.HandleFunc("GET /"+DefaultApp.PublicPath+"/", staticDirectoryMiddleware())

	var handler http.Handler = DefaultApp.Mux
	for _, v := range defaultMiddlewares() {
		handler = v(handler)
	}

	srv := &http.Server{
		Handler:      handler,
		Addr:         DefaultApp.Address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	slog.Info("Server listening", "address", DefaultApp.Address)
	slog.Info("Server closing", "error", srv.ListenAndServe())
}

// LOGGING ===============================================

func LogDuration() func(msg string, args ...interface{}) {
	start := time.Now()

	return func(msg string, args ...interface{}) {
		slog.
			With("duration", time.Now().Sub(start)).
			With(args...).
			Debug(msg)
	}
}

// Responses functions ==========================================

type HandlerFunc func(Request) Response

func handlerFuncToHttpHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(r)(w, r)
	}
}

func Ok(out gomponents.Node) Response {
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
	DefaultApp.Mux.HandleFunc("GET "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

func Post(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	DefaultApp.Mux.HandleFunc("POST "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

func Delete(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	DefaultApp.Mux.HandleFunc("DELETE "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

// SESSION =================================

func Session(r *http.Request) *sessions.Session {
	cookieName := DefaultApp.Name + "_session"
	s, _ := DefaultApp.Session.Get(r, cookieName)
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
	dir := http.Dir(DefaultApp.PublicPath)
	server := http.FileServer(dir)
	handler := http.StripPrefix("/"+DefaultApp.PublicPath, server)

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
		defer LogDuration()(r.URL.Path, "method", r.Method)
		h.ServeHTTP(w, r)
	})
}

// HELPERS FUNCTIONS ======================

var CSRF = csrf.TemplateField

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
