package main

import (
	"fmt"
	"log"
	"task3/internal/fetcher"
	"task3/internal/parser"
	"time"
)

const cbApiUrl = "http://www.cbr.ru/scripts/XML_daily_eng.asp"
const daysToFetch = 90

func main() {
	cbClient := fetcher.NewClient(cbApiUrl)

	var totalCurrencyRates []parser.CurrencyRate

	for i := 0; i < daysToFetch; i++ {

		date := time.Now().AddDate(0, 0, -i)

		xml, err := cbClient.GetCourseByDate(date)
		if err != nil {
			log.Println("Get data from API error:", err)
			continue
		}

		parsedXml, err := parser.ParseRates(xml)
		if err != nil {
			log.Println("Parse error:", err)
			continue
		}
		totalCurrencyRates = append(totalCurrencyRates, parsedXml...)
	}

	if len(totalCurrencyRates) == 0 {
		log.Fatal("no data collected")
	}

	minRate := totalCurrencyRates[0]
	maxRate := totalCurrencyRates[0]
	var totalRate float64
	for _, rate := range totalCurrencyRates {
		if rate.Rate < minRate.Rate {
			minRate = rate
		}
		if rate.Rate > maxRate.Rate {
			maxRate = rate
		}
		totalRate += rate.Rate
	}
	fmt.Printf("Максимум: %s — %.4f руб. на %s\n", maxRate.Name, maxRate.Rate, maxRate.Date.Format("2006-01-02"))
	fmt.Printf("Минимум: %s — %.4f руб. на %s\n", minRate.Name, minRate.Rate, minRate.Date.Format("2006-01-02"))
	res := totalRate / float64(len(totalCurrencyRates))
	fmt.Printf("Среднее значение курса: %.4f руб.\n", res)
}
