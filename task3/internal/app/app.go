package app

import (
	"context"
	"fmt"
	"sync"
	"task3/internal/fetcher"
	"task3/internal/model"
	"task3/internal/parser"
	"task3/internal/reporter"
	"time"

	"golang.org/x/sync/errgroup"
)

const workersNum = 10

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

func (a *App) Run(ctx context.Context, daysToFetch int, now time.Time) error {
	eg, gCtx := errgroup.WithContext(ctx)
	eg.SetLimit(workersNum)
	var mu sync.Mutex
	var totalCurrencyRates []model.CurrencyRate

	for i := 0; i < daysToFetch; i++ {

		date := now.AddDate(0, 0, -i)
		eg.Go(func() error {
			xml, err := a.fetcher.GetCourseByDate(gCtx, date)
			if err != nil {
				return fmt.Errorf("failed to get course by date %v: %w", date, err)
			}

			if len(xml) == 0 {
				return nil
			}

			parsedRates, err := parser.ParseRates(xml)
			if err != nil {
				return fmt.Errorf("failed to parse rates for date %v: %w", date, err)
			}
			mu.Lock()
			defer mu.Unlock()
			totalCurrencyRates = append(totalCurrencyRates, parsedRates...)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
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
