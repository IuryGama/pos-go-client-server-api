package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

type Cambio struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	database, err := sql.Open("sqlite", "./cambio.db")
	if err != nil {
		panic(err)
	}
	defer database.Close()
	err = CreateTableCambio(database)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler(database))

	http.ListenAndServe(":8080", mux)
}

func CotacaoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		cambio, err := ConsultaCambio(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Timeout ao consultar a API de cotação: ", err)
				http.Error(w, "Timeout ao consultar cambio: "+err.Error(), http.StatusInternalServerError)
				return
			}
			log.Println("Erro ao consultar a API de cotação: ", err)
			http.Error(w, "Erro ao consultar cambio: "+err.Error(), http.StatusInternalServerError)
			return
		}

		ctxInsert, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		err = InserDatabase(ctxInsert, db, cambio)
		if err != nil {
			if ctxInsert.Err() == context.DeadlineExceeded {
				log.Println("Timeout ao inserir dado no banco de dados: ", err)
			} else {
				log.Println("Erro ao inserir dado no banco de dados: ", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(cambio)

	}
}

func ConsultaCambio(ctx context.Context) (*Cambio, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cambio Cambio
	err = json.Unmarshal(data, &cambio)
	if err != nil {
		return nil, err
	}
	return &cambio, nil
}

func CreateTableCambio(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cambio (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT,
			codein TEXT,
			name TEXT,
			high TEXT,
			low TEXT,
			var_bid TEXT,
			pct_change TEXT,
			bid TEXT,
			ask TEXT,
			timestamp TEXT,
			create_date TEXT
		)
	`)
	if err != nil {
		return err
	}
	return nil
}

func InserDatabase(ctx context.Context, db *sql.DB, cambio *Cambio) error {
	query := `
        INSERT INTO cambio (
            code,
            codein,
            name,
            high,
            low,
            var_bid,
            pct_change,
            bid,
            ask,
            timestamp,
            create_date
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err := db.ExecContext(ctx,
		query,
		cambio.USDBRL.Code,
		cambio.USDBRL.Codein,
		cambio.USDBRL.Name,
		cambio.USDBRL.High,
		cambio.USDBRL.Low,
		cambio.USDBRL.VarBid,
		cambio.USDBRL.PctChange,
		cambio.USDBRL.Bid,
		cambio.USDBRL.Ask,
		cambio.USDBRL.Timestamp,
		cambio.USDBRL.CreateDate,
	)
	if err != nil {
		return err
	}
	return nil
}
