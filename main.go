package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"flickfeed/api"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var tpl *template.Template

func main() {
	tpl = template.Must(template.ParseGlob("templates/*.html"))

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	var err error

	db, err = sql.Open("mysql", dsn)

	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	defer db.Close()

	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
	
		log.Println("Waiting for database to be ready...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Error connecting to database after waiting: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS shows (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		year INT NOT NULL,
		poster VARCHAR(255) NOT NULL,
		rating FLOAT NOT NULL
	);`

	if _, err = db.Exec(createTableQuery); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/imgs/", http.StripPrefix("/imgs/", http.FileServer(http.Dir("imgs"))))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/addShow", addShowHandler)
	http.HandleFunc("/mylist", myListHandler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "index.html", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		query := r.FormValue("query")

		if query == "" {
			http.Error(w, "Search query is empty", http.StatusBadRequest)
			return
		}

		omdbAPIKey := os.Getenv("OMDB_API_KEY")

		if omdbAPIKey == "" {
			http.Error(w, "OMDB API key is not configured", http.StatusInternalServerError)
			return
		}

		omdbResp, err := api.SearchMovies(omdbAPIKey, query)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error searching movies: %v", err), http.StatusInternalServerError)
			return
		}
		
		var detailedMovies []api.Movie

		for _, m := range omdbResp.Search {
			details, err := api.GetMovieDetails(omdbAPIKey, m.ImdbID)
			if err != nil {
				log.Printf("Error fetching details for imdbID %s: %v", m.ImdbID, err)
				continue
			}

			detailedMovies = append(detailedMovies, *details)
		}

		tpl.ExecuteTemplate(w, "search_results.html", detailedMovies)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addShowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		year := r.FormValue("year")
		poster := r.FormValue("poster")
		rating := r.FormValue("rating")

		if title == "" {
			http.Error(w, "Show title is required", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO shows (title, year, poster, rating) VALUES (?, ?, ?, ?)", title, year, poster, rating)

		if err != nil {
			http.Error(w, "Error adding show", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/mylist", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func myListHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM shows")

	if err != nil {
		http.Error(w, "Error fetching shows", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	type Show struct {
		ID         int
		Title      string
		Year       string
		Poster     string
		ImdbRating string
	}

	var shows []Show

	for rows.Next() {
		var s Show
		if err := rows.Scan(&s.ID, &s.Title, &s.Year, &s.Poster, &s.ImdbRating); err != nil {
			continue
		}

		shows = append(shows, s)
	}

	tpl.ExecuteTemplate(w, "mylist.html", shows)
}
