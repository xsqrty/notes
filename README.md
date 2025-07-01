# Simple notes app (template)

[![Deploy Status](https://github.com/xsqrty/notes/actions/workflows/deploy.yml/badge.svg)](https://github.com/xsqrty/notes/actions/workflows/deploy.yml)

## Install dependencies

* Install task [task manager](https://taskfile.dev/installation/)

```shell
npm install -g @go-task/cli
```

*or*

```shell
pip install go-task-bin
```

* Install dependencies

```shell
task install
```

> for get list of task commands run `task --list-all`

* Download go.mod deps

```shell
go mod download
```

## Configuration

> Configure environment

* DSN=postgres_connection_string (postgres://postgres:@127.0.0.1:5432/db?sslmode=disable)

## Build

```shell
task build:dev
```

*build with version*

```shell
task build:v1.0.0
```

*binary: /bin/notes*

## Migrations

* Create new migration

```shell
task migrate:create:migation_name
```

* Check version

```shell
task migrate:version
```

* Migrate up

```shell
task migrate:up
```

* Migrate downgrade

```shell
task migrate:down
```

## Run

```shell
task run
```

*or*

```shell
go run cmd/notes.go
```

*run tests*

```shell
task test
```

## Docs

* Generate docs

```shell
task swag:generate
```

Run service `task run` and open in browser http://localhost:1323

* Format swagger comments

```shell
task swag:fmt
```