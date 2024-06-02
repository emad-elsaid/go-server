//go:generate sqlc generate --file db/sqlc.yaml
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

const (
	APP_NAME            = "go-server"
	STATIC_DIR_PATH     = "public"
	BIND_ADDRESS        = "0.0.0.0:3000"
	VIEWS_EXTENSION     = ".html"
	SESSION_COOKIE_NAME = APP_NAME + "_session"
	CSRF_COOKIE_NAME    = APP_NAME + "_csrf"
)

var (
	Q       *Queries
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
	Response = http.ResponseWriter
	Request  = *http.Request
	Output   = http.HandlerFunc
	Locals   map[string]interface{} // passed to views/templates
)

func init() {
	log.SetFlags(log.Ltime)

	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	Q = New(queryLogger{db})
	session.Options.HttpOnly = true
}

func START() {
	compileViews()
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

const (
	DEBUG = "\033[97;41m"
	INFO  = "\033[97;44m"
)

func LogDuration(level, label, text string, args ...interface{}) func() {
	start := time.Now()
	return func() {
		if len(args) > 0 {
			log.Printf("%s %-6s \033[0m (%s) %s %v", level, label, time.Now().Sub(start), text, args)
		} else {
			log.Printf("%s %-6s \033[0m (%s) %s", level, label, time.Now().Sub(start), text)
		}
	}
}

// DATABASE LOGGER ===================================

type queryLogger struct {
	db *pgxpool.Pool
}

func (p queryLogger) Exec(ctx context.Context, q string, args ...interface{}) (pgconn.CommandTag, error) {
	defer LogDuration(DEBUG, "DB Exec", q, args)()
	return p.db.Exec(ctx, q, args...)
}

func (p queryLogger) Query(ctx context.Context, q string, args ...interface{}) (pgx.Rows, error) {
	defer LogDuration(DEBUG, "DB Query", q, args)()
	return p.db.Query(ctx, q, args...)
}

func (p queryLogger) QueryRow(ctx context.Context, q string, args ...interface{}) pgx.Row {
	defer LogDuration(DEBUG, "DB Row", q, args)()
	return p.db.QueryRow(ctx, q, args...)
}

// Responses functions ==========================================

type HandlerFunc func(http.ResponseWriter, *http.Request) http.HandlerFunc

func handlerFuncToHttpHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)(w, r)
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

func GET(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.HandleFunc("GET "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

func POST(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.HandleFunc("POST "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

func DELETE(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.HandleFunc("DELETE "+path,
		applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...),
	)
}

// VIEWS ====================

//go:embed views
var views embed.FS
var templates *template.Template
var helpers = template.FuncMap{}

func compileViews() {
	templates = template.New("")
	fs.WalkDir(views, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, VIEWS_EXTENSION) && d.Type().IsRegular() {
			name := strings.TrimPrefix(path, "views/")
			name = strings.TrimSuffix(name, VIEWS_EXTENSION)
			defer LogDuration(DEBUG, "View", name)()

			c, err := fs.ReadFile(views, path)
			if err != nil {
				return err
			}

			template.Must(templates.New(name).Funcs(helpers).Parse(string(c)))
		}

		return nil
	})
}

func partial(path string, data interface{}) string {
	v := templates.Lookup(path)
	if v == nil {
		return fmt.Sprintf("view %s not found", path)
	}

	w := bytes.NewBufferString("")
	err := v.Execute(w, data)
	if err != nil {
		return "rendering error " + path + " " + err.Error()
	}

	return w.String()
}

func Render(path string, view string, data Locals) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data["view"] = view
		data["request"] = r
		fmt.Fprint(w, partial(path, data))
	}
}

func HELPER(name string, f interface{}) {
	if _, ok := helpers[name]; ok {
		log.Fatalf("Helper: %s has been defined already", name)
	}

	helpers[name] = f
}

// SESSION =================================

func SESSION(r *http.Request) *sessions.Session {
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
		defer LogDuration(INFO, r.Method, r.URL.Path)()
		h.ServeHTTP(w, r)
	})
}

// HELPERS FUNCTIONS ======================

func atoi32(s string) int32 {
	i, _ := strconv.ParseInt(s, 10, 32)
	return int32(i)
}

func atoi64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func NullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  len(s) > 0,
	}
}
