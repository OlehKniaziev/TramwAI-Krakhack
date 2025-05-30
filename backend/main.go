package main

import (
	"fmt"
	"strings"
	"os"
	"log"
	"net/http"
	"database/sql"
	"io"
	_ "github.com/lib/pq"
)

type llmPromptDissect struct {
	keywords []string
	date     string
}

type event struct {
	Date        string
	Title       string
	Description string
	Link        string
}

var db *sql.DB

func sendPromptToLLM(prompt string) (response llmPromptDissect, err error) {
	log.Fatal("TODO")
	return
}

func queryDatabaseForEvents(date string, keywords []string) (events []event, err error) {
	var sb strings.Builder
	for i, keyword := range keywords {
		s := fmt.Sprintf("\"%s\"", keyword)
		sb.WriteString(s)
		if i != len(keywords) - 1 {
			sb.WriteString(", ")
		}
	}

	keywordsString := sb.String()

	sqlString := fmt.Sprintf("SELECT date, title, description, link FROM Events WHERE date LIKE \"%s\" AND event_id in (SELECT event_id WHERE keyword in (%s))", date, keywordsString)
	var rows *sql.Rows
	rows, err = db.Query(sqlString)
	if err != nil {
		return
	}

	events = make([]event, 0, 100)

	for rows.Next() {
		var (
			date string
			title string
			description string
			link string
		)

		err = rows.Scan(&date, &title, &description, &link)
		if err != nil {
			return
		}

		events = append(events, event{
			Date: date,
			Title: title,
			Description: description,
			Link: link,
		})
	}

	if err = rows.Err(); err != nil {
		return
	}

	return
}

func filterEventsWithLLM(events []event) (resp string, err error) {
	log.Fatal("TODO")
	return
}

func promptHandler(w http.ResponseWriter, r *http.Request) {
	promptBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prompt := string(promptBytes)
	dissect, err := sendPromptToLLM(prompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	events, err := queryDatabaseForEvents(dissect.date, dissect.keywords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := filterEventsWithLLM(events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(response))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	contentsBytes, err := os.ReadFile(".env")
	contents := string(contentsBytes)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(contents, "\n")
	for _, envPair := range lines {
		pair := strings.Split(envPair, "=")
		err = os.Setenv(pair[0], pair[1])
		if err != nil {
			log.Fatal(err)
		}
	}

	dbConnString := os.Getenv("POSTGRES_CONN_STRING")
	if len(dbConnString) == 0 {
		log.Fatal("Fuck you")
	}

	db, err = sql.Open("postgres", dbConnString)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("POST /prompt", promptHandler)
	if err = http.ListenAndServe(":42069", nil); err != nil {
		log.Fatal(err)
	}
}
