package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"task3/internal/model"
)

type MockFetcher struct {
	mu      sync.Mutex
	FetchFn func(ctx context.Context, date time.Time) ([]byte, error)
	CallLog []time.Time
}

func (m *MockFetcher) GetCourseByDate(ctx context.Context, date time.Time) ([]byte, error) {
	m.mu.Lock()
	m.CallLog = append(m.CallLog, date)
	m.mu.Unlock()
	if m.FetchFn != nil {
		return m.FetchFn(ctx, date)
	}
	return []byte{}, nil
}

type MockReporter struct {
	mu         sync.Mutex
	ReportFn   func(max, min model.CurrencyRate, avg float64)
	ReportCall *reportCall
}

type reportCall struct {
	max model.CurrencyRate
	min model.CurrencyRate
	avg float64
}

func (m *MockReporter) Report(max, min model.CurrencyRate, avg float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ReportCall = &reportCall{max: max, min: min, avg: avg}
	if m.ReportFn != nil {
		m.ReportFn(max, min, avg)
	}
}

func TestApp_Run_Success(t *testing.T) {
	now := time.Date(2025, 10, 22, 0, 0, 0, 0, time.UTC)

	mockFetcher := &MockFetcher{
		FetchFn: func(_ context.Context, date time.Time) ([]byte, error) {
			return []byte(fmt.Sprintf(`<ValCurs Date="%s">
				<Valute ID="R01235">
					<NumCode>840</NumCode>
					<CharCode>USD</CharCode>
					<Nominal>1</Nominal>
					<Name>Доллар США</Name>
					<Value>75.%02d</Value>
				</Valute>
			</ValCurs>`, date.Format("02.01.2006"), 50-date.Day())), nil
		},
	}

	mockReporter := &MockReporter{}
	app := NewApp(mockFetcher, mockReporter)

	ctx := context.Background()
	err := app.Run(ctx, 3, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем, что вызывались ВСЕ нужные даты (в любом порядке)
	expectedDates := map[time.Time]bool{
		now:                   true,
		now.AddDate(0, 0, -1): true,
		now.AddDate(0, 0, -2): true,
	}

	for _, called := range mockFetcher.CallLog {
		if !expectedDates[called] {
			t.Errorf("Unexpected fetch date: %v", called)
		}
		delete(expectedDates, called)
	}
	if len(expectedDates) != 0 {
		t.Errorf("Missing fetch dates: %v", expectedDates)
	}

	// Проверка отчёта (остаётся без изменений)
	if mockReporter.ReportCall == nil {
		t.Fatal("Reporter.Report was not called")
	}

	// USD 22.10.2025 → 75.28
	// USD 21.10.2025 → 75.29
	// USD 20.10.2025 → 75.30
	// max = 75.30, min = 75.28, avg = (75.28+75.29+75.30)/3 ≈ 75.29
	expectedMax := model.CurrencyRate{
		Date: time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC),
		Name: "Доллар США",
		Rate: 75.30,
	}
	expectedMin := model.CurrencyRate{
		Date: time.Date(2025, 10, 22, 0, 0, 0, 0, time.UTC),
		Name: "Доллар США",
		Rate: 75.28,
	}
	expectedAvg := (75.28 + 75.29 + 75.30) / 3

	if mockReporter.ReportCall.max.Rate != expectedMax.Rate {
		t.Errorf("Max rate: expected %.2f, got %.2f", expectedMax.Rate, mockReporter.ReportCall.max.Rate)
	}
	if mockReporter.ReportCall.min.Rate != expectedMin.Rate {
		t.Errorf("Min rate: expected %.2f, got %.2f", expectedMin.Rate, mockReporter.ReportCall.min.Rate)
	}
	if mockReporter.ReportCall.avg != expectedAvg {
		t.Errorf("Avg rate: expected %.6f, got %.6f", expectedAvg, mockReporter.ReportCall.avg)
	}
}

func TestApp_Run_FetchError(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFn: func(_ context.Context, _ time.Time) ([]byte, error) {
			return nil, errors.New("network error")
		},
	}
	mockReporter := &MockReporter{}
	app := NewApp(mockFetcher, mockReporter)

	ctx := context.Background()
	err := app.Run(ctx, 1, time.Now())
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if mockReporter.ReportCall != nil {
		t.Error("Reporter should not be called on fetch error")
	}
}

func TestApp_Run_ParseError(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFn: func(_ context.Context, _ time.Time) ([]byte, error) {
			return []byte("invalid xml"), nil
		},
	}
	mockReporter := &MockReporter{}
	app := NewApp(mockFetcher, mockReporter)

	ctx := context.Background()
	err := app.Run(ctx, 1, time.Now())
	if err == nil {
		t.Fatal("Expected error from parser, got nil")
	}
	if mockReporter.ReportCall != nil {
		t.Error("Reporter should not be called on parse error")
	}
}

func TestApp_Run_NoDataCollected(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFn: func(_ context.Context, _ time.Time) ([]byte, error) {
			return []byte{}, nil
		},
	}
	mockReporter := &MockReporter{}
	app := NewApp(mockFetcher, mockReporter)

	ctx := context.Background()
	err := app.Run(ctx, 2, time.Now())
	if err == nil {
		t.Fatal("Expected 'no data collected' error")
	}
	if err.Error() != "no data collected after 2 days" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestApp_Run_NoRatesAfterParsing(t *testing.T) {
	mockFetcher := &MockFetcher{
		FetchFn: func(_ context.Context, _ time.Time) ([]byte, error) {
			// Валидный XML без валют
			return []byte(`<ValCurs Date="22.10.2025"></ValCurs>`), nil
		},
	}
	mockReporter := &MockReporter{}
	app := NewApp(mockFetcher, mockReporter)

	ctx := context.Background()
	err := app.Run(ctx, 1, time.Now())
	if err == nil {
		t.Fatal("Expected error")
	}
	// После фетчинга allRates остаётся пустым → ошибка "no data collected"
	if err.Error() != "no data collected after 1 days" {
		t.Errorf("Unexpected error: %v", err)
	}
	// Отчёт не должен вызываться
	if mockReporter.ReportCall != nil {
		t.Error("Reporter should not be called")
	}
}

func TestApp_Run_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockFetcher := &MockFetcher{
		FetchFn: func(ctx context.Context, _ time.Time) ([]byte, error) {
			// Имитируем "долгий" запрос, который должен быть прерван
			select {
			case <-time.After(100 * time.Millisecond):
				return []byte(`<ValCurs Date="01.01.2025"><Valute ID="R01235"><CharCode>USD</CharCode><Nominal>1</Nominal><Name>Доллар</Name><Value>75.00</Value></Valute></ValCurs>`), nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}
	mockReporter := &MockReporter{}
	app := NewApp(mockFetcher, mockReporter)

	// Запускаем в фоне
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, 5, time.Now())
	}()

	// Сразу отменяем контекст
	cancel()

	// Ждём результат
	err := <-done
	if err == nil {
		t.Fatal("Expected context cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}
