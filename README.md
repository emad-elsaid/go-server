GO SERVER
=========

A template for Go servers. I found that I reuse some code everytime I write a go service. This is a starting point so I can clone and start new projects faster.

## Background

It started when I was [converiting a ruby service to go](https://www.emadelsaid.com/converting-Ruby-sinatra-project-to-Go/). I wanted to write the minimum that gives me same ruby sinatra framework feeling without using a framework. for that I broke all community rules while writing this code but at the end it worked and felt good to achieve this result.

## Features

- Compiles and embeds views from `views` directory
- Postgresql operations (create, drop, setup, seed, migrate up/down)
- Sinatra shorthand functions to define routes (GET, POST, DELETE)
- Session/Cookies
- Logging requests to STDOUT + response time
- Method override with `_method` param.
- Sqlc setup for converting queries to go
- Docker and docker compose setups
- Deployment script
- Backup script

## Usage

- This is meant to be cloned
- Copy `.env.sample` to `.env` and edit the values and make sure it's loaded to your environment
- Edit the `common.go` constants.
- Generate database code with [Sqlc](main) `bin/db setup`
- Run `bin/css` to download and compile css and icons
- Use `router` gorilla router or `GET`, `POST` shorthand functions...etc.

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

## Backups

a shell script in `bin/backup` will run as a service on server to backup the database everyday.

## Database

SQLX package is used and PQ package to connect to postgres database. the database URL is read from the environment.

`bin/db` is a shell script that include basic commands needed to manage database migrations. similar to `rails db` tasks. (create, seed, setup, migrate, rollback, dump schema, load schema, create_migration, drop database, reset). a small ~150 LOC.

If you don't need the database then remove:

- `db` directory
- `bin/db` file
- `go generate` line from `common.go`
- remove `sqlx` code from `common.go`

## Guidelines

- Respect no "best practice" unless there is a reason it's better for this code/server
- Reduce dependencies as much as possible
- Reduce code to the minimum
- All code is in main package no subpackages
- Should provide main features we're used to like db connection, migrations, logging. monitoring...etc
