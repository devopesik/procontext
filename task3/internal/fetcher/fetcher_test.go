package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetCourseByDate_Success(t *testing.T) {
	expectedBody := `<ValCurs Date="22/10/2025" name="Official exchange rate">...</ValCurs>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("date_req") != "22/10/2025" {
			t.Errorf("Expected date_req=22/10/2025, got %s", r.URL.Query().Get("date_req"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	fetcher := NewClient(server.URL)

	ctx := context.Background()
	date := time.Date(2025, time.October, 22, 0, 0, 0, 0, time.UTC)

	body, err := fetcher.GetCourseByDate(ctx, date)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(body) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, string(body))
	}
}

func TestGetCourseByDate_BadStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	fetcher := NewClient(server.URL)

	ctx := context.Background()
	date := time.Now()

	_, err := fetcher.GetCourseByDate(ctx, date)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "bad status code") {
		t.Errorf("Expected 'bad status code' in error, got %v", err)
	}
}

func TestGetCourseByDate_ReadBodyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		// Намеренно не пишем тело → при чтении будет ошибка (например, unexpected EOF)
	}))
	defer server.Close()

	fetcher := NewClient(server.URL)

	ctx := context.Background()
	date := time.Now()

	_, err := fetcher.GetCourseByDate(ctx, date)
	if err == nil {
		t.Fatal("Expected error due to incomplete body, got nil")
	}
}

func TestGetCourseByDate_RequestError(t *testing.T) {
	// Используем недопустимый URL, чтобы вызвать ошибку при создании запроса
	fetcher := NewClient("://invalid-url")

	ctx := context.Background()
	date := time.Now()

	_, err := fetcher.GetCourseByDate(ctx, date)
	if err == nil {
		t.Fatal("Expected error from invalid URL, got nil")
	}
}

func TestGetCourseByDate_HttpClientError(t *testing.T) {
	// Используем несуществующий хост
	fetcher := &cbClient{
		baseURL: "http://nonexistent.local",
		httpClient: &http.Client{
			Timeout: 100 * time.Millisecond, // короткий таймаут для быстрого завершения
		},
	}

	ctx := context.Background()
	date := time.Now()

	_, err := fetcher.GetCourseByDate(ctx, date)
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
}

func TestNewClient_ReturnsInterface(t *testing.T) {
	fetcher := NewClient("http://example.com")
	if fetcher == nil {
		t.Fatal("NewClient returned nil")
	}
	_, ok := fetcher.(CurrencyRateFetcher)
	if !ok {
		t.Fatal("NewClient did not return CurrencyRateFetcher")
	}
}
