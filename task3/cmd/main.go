package main

import (
	"log"
	"task3/internal/app"
	"task3/internal/fetcher"
	"task3/internal/reporter"
)

const cbApiUrl = "http://www.cbr.ru/scripts/XML_daily_eng.asp"
const daysToFetch = 90

func main() {
	client := fetcher.NewClient(cbApiUrl)
	rep := reporter.NewConsoleReporter()
	application := app.NewApp(client, rep)

	err := application.Run(daysToFetch)
	if err != nil {
		log.Fatal(err)
	}

}
