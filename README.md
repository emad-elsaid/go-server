GO SERVER
=========

A template for Go servers. I found that I reuse some code everytime I write a go service. This is a starting point so I can clone and start new projects faster.

## Background

It started when I was [converiting a ruby service to go](https://www.emadelsaid.com/converting-Ruby-sinatra-project-to-Go/). I wanted to write the minimum that gives me same ruby sinatra framework feeling without using a framework. for that I broke all community rules while writing this code but at the end it worked and felt good to achieve this result.

## Usage

- This is meant to be cloned
- Edit the main.go constants.
- Edit variable in db/deploy
- Edit docker-compose.yml file to change volumes paths
- Run bin/css to download and compile css and icons
- Copy `.env.sample` to `.env` and edit the values
- Use `router` gorilla router or `GET`, `POST` shorthand functions...etc.

## Running

All the code is in `main` package in this directory to run it with

```
go generate // in case you changed db/query.sql
go run *.go
```

## Scripts

- Deployment: a shell script in `db/deploy` can be used to deploy to your server. it needs `docker-compose`
- Backup: a shell script in `db/backup` will run as a service on server to backup the database everyday
- css compiler: a script in `db/css` will download bulma.io and fontawesome icons and compile them in one css file `public/style.css`
- database operations: `bin/db` can be used to do database operations like migrating up/down, create, setup, seed. read it for more details

## Guidelines

- Respect no "best practice" unless there is a reason it's better for this code/server
- Reduce dependencies as much as possible
- Reduce code to the minimum
- All code is in main package no subpackages
- Should provide main features we're used to like db connection, migrations, logging. monitoring...etc
