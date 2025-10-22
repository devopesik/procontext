package parser

import (
	"testing"
	"time"
)

func TestParseRates_Success(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="22.10.2025" name="Курс валют">
	<Valute ID="R01235">
		<NumCode>840</NumCode>
		<CharCode>USD</CharCode>
		<Nominal>1</Nominal>
		<Name>US Dollar</Name>
		<Value>75,50</Value>
	</Valute>
	<Valute ID="R01239">
		<NumCode>978</NumCode>
		<CharCode>EUR</CharCode>
		<Nominal>1</Nominal>
		<Name>Euro</Name>
		<Value>82,30</Value>
	</Valute>
</ValCurs>`)

	rates, err := ParseRates(xmlData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(rates) != 2 {
		t.Fatalf("Expected 2 rates, got %d", len(rates))
	}

	expectedDate := time.Date(2025, time.October, 22, 0, 0, 0, 0, time.UTC)

	// Проверяем USD
	if rates[0].Name != "US Dollar" {
		t.Errorf("Expected name 'US Dollar', got %q", rates[0].Name)
	}
	if rates[0].Rate != 75.50 {
		t.Errorf("Expected rate 75.50, got %.2f", rates[0].Rate)
	}
	if !rates[0].Date.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, rates[0].Date)
	}

	// Проверяем EUR
	if rates[1].Name != "Euro" {
		t.Errorf("Expected name 'Euro', got %q", rates[1].Name)
	}
	if rates[1].Rate != 82.30 {
		t.Errorf("Expected rate 82.30, got %.2f", rates[1].Rate)
	}
	if !rates[1].Date.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, rates[1].Date)
	}
}

func TestParseRates_EmptyValutes(t *testing.T) {
	xmlData := []byte(`<ValCurs Date="22.10.2025"></ValCurs>`)

	rates, err := ParseRates(xmlData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(rates) != 0 {
		t.Errorf("Expected 0 rates, got %d", len(rates))
	}
}

func TestParseRates_InvalidXML(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0"?><ValCurs><Valute><Value>75,50</Valu></ValCurs>`)

	_, err := ParseRates(xmlData)
	if err == nil {
		t.Fatal("Expected XML parsing error, got nil")
	}
}

func TestParseRates_InvalidDate(t *testing.T) {
	xmlData := []byte(`<ValCurs Date="2025-10-22"><Valute><Nominal>1</Nominal><Name>Test</Name><Value>75,50</Value></Valute></ValCurs>`)

	_, err := ParseRates(xmlData)
	if err == nil {
		t.Fatal("Expected date parsing error, got nil")
	}
}

func TestParseRates_InvalidValueFormat(t *testing.T) {
	xmlData := []byte(`<ValCurs Date="22.10.2025"><Valute><Nominal>1</Nominal><Name>Test</Name><Value>not_a_number</Value></Valute></ValCurs>`)

	_, err := ParseRates(xmlData)
	if err == nil {
		t.Fatal("Expected value parsing error, got nil")
	}
}

func TestParseRates_ZeroNominal(t *testing.T) {
	xmlData := []byte(`<ValCurs Date="22.10.2025"><Valute><Nominal>0</Nominal><Name>Test</Name><Value>75,50</Value></Valute></ValCurs>`)

	_, err := ParseRates(xmlData)
	if err == nil {
		t.Fatal("Expected error for zero nominal, got nil")
	}
	if err.Error() != "nominal is zero for currency Test" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestParseRates_CommaAsDecimalSeparator(t *testing.T) {
	xmlData := []byte(`<ValCurs Date="22.10.2025"><Valute><Nominal>10</Nominal><Name>RUB</Name><Value>100,50</Value></Valute></ValCurs>`)

	rates, err := ParseRates(xmlData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(rates) != 1 {
		t.Fatalf("Expected 1 rate, got %d", len(rates))
	}

	// 100,50 / 10 = 10.05
	expectedRate := 100.50 / 10.0
	if rates[0].Rate != expectedRate {
		t.Errorf("Expected rate %.2f, got %.2f", expectedRate, rates[0].Rate)
	}
}

func TestParseRates_MissingOptionalFields(t *testing.T) {
	// Проверяем, что парсер не падает, если нет NumCode или CharCode (они не используются)
	xmlData := []byte(`<ValCurs Date="22.10.2025"><Valute><Nominal>1</Nominal><Name>Test</Name><Value>1,00</Value></Valute></ValCurs>`)

	rates, err := ParseRates(xmlData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(rates) != 1 || rates[0].Rate != 1.0 {
		t.Errorf("Unexpected result: %+v", rates)
	}
}
