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
	Keywords     []string `json:"keywords"`
	DateStart     string `json:"dateStart"`
	DateEnd        string `json:"dateEnd"`
	Preferences     []string `json:"preferences"`
}

type event struct {
	Date        string
	Title       string
	Description string
	Link        string
	Location    string
}

var db *sql.DB

const modelName = "llama3.2"

const llmGenerateUrl = "http://llm:11434/api/generate"

const dissectPromptIntro = `Ty jesteś bardzo pomocnym asystentem do wyszukiwania wydarzeń w Krakowie, użytkownik podaje Ci jakąś informację o wydarzeniu na które chce trafić, w tym jakieś słowa kluczowe, może też podać datę lub jakiś okres dat, w tym powiedzieć coś w stylu jutro czy za tydzień, masz bazę danych z eventami, która jest dosyć duża, więc chcemy rozkminić jakie są słowa kluczowe i zapisywać to do json structury `

// `Po dwukropku dostaniesz opis wydarzenia na które chce trafić użytkownik. Z tego opisu musisz zczytać interwał dat w formacie 'Rok'-'Miesiąc'-'Dzień', jeżeli użytkownik poda datę, w przeciwnym przypadku wykorzystaj 2025-05-31 jako początek i koniec interwału, oraz listę słów kluczowych jako stringi oraz dodatkowe preferencje użytkownika też jako listę stringów. Jeżeli użytkownik nie poda roku, znaczy że rok 2025. Sformatuj swoją odpowiedź w następnym formacie JSON bez żadnych komentarzy oraz formatowania - '{\"dateStart\": <początek interwalu czasowego>, \"dateEnd\": <koniec interwalu czasowego>, \"keywords\": <lista słów kluczowych>, \"preferences\": <preferencje>}'. Masz następne słowa kluczowe -`


type llmRawResponse struct {
	Response string `json:"response"`
}

type llmParams struct {
	Model string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool `json:"stream"`
}

func askLLM(prompt string) (result string, err error) {
	// var sb strings.Builder
	// for _, c := range prompt {
	// 	if c == '"' {
	// 		sb.WriteString("\\")
	// 	}
	// 	sb.WriteRune(c)
	// }

	// prompt = sb.String()

	var jsonPrompt []byte
	jsonPrompt, err = json.Marshal(&llmParams{Model: modelName, Prompt: prompt, Stream: false})

	reader := strings.NewReader(string(jsonPrompt))

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

func queryDatabaseForEvents(startDate, endDate string, keywords []string) (events []event, err error) {
	var sb strings.Builder
	for i, keyword := range keywords {
		s := fmt.Sprintf("'%s'", keyword)
		sb.WriteString(s)
		if i != len(keywords) - 1 {
			sb.WriteString(", ")
		}
	}

	keywordsString := sb.String()

	sqlString := fmt.Sprintf("SELECT date, title, description, link, location FROM Events WHERE date >= '%s' AND date <= '%s' AND id in (SELECT event_id FROM Keywords WHERE keyword in (%s))", startDate, endDate, keywordsString)
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
			location string
		)

		err = rows.Scan(&date, &title, &description, &link, &location)
		if err != nil {
			return
		}

		events = append(events, event{
			Date: date,
			Title: title,
			Description: description,
			Link: link,
			Location: location,
		})
	}

	if err = rows.Err(); err != nil {
		return
	}

	return
}

const filterPromptIntro = `Search through given events and filter them with given preferences, you should look through some parts of title and description as there might be some information provided. You should only answer in Polish and do not repeat events. Please give at least three of the relevant ones. The data is given in JSON format and comes after a colon.`

func filterEventsWithLLM(events []event, preferences []string) (resp string, err error) {
	var payload struct {
		Events []event `json:"events"`
		Preferences []string `json:"preferences"`
	}
	payload.Events = events
	payload.Preferences = preferences

	var payloadJson []byte
	payloadJson, err = json.Marshal(&payload)
	if err != nil {
		return
	}

	payloadString := fmt.Sprintf(`%s: %s`, filterPromptIntro, string(payloadJson))

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

	events, err := queryDatabaseForEvents(dissect.DateStart, dissect.DateEnd, dissect.Keywords)
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
