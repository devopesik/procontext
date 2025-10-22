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
	allRates, err := a.fetchAllRates(ctx, daysToFetch, now)
	if err != nil {
		return fmt.Errorf("failed to fetch rates: %w", err)
	}

	if len(allRates) == 0 {
		return fmt.Errorf("no data collected after %d days", daysToFetch)
	}

	err = a.calculateAndReport(allRates)
	if err != nil {
		return fmt.Errorf("failed to calculate and report: %w", err)
	}
	return nil
}

func (a *App) fetchAllRates(ctx context.Context, daysToFetch int, now time.Time) (map[time.Time][]model.CurrencyRate, error) {
	eg, gCtx := errgroup.WithContext(ctx)
	eg.SetLimit(workersNum)

	var mu sync.Mutex
	allRates := make(map[time.Time][]model.CurrencyRate)

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

			if len(parsedRates) == 0 {
				return nil
			}

			valDate := parsedRates[0].Date

			mu.Lock()
			allRates[valDate] = parsedRates
			mu.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return allRates, nil
}

func (a *App) calculateAndReport(allRates map[time.Time][]model.CurrencyRate) error {
	var minRate, maxRate model.CurrencyRate
	var totalRate float64
	totalRateLen := 0

	for _, ratesForDay := range allRates {
		for _, r := range ratesForDay {
			if totalRateLen == 0 {
				minRate = r
				maxRate = r
			}
			if r.Rate < minRate.Rate {
				minRate = r
			}
			if r.Rate > maxRate.Rate {
				maxRate = r
			}
			totalRate += r.Rate
			totalRateLen++
		}
	}

	if totalRateLen == 0 {
		return fmt.Errorf("no rate data found to calculate statistics")
	}

	avgRub := totalRate / float64(totalRateLen)
	a.reporter.Report(maxRate, minRate, avgRub)
	return nil
}
