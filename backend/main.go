package main

import (
	"fmt"
	"strings"
	"os"
	"log"
	"net/http"
	"database/sql"
	"io"
	"encoding/json"
	_ "github.com/lib/pq"
)

type llmPromptDissect struct {
	Keywords []string `json:"keywords"`
	Date     string `json:"date"`
}

type event struct {
	Date        string
	Title       string
	Description string
	Link        string
}

var db *sql.DB

const modelName = "llama3.2"

const llmGenerateUrl = "http://llm:11434/api/generate"

const promptIntro = `Po dwukropku dostaniesz opis wydarzenia na które chce trafić użytkownik. Z tego opisu musisz zczytać datę w formacie 'Dzień'.'Miesiąc'.'Rok' oraz listę słów kluczowych. Jeżeli użytkownik nie poda roku, znaczy że rok 2025. Sformatuj swoją odpowiedź w następnym formacie JSON bez żadnych komentarzy oraz formatowania - '{\"date\": <data wydarzenia>, \"keywords\": <lista słów kluczowych>}'.`

type llmRawResponse struct {
	Response string `json:"response"`
}

func sendPromptToLLM(prompt string) (dissect llmPromptDissect, err error) {
	jsonPrompt := fmt.Sprintf(`{"model": "%s", "prompt": "%s: %s", "stream": false}`, modelName, promptIntro, prompt)
	reader := strings.NewReader(jsonPrompt)

	var resp *http.Response
	resp, err = http.Post(llmGenerateUrl, "application/json", reader)
	if err != nil {
		return
	}

	var rawResp llmRawResponse

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&rawResp)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(rawResp.Response), &dissect)
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
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("%+v\n", dissect)

	events, err := queryDatabaseForEvents(dissect.Date, dissect.Keywords)
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
	var err error
	dbConnString := os.Getenv("POSTGRES_CONN_STRING")
	if len(dbConnString) == 0 {
		log.Fatalln("Empty POSTGRES_CONN_STRING env var")
	}

	db, err = sql.Open("postgres", dbConnString)
	if err != nil {
		log.Fatalln(err)
	}

	portString := os.Getenv("HTTP_PORT")
	if len(portString) == 0 {
		log.Fatalln("Empty HTTP_PORT env var")
	}
	portString = ":" + portString

	http.HandleFunc("POST /prompt", promptHandler)
	if err = http.ListenAndServe(portString, nil); err != nil {
		log.Fatalln(err)
	}
}
