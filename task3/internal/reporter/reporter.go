package reporter

import (
	"fmt"
	"task3/internal/model"
)

const dateLayout = "2006-01-02"

type Reporter interface {
	Report(max, min model.CurrencyRate, avg float64)
}

type ConsoleReporter struct {
}

func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

func (r *ConsoleReporter) Report(maxRate, minRate model.CurrencyRate, avg float64) {
	fmt.Printf("Максимум: %s — %.4f руб. на %s\n", maxRate.Name, maxRate.Rate, maxRate.Date.Format(dateLayout))
	fmt.Printf("Минимум: %s — %.4f руб. на %s\n", minRate.Name, minRate.Rate, minRate.Date.Format(dateLayout))
	fmt.Printf("Среднее значение курса: %.4f руб.\n", avg)
}
