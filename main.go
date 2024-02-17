package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gocolly/colly"
)

type Stock struct {
	company, price, change string
}

type Date struct {
	month, day, year string
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
		rawChange := e.ChildText("fin-streamer[data-field='regularMarketChangePercent']")

		change := strings.ReplaceAll(rawChange, "(", "")
		change = strings.ReplaceAll(change, ")", "")
		change = strings.ReplaceAll(change, "%", "")

		changeNumber, err := strconv.ParseFloat(change, 64)

		if err != nil {
			fmt.Println("Error parsing change %:", err)
		}

		stock.change = strconv.FormatFloat(changeNumber, 'f', 2, 64)

		stocks = append(stocks, stock)
	})

	c.Wait()

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

	s3Client := getS3Client()

	date := getDate()

	bucketName := os.Getenv("AWS_BUCKET_NAME")
	objectKey := fmt.Sprintf("%s/%s/%s/stocks.csv", date.year, date.month, date.day)

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(csvData.Bytes()),
	})

	if err != nil {
		log.Fatalf("Unable to upload CSV to S3, %v", err)
	}

	fmt.Println("Successfully uploaded CSV to S3")
}

func getDate() *Date {
	now := time.Now()

	day := fmt.Sprintf("%02d", now.Day())
	month := fmt.Sprintf("%02d", int(now.Month()))
	year := fmt.Sprintf("%d", now.Year())

	return &Date{day, month, year}
}

func getS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	return s3.NewFromConfig(cfg)
}
