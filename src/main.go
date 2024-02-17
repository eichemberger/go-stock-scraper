package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/eichemberger/go-stock-scraper/src/customAWS"
	"github.com/eichemberger/go-stock-scraper/src/logger"
	"github.com/eichemberger/go-stock-scraper/src/utils"
	"github.com/gocolly/colly"
)

type Stock struct {
	company, price, change string
}

var sugar *zap.SugaredLogger

func init() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar = logger.Sugar()
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
		sugar.Debugw("Visiting",
			"url", r.URL,
		)
	})

	c.OnError(func(r *colly.Response, err error) {
		sugar.Errorw("Error",
			"url", r.Request.URL,
			"response", r,
			"error", err,
		)
	})

	c.OnHTML("div#quote-header-info", func(e *colly.HTMLElement) {
		stock := Stock{}

		stock.company = e.ChildText("h1")
		stock.price = e.ChildText("fin-streamer[data-field='regularMarketPrice']")
		rawChange := e.ChildText("fin-streamer[data-field='regularMarketChangePercent']")

		change := strings.ReplaceAll(rawChange, "(", "")
		change = strings.ReplaceAll(change, ")", "")
		change = strings.ReplaceAll(change, "%", "")

		changeNumber, err := strconv.ParseFloat(change, 64)

		if err != nil {
			sugar.Errorw("Error parsing change %",
				"error", err,
			)
		}

		stock.change = strconv.FormatFloat(changeNumber, 'f', 2, 64)

		stocks = append(stocks, stock)
	})

	c.Wait()

	logger.Sugar.Infow("Retriving stocks")

	for _, t := range ticker {
		c.Visit("https://finance.yahoo.com/quote/" + t + "/")
	}

	var csvData bytes.Buffer
	writer := csv.NewWriter(&csvData)

	headers := []string{"Company", "Price", "Change"}

	writer.Write(headers)

	for _, s := range stocks {
		writer.Write([]string{s.company, s.price, s.change})
	}

	writer.Flush()

	date := utils.GetDate()

	bucketName := os.Getenv("AWS_BUCKET_NAME")
	objectKey := fmt.Sprintf("%s/%s/%s/stocks.csv", date.Year, date.Month, date.Day)

	logger.Sugar.Infow("Uploading CSV to S3",
		"bucket", bucketName,
		"key", objectKey,
	)
	err := customAWS.S3PutObject(csvData.Bytes(), bucketName, objectKey)

	if err != nil {
		sugar.Fatalw("Unable to upload CSV to S3",
			"error", err,
			"bucket", bucketName,
			"key", objectKey,
		)
	}

	sugar.Infow("Uploaded CSV to S3",
		"bucket", bucketName,
		"key", objectKey,
	)
}
