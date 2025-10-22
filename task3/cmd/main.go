package main

import (
	"context"
	"flag"
	"log"
	"task3/internal/app"
	"task3/internal/fetcher"
	"task3/internal/reporter"
	"time"
)

var (
	apiUrl      = flag.String("api-url", "http://www.cbr.ru/scripts/XML_daily_eng.asp", "URL of Central Bank API")
	daysToFetch = flag.Int("days", 90, "Number of days to fetch")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := fetcher.NewClient(*apiUrl)
	rep := reporter.NewConsoleReporter()
	application := app.NewApp(client, rep)

	currentDate := time.Now()
	err := application.Run(ctx, *daysToFetch, currentDate)
	if err != nil {
		log.Fatal(err)
	}

}
