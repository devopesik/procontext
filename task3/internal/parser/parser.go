package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

type ValCurs struct {
	Valutes []Valute `xml:"Valute"`
	Date    string   `xml:"Date,attr"`
}

type Valute struct {
	Name     string `xml:"Name"`
	Nominal  int    `xml:"Nominal"`
	ValueStr string `xml:"Value"`
}

type CurrencyRate struct {
	Name string
	Rate float64
	Date time.Time
}

func ParseRates(xmlData []byte) ([]CurrencyRate, error) {
	var vals ValCurs
	reader := bytes.NewReader(xmlData)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&vals)
	if err != nil {
		return nil, err
	}

	var result []CurrencyRate
	date, err := time.Parse("02.01.2006", vals.Date)
	if err != nil {
		return nil, err
	}
	for _, valute := range vals.Valutes {
		valuteStrFloat64 := strings.Replace(valute.ValueStr, ",", ".", -1)
		if valute.Nominal == 0 {
			return nil, fmt.Errorf("nominal is zero for currency %s", valute.Name)
		}
		valuteFloat64, err := strconv.ParseFloat(valuteStrFloat64, 64)
		if err != nil {
			return nil, err
		}
		valuteRate := valuteFloat64 / float64(valute.Nominal)
		result = append(result, CurrencyRate{
			Name: valute.Name,
			Rate: valuteRate,
			Date: date,
		})
	}

	return result, nil
}
