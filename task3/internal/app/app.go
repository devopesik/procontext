package app

import (
	"fmt"
	"log"
	"task3/internal/fetcher"
	"task3/internal/model"
	"task3/internal/parser"
	"task3/internal/reporter"
	"time"
)

type App struct {
	fetcher  fetcher.CurrencyRateFetcher
	reporter reporter.Reporter
}

func NewApp(fetcher fetcher.CurrencyRateFetcher, reporter reporter.Reporter) *App {
	return &App{
		fetcher:  fetcher,
		reporter: reporter,
	}
}

func (a *App) Run(daysToFetch int) error {
	var totalCurrencyRates []model.CurrencyRate

	for i := 0; i < daysToFetch; i++ {

		date := time.Now().AddDate(0, 0, -i)

		xml, err := a.fetcher.GetCourseByDate(date)
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
		return fmt.Errorf("no data collected after %d days", daysToFetch)
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
	avgRub := totalRate / float64(len(totalCurrencyRates))
	a.reporter.Report(maxRate, minRate, avgRub)
	return nil
}
