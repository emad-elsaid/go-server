GO SERVER
=========

A template for Go servers. I found that I reuse some code everytime I write a go service. This is a starting point so I can clone and start new projects faster.

## Background

It started when I was [converiting a ruby service to go](https://www.emadelsaid.com/converting-Ruby-sinatra-project-to-Go/). I wanted to write the minimum that gives me same ruby sinatra framework feeling without using a framework. for that I broke all community rules while writing this code but at the end it worked and felt good to achieve this result.

## Usage

This is meant to be cloned and edit the main.go file.

## Running

All the code is in `main` package in this directory to run it with

```
go run *.go
```

## Guidelines

- Respect no "best practice" unless there is a reason it's better for this code/server
- Reduce dependencies as much as possible
- Reduce code to the minimum
- All code is in main package no subpackages
- Should provide main features we're used to like db connection, migrations, logging. monitoring...etc
