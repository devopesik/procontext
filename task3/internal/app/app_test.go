package app

import (
	"task3/internal/model"
	"testing"
	"time"
)

type mockFetcher struct {
	data map[time.Time][]byte // дата → XML-данные
}

func (m *mockFetcher) GetCourseByDate(date time.Time) ([]byte, error) {
	if data, ok := m.data[date]; ok {
		return data, nil
	}
	return nil, nil
}

type mockReporter struct {
	max, min model.CurrencyRate
	avg      float64
	called   bool
}

func (m *mockReporter) Report(max, min model.CurrencyRate, avg float64) {
	m.max = max
	m.min = min
	m.avg = avg
	m.called = true
}

func TestApp_Run(t *testing.T) {
	testNow := time.Date(2025, 10, 21, 0, 0, 0, 0, time.UTC)

	day1 := testNow.AddDate(0, 0, -1)
	day2 := testNow

	xmlDay1 := []byte(`
        <ValCurs Date="20.10.2025">
            <Valute ID="R01235">
                <Name>US Dollar</Name>
                <Nominal>1</Nominal>
                <Value>95,0000</Value>
            </Valute>
            <Valute ID="R01239">
                <Name>Euro</Name>
                <Nominal>1</Nominal>
                <Value>105,0000</Value>
            </Valute>
        </ValCurs>
    `)

	xmlDay2 := []byte(`
        <ValCurs Date="21.10.2025">
            <Valute ID="R01235">
                <Name>US Dollar</Name>
                <Nominal>1</Nominal>
                <Value>96,0000</Value>
            </Valute>
            <Valute ID="R01239">
                <Name>Euro</Name>
                <Nominal>1</Nominal>
                <Value>104,0000</Value>
            </Valute>
        </ValCurs>
    `)

	fetcher := &mockFetcher{
		data: map[time.Time][]byte{
			day1: xmlDay1,
			day2: xmlDay2,
		},
	}
	reporter := &mockReporter{}

	app := NewApp(fetcher, reporter)

	err := app.Run(2, testNow) // 2 дня
	if err != nil {
		t.Fatalf("App.Run() failed: %v", err)
	}

	if !reporter.called {
		t.Fatal("Reporter.Report was not called")
	}

	if reporter.max.Rate != 105.0 {
		t.Errorf("Expected max rate 105.0, got %f", reporter.max.Rate)
	}
	if reporter.min.Rate != 95.0 {
		t.Errorf("Expected min rate 95.0, got %f", reporter.min.Rate)
	}
	// Среднее: (95 + 105 + 96 + 104) / 4 = 100.0
	if reporter.avg != 100.0 {
		t.Errorf("Expected avg 100.0, got %f", reporter.avg)
	}
}
