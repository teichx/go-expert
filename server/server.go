package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DollarBid struct {
	Bid string `json:"bid"`
}

type Response struct {
	USDBRL DollarBid `json:"USDBRL"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", FetchDollarExchangeRate)
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func FetchDollarExchangeRate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resultJson string
	var resultStatus int
	var resultError error
	defer func() {
		log.Default().Printf(`{"status": %d, "response": %s, "error": %v}`, resultStatus, resultJson, resultError)
		w.WriteHeader(resultStatus)
		w.Write([]byte(resultJson))
	}()
	select {
	case <-r.Context().Done():
		resultText := `{"error": "Request canceled by client"}`
		log.Default().Printf("[%d] %s - %v", http.StatusRequestTimeout, resultText, r.Context().Err())
		http.Error(w, resultText, http.StatusRequestTimeout)
		return
	default:
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			resultJson = `{"error": "Failed to create request"}`
			resultStatus = http.StatusInternalServerError
			resultError = err
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			resultJson = `{"error": "Request timed out or failed"}`
			resultStatus = http.StatusGatewayTimeout
			resultError = err
			return
		}
		defer resp.Body.Close()
		var jsonApiResponse Response
		err = json.NewDecoder(resp.Body).Decode(&jsonApiResponse)
		if err != nil {
			resultJson = `{"error": "Failed to decode JSON response"}`
			resultStatus = http.StatusBadGateway
			resultError = err
			return
		}
		routeResponse := DollarBid{
			Bid: jsonApiResponse.USDBRL.Bid,
		}
		body, err := json.Marshal(routeResponse)
		if err != nil {
			resultJson = `{"error": "Failed to marshal response body"}`
			resultStatus = http.StatusBadGateway
			resultError = err
			return
		}
		ctx, cancel = context.WithTimeout(r.Context(), 10*time.Millisecond)
		defer cancel()

		statement, err := db.PrepareContext(ctx, "INSERT INTO cotacoes(bid) VALUES(?)")
		if err != nil {
			resultJson = `{"error": "Failed to prepare statement"}`
			resultStatus = http.StatusBadGateway
			resultError = err
			return
		}
		defer statement.Close()

		_, err = statement.ExecContext(ctx, routeResponse.Bid)
		if err != nil {
			resultJson = `{"error": "Failed to execute statement"}`
			resultStatus = http.StatusBadGateway
			resultError = err
			return
		}
		resultJson = string(body)
		resultStatus = http.StatusOK
	}
}
