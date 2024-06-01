GO SERVER
=========

A template for Go servers. I found that I reuse some code everytime I write a go service. This is a starting point so I can clone and start new projects faster.

## Background

It started when I was [converting a ruby service to go](https://www.emadelsaid.com/converting-Ruby-sinatra-project-to-Go/). I wanted to write the minimum that gives me same ruby sinatra framework feeling without using a framework. for that I broke all community rules while writing this code but at the end it worked and felt good to achieve this result.

## Features

- Compiles and embeds views from `views` directory
- Postgresql operations (create, drop, setup, seed, migrate up/down)
- Sinatra shorthand functions to define routes (GET, POST, DELETE)
- Session/Cookies
- Logging requests to STDOUT + response time
- Logging DB queries + execution time
- Method override with `_method` param.
- Sqlc setup for converting queries to go
- Docker and docker compose setups
- Deployment script
- Backup script

## Dependencies

- Go
- Postgresql
- [SQLc](https://sqlc.dev/)
- Docker, Docker-compose (if you want to deploy with docker)
- [Sass](https://sass-lang.com/install)
- wget
- unzip

## Usage

- This is meant to be cloned
- Edit `.env` values and make sure it's loaded to your environment
- Edit the `common.go` constants.
- Generate database code with [Sqlc](main) `bin/db setup`
- Run `bin/css` to download and compile css and icons
- Use `router` gorilla router or `GET`, `POST` shorthand functions...etc.

## Routes

`common.go` has couple functions to modify the `http.Handler` struct.

The most generic one is `ROUTE` which defines a function that gets executed if all `RouteCheck`s functions are true.
```go
func ROUTE(route http.HandlerFunc, checks ...RouteCheck)
```
`RouteCheck` function is a function that takes the request and returns true if the request should be executed with that handler function.

`GET`, `POST`, `DELETE` functions defined a route to a handler based on request path.

```go
func GET(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc)
func POST(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc)
func DELETE(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc)
```

for example, this will match only the GET requests to `/` path and execute the function.

```go
GET("/", func(w Response, r Request) Output {
    return Render("layout", "index", Locals{})
})
```

You can also pass middleware functions as the last parameter to execute them in order before the handler function.


## Handler functions

The code started by a simple `net/http.HandlerFunc` but this interface doesn't have a return type. so to redirect and return for example that costs you two lines. to have more compact handler functions I created another interface

```go
type HandlerFunc func(http.ResponseWriter, *http.Request) http.HandlerFunc
```

Which then can be converted to `net/http.HandlerFunc` itself using `handlerFuncToHttpHandler` function.

Also `http.ResponseWriter` and `*http.Request` are aliased to `Response` and `Request` so your function should look like so


```go
func Users(w Response, r Request) Output {
  return Render("layout", "index", Locals{})
}
```

returned function `Output` is an alias to `http.HandlerFunc` and there are couple functions that can be used as return response like `Redirect, Render, NotFound, BadRequest, Unauthorized, InternalServerError` `Redirect` function for example is defined as follows


```go
func Redirect(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	}
}
```

## Logging

Method `Log` can be used to log a line with label.
```go
Log(DEBUG, "View", name)
```


Method `LogDuration` can be used to log execution time a colored label and a string. adding this line to a function will print something like

```go
defer LogDuration(DEBUG, "View", name)()
```

`LogDuration` will return a function that when executed by `defer` it knows the time it was created an the time it's executed and will print that time difference + `View` label colored with `DEBUG` color and the value of `name` at the end.

```
16:30:39  View  (40.916µs) index
```

There are two colors defined so far `DEBUG` and `INFO` these are constants that defines shell escape characters for coloring the following text.


## Views

views are `.html` files under `views` directory. they're embedded to the program with go `embed` package and parsed with `html/template` package before starting the server.

Rendering the view can be done using `Render` function as the example above.


## Helpers

helper functions are passed while parsing the views files. you can use `HELPER` function to add more helpers. here is an example

```go
// helper to check if list include a string
HELPER("include", func(list []string, str string) bool {
    for _, i := range list {
        if i == str {
            return true
        }
    }

    return false
})
```

## Session

the code depends on `gorilla/session` . `SESSION` function returns an instance of the current request session.

## Assets

`bin/css` is a shell script that will download bulma.io and fontawesome and compile them into one css file written under `public/style.css` and will copy fontawesome fonts to `public/fonts`.

`public` directory is served if there is no matching route for the request.

everytime you change `bin/css` or any css file that it imports you'll need to run it again to generate the new `style.css` file.

the layout include the `style.css` file and will compute the `sha` hash and include it as part of the url to force cache flushing when the file changes.

## Running

All the code is in `main` package in this directory.

```
go generate // in case you changed db/query.sql
go run *.go
```

## Deployment

a Dockerfile is included and `docker-compose.yml` to run a database, server and backup script containers.

- Edit variables in `db/deploy`
- Edit `docker-compose.yml` file to change volumes paths
- run `bin/deploy master user@server` to deploy master branch to server with ssh

## Database

SQLX package is used and PQ package to connect to postgres database. the database URL is read from the environment.

`bin/db` is a shell script that include basic commands needed to manage database migrations. similar to `rails db` tasks. (create, seed, setup, migrate, rollback, dump schema, load schema, create_migration, drop database, reset). a small ~150 LOC.

If you don't need the database then remove:

- `db` directory
- `bin/db` file
- `go generate` line from `common.go`
- remove `sqlx` code from `common.go`

## Backups

a shell script in `bin/backup` will run as a service on server to backup the database everyday.

## Guidelines

- Respect no "best practice" unless there is a reason it's better for this code/server
- Reduce dependencies as much as possible
- Reduce code to the minimum
- All code is in main package no subpackages
- Should provide main features we're used to like db connection, migrations, logging. monitoring...etc
