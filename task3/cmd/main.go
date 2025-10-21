package main

import (
	"log"
	"task3/internal/app"
	"task3/internal/fetcher"
	"task3/internal/reporter"
	"time"
)

const cbApiUrl = "http://www.cbr.ru/scripts/XML_daily_eng.asp"
const daysToFetch = 90

func main() {
	client := fetcher.NewClient(cbApiUrl)
	rep := reporter.NewConsoleReporter()
	application := app.NewApp(client, rep)

	currentDate := time.Now()
	err := application.Run(daysToFetch, currentDate)
	if err != nil {
		log.Fatal(err)
	}

}
