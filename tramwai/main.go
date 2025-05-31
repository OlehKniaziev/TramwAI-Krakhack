package main

import (
	"time"
	"os/exec"
	"database/sql"
	"os"
	"log"
	"encoding/json"
	_ "github.com/lib/pq"
)

type event struct {
	Title string `json:"title"`
	Date string `json:"data"`
	Description string `json:"description"`
	Location string `json:"location"`
	Category string `json:"category"`
	Url string `json:"url"`
}

var db *sql.DB

func scrap() error {
	cmd := exec.Command("python3", "scrapper/scrapper.py")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("PETUCHON error: %s\n", err)
		return err
	}

	var jsonBytes []byte
	jsonBytes, err = os.ReadFile("data.json")
	if err != nil {
		log.Printf("File read error: %s\n", err)
		return err
	}

	var events []event
	err = json.Unmarshal(jsonBytes, &events)
	if err != nil {
		log.Printf("JSON error: %s\n", err)
		return err
	}

	for _, event := range events {
		row := db.QueryRow("INSERT INTO Events (date, title, description, location, link) VALUES ($1, $2, $3, $4, $5) RETURNING id", event.Date, event.Title, event.Description, event.Location, event.Url)

		var insertId int64
		err = row.Scan(&insertId)

		if err != nil {
			log.Printf("Last insert id error (why): %s\n", err)
			return err
		}

		_, err = db.Exec("INSERT INTO Keywords (keyword, event_id) VALUES ($1, $2)", event.Category, insertId)
		if err != nil {
			log.Printf("Keyword insert error: %s\n", err)
			return err
		}
	}

	return nil
}

func main() {
	dbConnString := os.Getenv("POSTGRES_CONN_STRING")
	var err error
	db, err = sql.Open("postgres", dbConnString)
	if err != nil {
		log.Fatalf("DB connection err: %s\n", err)
	}

	scrap()

	for _ = range time.Tick(1 * time.Hour) {
		scrap()
	}
}
