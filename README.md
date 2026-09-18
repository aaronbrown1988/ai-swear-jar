# Claude Says Jar

A single-binary virtual swear jar built with Go, HTMX, Pico CSS, and SQLite.

## Run locally

```sh
go run .
```

Open `http://localhost:8080`. Each click records a $5 contribution in `jar.db`.

## Build a deployable binary

```sh
go build -o claude-says-jar .
```

Templates, Pico CSS, custom CSS, and HTMX are embedded into the binary. By default the SQLite database is stored at `jar.db` beside the process working directory. Set `JAR_DB` to select another database path and `JAR_ADDR` to change the listening address.
