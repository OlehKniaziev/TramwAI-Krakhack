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
	Preferences     []string `json:"preferences"`
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

const dissectPromptIntro = `Po dwukropku dostaniesz opis wydarzenia na które chce trafić użytkownik. Z tego opisu musisz zczytać datę w formacie 'Dzień'.'Miesiąc'.'Rok' oraz listę słów kluczowych jako stringi oraz dodatkowe preferencje użytkownika też jako listę stringów. Jeżeli użytkownik nie poda roku, znaczy że rok 2025. Sformatuj swoją odpowiedź w następnym formacie JSON bez żadnych komentarzy oraz formatowania - '{\"date\": <data wydarzenia>, \"keywords\": <lista słów kluczowych>, \"preferences\": <preferencje>}'. Masz następne słowa kluczowe -`

type llmRawResponse struct {
	Response string `json:"response"`
}

func askLLM(prompt string) (result string, err error) {
	jsonPrompt := fmt.Sprintf(`{"model": "%s", "prompt": "%s", "stream": false}`, modelName, prompt)
	reader := strings.NewReader(jsonPrompt)

	var resp *http.Response
	resp, err = http.Post(llmGenerateUrl, "application/json", reader)
	if err != nil {
		return
	}

	var rawResp llmRawResponse

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&rawResp)
	result = rawResp.Response
	return
}

func sendPromptToLLM(prompt string) (dissect llmPromptDissect, err error) {
	var rows *sql.Rows
	rows, err = db.Query("SELECT DISTINCT keyword from Keywords")
	if err != nil {
		return
	}

	keywords := make([]string, 0, 69)

	defer rows.Close()
	for rows.Next() {
		var (
			keyword string
		)

		if err = rows.Scan(&keyword); err != nil {
			return
		}

		keywords = append(keywords, keyword)
	}

	if err = rows.Err(); err != nil {
		return
	}

	var sb strings.Builder

	for i, keyword := range keywords {
		sb.WriteString(keyword)

		if i != len(keywords) - 1 {
			sb.WriteString(", ")
		}
	}

	keywordsString := sb.String()

	prompt = fmt.Sprintf(`%s %s: %s`, dissectPromptIntro, keywordsString, prompt)

	var dissectString string
	dissectString, err = askLLM(prompt)
	if err != nil {
		return
	}

	fmt.Printf("dissect string: %s\n", dissectString)

	err = json.Unmarshal([]byte(dissectString), &dissect)
	return
}

func queryDatabaseForEvents(date string, keywords []string) (events []event, err error) {
	var sb strings.Builder
	for i, keyword := range keywords {
		s := fmt.Sprintf("'%s'", keyword)
		sb.WriteString(s)
		if i != len(keywords) - 1 {
			sb.WriteString(", ")
		}
	}

	keywordsString := sb.String()

	sqlString := fmt.Sprintf("SELECT date, title, description, link FROM Events WHERE date LIKE '%s' AND id in (SELECT event_id FROM Keywords WHERE keyword in (%s))", date, keywordsString)
	fmt.Println(sqlString)
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

const filterPromptIntro = `Search through given events and filter them with given preferences, you should look through some parts of title and description as there might be some information provided. You should only answer in Polish and do not repeat events. Please give at least three of the relevant ones. The data is given in JSON format and comes after a colon.`

func filterEventsWithLLM(events []event, preferences []string) (resp string, err error) {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, event := range events {
		eventString := fmt.Sprintf(`{\"date\": \"%s\", \"title\": \"%s\", \"description\": \"%s\", \"link\": \"%s\"}`, event.Date, event.Title, event.Description, event.Link)
		sb.WriteString(eventString)
		if i != len(events) - 1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune(']')

	eventsString := sb.String()

	sb.Reset()

	sb.WriteRune('[')
	for i, preference := range preferences {
		p := fmt.Sprintf(`\"%s\"`, preference)
		sb.WriteString(p)
		if i != len(preferences) - 1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune(']')

	preferencesString := sb.String()

	payloadString := fmt.Sprintf(`%s: {\"events\": %s, \"preferences\": %s}`, filterPromptIntro, eventsString, preferencesString)

	log.Printf("paylod string: %s\n", payloadString)

	resp, err = askLLM(payloadString)
	log.Printf("llm responded: %s\n", resp)
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

	response, err := filterEventsWithLLM(events, dissect.Preferences)
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
