// fetchmarket is a standalone tool to fetch market data from Yahoo Finance
// and save it to a JSON file that can be checked into the repository.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

const defaultOutputFile = "market_data.json"

// MarketData stores historical annual returns
type MarketData struct {
	LastUpdated string             `json:"last_updated"`
	VOO         map[string]float64 `json:"voo"`    // Year -> Annual return % (S&P 500)
	QQQ         map[string]float64 `json:"qqq"`    // Year -> Annual return % (Nasdaq 100)
	VTI         map[string]float64 `json:"vti"`    // Year -> Annual return % (Total Stock Market)
	BND         map[string]float64 `json:"bnd"`    // Year -> Annual return % (Total Bond Market)
}

// YahooChartResponse represents the JSON response from Yahoo Finance chart API
type YahooChartResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Adjclose []struct {
					Adjclose []float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// fetchYahooFinanceData fetches historical price data from Yahoo Finance using chart API
func fetchYahooFinanceData(ticker string, startDate, endDate time.Time) ([][]string, error) {
	// Convert to Unix timestamps
	period1 := startDate.Unix()
	period2 := endDate.Unix()

	// Build URL using chart API (more reliable than download endpoint)
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d",
		ticker, period1, period2)

	// Create request with headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo finance returned status %d", resp.StatusCode)
	}

	// Parse JSON
	var chartResp YahooChartResponse
	err = json.NewDecoder(resp.Body).Decode(&chartResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned")
	}

	result := chartResp.Chart.Result[0]
	timestamps := result.Timestamp
	adjCloses := result.Indicators.Adjclose[0].Adjclose

	if len(timestamps) != len(adjCloses) {
		return nil, fmt.Errorf("data length mismatch")
	}

	// Convert to CSV format: Date, Adj Close
	records := [][]string{{"Date", "Adj Close"}}
	for i, ts := range timestamps {
		date := time.Unix(ts, 0).Format("2006-01-02")
		adjClose := fmt.Sprintf("%.6f", adjCloses[i])
		records = append(records, []string{date, adjClose})
	}

	return records, nil
}

// calculateAnnualReturns calculates annual returns from daily price data
func calculateAnnualReturns(records [][]string) (map[string]float64, error) {
	if len(records) < 2 {
		return nil, fmt.Errorf("insufficient data")
	}

	// Skip header row
	records = records[1:]

	// Group by year and get first/last prices
	type yearData struct {
		firstPrice float64
		lastPrice  float64
	}
	yearPrices := make(map[string]*yearData)

	for _, record := range records {
		if len(record) < 2 {
			continue
		}

		// Parse date (format: YYYY-MM-DD)
		date := record[0]
		if len(date) < 4 {
			continue
		}
		year := date[:4]

		// Parse adjusted close price (column 1)
		adjClose, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			continue
		}

		// Initialize year data if needed
		if yearPrices[year] == nil {
			yearPrices[year] = &yearData{firstPrice: adjClose, lastPrice: adjClose}
		}

		// Update last price (data is in chronological order)
		yearPrices[year].lastPrice = adjClose
	}

	// Calculate annual returns
	returns := make(map[string]float64)
	for year, data := range yearPrices {
		if data.firstPrice > 0 {
			returnPct := ((data.lastPrice - data.firstPrice) / data.firstPrice) * 100
			returns[year] = returnPct
		}
	}

	return returns, nil
}

func main() {
	outputFile := flag.String("o", defaultOutputFile, "Output JSON file path")
	years := flag.Int("years", 11, "Number of years of data to fetch")
	flag.Parse()

	fmt.Println("Fetching market data from Yahoo Finance...")

	md := &MarketData{
		VOO: make(map[string]float64),
		QQQ: make(map[string]float64),
		VTI: make(map[string]float64),
		BND: make(map[string]float64),
	}

	// Fetch data for specified years
	startDate := time.Now().AddDate(-*years, 0, 0)
	endDate := time.Now()

	// Define tickers to fetch
	tickers := []struct {
		symbol string
		target *map[string]float64
		name   string
	}{
		{"VOO", &md.VOO, "S&P 500 (VOO)"},
		{"QQQ", &md.QQQ, "Nasdaq 100 (QQQ)"},
		{"VTI", &md.VTI, "Total Stock Market (VTI)"},
		{"BND", &md.BND, "Total Bond Market (BND)"},
	}

	// Fetch each ticker
	for i, ticker := range tickers {
		fmt.Printf("  Fetching %s...\n", ticker.name)

		records, err := fetchYahooFinanceData(ticker.symbol, startDate, endDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %v\n", ticker.symbol, err)
			os.Exit(1)
		}

		returns, err := calculateAnnualReturns(records)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error calculating %s returns: %v\n", ticker.symbol, err)
			os.Exit(1)
		}

		// Update data with new returns
		for year, ret := range returns {
			(*ticker.target)[year] = ret
		}

		// Wait a bit to avoid rate limiting (except on last iteration)
		if i < len(tickers)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	// Set last updated
	md.LastUpdated = time.Now().Format("2006-01-02")

	// Save to file
	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(*outputFile, data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nMarket data saved to %s\n", *outputFile)

	// Print summary
	fmt.Println("\nSummary:")
	printSummary(md)
}

func printSummary(md *MarketData) {
	// Get sorted years
	years := make([]string, 0)
	for year := range md.VOO {
		years = append(years, year)
	}
	sort.Strings(years)

	currentYear := time.Now().Year()
	var vooSum, qqqSum, vtiSum, bndSum float64
	count := 0

	fmt.Printf("%-8s %10s %10s %10s %10s %12s\n", "Year", "VOO", "QQQ", "VTI", "BND", "60/40")
	fmt.Println("------------------------------------------------------------------------")

	for _, year := range years {
		vooRet := md.VOO[year]
		qqqRet := md.QQQ[year]
		vtiRet := md.VTI[year]
		bndRet := md.BND[year]
		mix6040 := vtiRet*0.6 + bndRet*0.4

		yearInt, _ := strconv.Atoi(year)
		if yearInt < currentYear {
			vooSum += vooRet
			qqqSum += qqqRet
			vtiSum += vtiRet
			bndSum += bndRet
			count++
		}

		fmt.Printf("%-8s %9.2f%% %9.2f%% %9.2f%% %9.2f%% %11.2f%%\n",
			year, vooRet, qqqRet, vtiRet, bndRet, mix6040)
	}

	if count > 0 {
		avgMix := (vtiSum/float64(count))*0.6 + (bndSum/float64(count))*0.4
		fmt.Println("------------------------------------------------------------------------")
		fmt.Printf("%-8s %9.2f%% %9.2f%% %9.2f%% %9.2f%% %11.2f%%\n",
			"Average", vooSum/float64(count), qqqSum/float64(count),
			vtiSum/float64(count), bndSum/float64(count), avgMix)
	}
}
