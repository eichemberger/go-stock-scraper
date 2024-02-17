package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
)

type Stock struct {
	company, price, change string
}

func main() {
	ticker := []string{
		"IBM",
		"AAPL",
		"MPWR",
		"QCOM",
		"V",
		"CAT",
		"CSCO",
		"NKE",
		"PYPL",
	}

	stocks := []Stock{}

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.OnHTML("div#quote-header-info", func(e *colly.HTMLElement) {
		stock := Stock{}

		stock.company = e.ChildText("h1")
		stock.price = e.ChildText("fin-streamer[data-field='regularMarketPrice']")
		stock.change = e.ChildText("fin-streamer[data-field='regularMarketChangePercent']")

		stocks = append(stocks, stock)
	})

	c.Wait()

	for _, t := range ticker {
		c.Visit("https://finance.yahoo.com/quote/" + t + "/")
	}

	file, error := os.Create("stocks.csv")

	if error != nil {
		fmt.Println("Error:", error)
		log.Fatal("Cannot create file")
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	headers := []string{"Company", "Price", "Change"}

	writer.Write(headers)

	for _, s := range stocks {
		writer.Write([]string{s.company, s.price, s.change})
	}

	defer writer.Flush()
}
