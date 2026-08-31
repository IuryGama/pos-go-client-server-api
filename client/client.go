package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Bid struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Fatal("Timeout ao consultar a API de cotação: ", err)
			panic(err)
		}
		log.Fatal("Erro ao consultar a API de cotação: ", err)
		panic(err)
	}
	defer res.Body.Close()

	var b Bid
	err = json.NewDecoder(res.Body).Decode(&b)
	if err != nil {
		panic(err)
	}

	err = SalvarArquivo(b)
	if err != nil {
		panic(err)
	}
}

func SalvarArquivo(bid Bid) error {
	arquivo, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer arquivo.Close()

	_, err = arquivo.WriteString("Dólar: " + bid.USDBRL.Bid)
	if err != nil {
		return err
	}
	return nil
}
