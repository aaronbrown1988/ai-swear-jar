package main

import (
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

const contributionCents = 500

//go:embed templates/*.html static/*
var embeddedFiles embed.FS

type application struct {
	db        *sql.DB
	templates *template.Template
}

type jarSummary struct {
	Count       int
	TotalCents  int
	TotalAmount string
}

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	templates, err := template.ParseFS(embeddedFiles, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	app := &application{db: db, templates: templates}
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(mustSubFS(embeddedFiles, "static")))))
	mux.HandleFunc("GET /", app.home)
	mux.HandleFunc("POST /contributions", app.addContribution)

	address := envOrDefault("JAR_ADDR", ":8080")
	log.Printf("Virtual Swear Jar listening on http://localhost%s", address)
	log.Fatal(http.ListenAndServe(address, mux))
}

func openDatabase() (*sql.DB, error) {
	path := envOrDefault("JAR_DB", "jar.db")
	if directory := filepath.Dir(path); directory != "." {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS contributions (
			id INTEGER PRIMARY KEY,
			amount_cents INTEGER NOT NULL,
			created_at TEXT NOT NULL
		)`); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	summary, err := app.summary()
	if err != nil {
		http.Error(w, "Unable to load the jar.", http.StatusInternalServerError)
		return
	}
	render(w, app.templates, "page", summary)
}

func (app *application) addContribution(w http.ResponseWriter, r *http.Request) {
	_, err := app.db.Exec(
		"INSERT INTO contributions (amount_cents, created_at) VALUES (?, ?)",
		contributionCents,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		http.Error(w, "Unable to update the jar.", http.StatusInternalServerError)
		return
	}

	summary, err := app.summary()
	if err != nil {
		http.Error(w, "Unable to load the jar.", http.StatusInternalServerError)
		return
	}
	render(w, app.templates, "summary", summary)
}

func (app *application) summary() (jarSummary, error) {
	var summary jarSummary
	if err := app.db.QueryRow(
		"SELECT COUNT(*), COALESCE(SUM(amount_cents), 0) FROM contributions",
	).Scan(&summary.Count, &summary.TotalCents); err != nil {
		return jarSummary{}, err
	}
	summary.TotalAmount = "$" + strconv.Itoa(summary.TotalCents/100) + "." + twoDigits(summary.TotalCents%100)
	return summary, nil
}

func render(w http.ResponseWriter, templates *template.Template, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

func mustSubFS(files embed.FS, directory string) fs.FS {
	sub, err := fs.Sub(files, directory)
	if err != nil {
		panic(err)
	}
	return sub
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func twoDigits(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
