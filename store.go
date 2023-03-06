package main

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	DB *sql.DB
}

const (
	CrawlerJobToDo       = "TODO"
	CrawlerJobInProgress = "IN_PROGRESS"
	CrawlerJobDone       = "DONE"
)

func exec(db *sql.DB, stmt string) sql.Result {
	res, err := db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func initDB(db *sql.DB) {
	exec(db, `
		CREATE TABLE IF NOT EXISTS recipes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			scraper_name TEXT NOT NULL,
			url TEXT NOT NULL,
			recipe TEXT NOT NULL
		)
	`)
	exec(db, `
		CREATE TABLE IF NOT EXISTS crawler_jobs (
			url TEXT PRIMARY KEY NOT NULL,
			scraper_name TEXT NOT NULL,
			status TEXT NOT NULL
		)
	`)
}

func (store *Store) SaveRecipe(scraperName string, recipe *Recipe) {
	recipeBytes, err := json.Marshal(recipe)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := store.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(`
		INSERT INTO recipes(scraper_name, url, recipe)
		VALUES (?, ?, ?)
	`, scraperName, recipe.URL, string(recipeBytes))
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(`
		UPDATE crawler_jobs
		SET status = ?
		WHERE url = ? AND scraper_name = ?
	`, CrawlerJobDone, recipe.URL, scraperName)
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (store *Store) AddCrawlerJobs(scraperName string, sourceUrl string, urls []string) {
	tx, err := store.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(`
		UPDATE crawler_jobs
		SET status = ?
		WHERE url = ? AND scraper_name = ?
	`, CrawlerJobDone, sourceUrl, scraperName)
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO crawler_jobs (url, scraper_name, status)
		VALUES (?, ?, ?)
		ON CONFLICT(url) DO NOTHING
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err = stmt.Exec(url, scraperName, CrawlerJobToDo)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (store *Store) CrawlerJobByStatus(scraperName string, status string, count int) []string {
	var rows *sql.Rows
	var err error
	if count < 0 {
		rows, err = store.DB.Query(`
			SELECT url
			FROM crawler_jobs
			WHERE status = ? AND scraper_name = ?
		`, status, scraperName)
	} else {
		rows, err = store.DB.Query(`
			SELECT url
			FROM crawler_jobs
			WHERE status = ? AND scraper_name = ?
			LIMIT ?
		`, status, scraperName, count)
	}
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	res := []string{}

	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, url)
	}

	return res
}

func (store *Store) AllRecipes(recipes chan<- *Record[Recipe]) {
	rows, err := store.DB.Query(`
		SELECT id, scraper_name, url, recipe
		FROM recipes
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var scraperName string
		var url string
		var recipeStr string
		err = rows.Scan(&id, &scraperName, &url, &recipeStr)
		if err != nil {
			log.Fatal(err)
		}

		var recipe Recipe
		err = json.Unmarshal([]byte(recipeStr), &recipe)
		if err != nil {
			log.Fatal(err)
		}

		recipes <- &Record[Recipe]{
			Id:     id,
			Record: recipe,
		}
	}

	close(recipes)
}

func (store *Store) UpdateCrawlerJobStatus(scraperName string, urls []string, status string) {
	if len(urls) == 0 {
		return
	}

	tx, err := store.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	for _, url := range urls {
		_, err = tx.Exec(`
			UPDATE crawler_jobs
			SET status = ?
			WHERE url = ? AND scraper_name = ?
		`, status, url, scraperName)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (store *Store) Todo(scraperName string, count int) []string {
	todo := store.CrawlerJobByStatus(scraperName, CrawlerJobToDo, count)
	store.UpdateCrawlerJobStatus(scraperName, todo, CrawlerJobInProgress)
	return todo
}

func NewStore(dbPath string) *Store {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	initDB(db)

	return &Store{DB: db}
}

func (store *Store) Close() {
	store.DB.Close()
}
