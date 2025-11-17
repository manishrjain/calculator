package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const marketDataFile = ".rentobuy_market_data.json"

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
			Timestamp []int64 `json:"timestamp"`
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

// loadMarketData loads cached market data from file
func loadMarketData() (*MarketData, error) {
	data, err := os.ReadFile(marketDataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &MarketData{
				VOO: make(map[string]float64),
				QQQ: make(map[string]float64),
				VTI: make(map[string]float64),
				BND: make(map[string]float64),
			}, nil
		}
		return nil, err
	}

	var md MarketData
	err = json.Unmarshal(data, &md)
	if err != nil {
		return nil, err
	}

	if md.VOO == nil {
		md.VOO = make(map[string]float64)
	}
	if md.QQQ == nil {
		md.QQQ = make(map[string]float64)
	}
	if md.VTI == nil {
		md.VTI = make(map[string]float64)
	}
	if md.BND == nil {
		md.BND = make(map[string]float64)
	}

	return &md, nil
}

// saveMarketData saves market data to cache file
func saveMarketData(md *MarketData) error {
	md.LastUpdated = time.Now().Format("2006-01-02")

	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(marketDataFile, data, 0644)
}

// updateMarketData fetches and updates market data if needed
func updateMarketData() (*MarketData, error) {
	md, err := loadMarketData()
	if err != nil {
		return nil, fmt.Errorf("failed to load cache: %v", err)
	}

	// Check if we need to update
	now := time.Now()
	needsUpdate := false

	// Update if cache is older than 1 day
	if md.LastUpdated != "" {
		lastUpdate, err := time.Parse("2006-01-02", md.LastUpdated)
		if err == nil {
			if now.Sub(lastUpdate) > 24*time.Hour {
				needsUpdate = true
			}
		}
	} else {
		needsUpdate = true
	}

	// Also update if we don't have current year data
	currentYear := fmt.Sprintf("%d", now.Year())
	if _, ok := md.VOO[currentYear]; !ok {
		needsUpdate = true
	}

	if !needsUpdate {
		return md, nil
	}

	fmt.Println("Updating market data from Yahoo Finance...")

	// Fetch data for last 11 years (to ensure we have complete 10 years)
	startDate := time.Now().AddDate(-11, 0, 0)
	endDate := time.Now()

	// Define tickers to fetch
	tickers := []struct {
		symbol string
		target *map[string]float64
	}{
		{"VOO", &md.VOO},
		{"QQQ", &md.QQQ},
		{"VTI", &md.VTI},
		{"BND", &md.BND},
	}

	// Fetch each ticker
	for i, ticker := range tickers {
		records, err := fetchYahooFinanceData(ticker.symbol, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s data: %v", ticker.symbol, err)
		}

		returns, err := calculateAnnualReturns(records)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate %s returns: %v", ticker.symbol, err)
		}

		// Update cache with new data
		for year, ret := range returns {
			(*ticker.target)[year] = ret
		}

		// Wait a bit to avoid rate limiting (except on last iteration)
		if i < len(tickers)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	// Save to cache
	err = saveMarketData(md)
	if err != nil {
		return nil, fmt.Errorf("failed to save cache: %v", err)
	}

	fmt.Println("Market data updated successfully.")

	return md, nil
}

// calculateMarketAverages calculates 10-year averages for all ETFs
func calculateMarketAverages(md *MarketData) (voo, qqq, vti, bnd, mix6040 float64) {
	if md == nil {
		return 0, 0, 0, 0, 0
	}

	var vooSum, qqqSum, vtiSum, bndSum float64
	count := 0

	currentYear := time.Now().Year()

	for year, vooRet := range md.VOO {
		yearInt, _ := strconv.Atoi(year)
		// Only include complete years (not current year) from last 10 years
		if yearInt >= currentYear-10 && yearInt < currentYear {
			// Only count years where we have all data
			if qqqRet, ok := md.QQQ[year]; ok {
				if vtiRet, ok := md.VTI[year]; ok {
					if bndRet, ok := md.BND[year]; ok {
						vooSum += vooRet
						qqqSum += qqqRet
						vtiSum += vtiRet
						bndSum += bndRet
						count++
					}
				}
			}
		}
	}

	if count == 0 {
		return 0, 0, 0, 0, 0
	}

	voo = vooSum / float64(count)
	qqq = qqqSum / float64(count)
	vti = vtiSum / float64(count)
	bnd = bndSum / float64(count)
	mix6040 = vti*0.6 + bnd*0.4

	return
}

// displayMarketData shows historical returns and averages
func displayMarketData(md *MarketData) {
	fmt.Println("\n=== MARKET DATA ===")

	// Get sorted years
	years := make([]string, 0)
	for year := range md.VOO {
		// Only show last 10 complete years
		yearInt, _ := strconv.Atoi(year)
		if yearInt >= time.Now().Year()-10 {
			years = append(years, year)
		}
	}
	sort.Strings(years)

	// Display table
	fmt.Printf("\n%-10s %-10s %-10s %-10s %-10s %-12s\n", "Period", "VOO", "QQQ", "VTI", "BND", "60/40 Mix")
	fmt.Println(strings.Repeat("-", 68))

	var vooSum, qqqSum, vtiSum, bndSum float64
	count := 0

	for _, year := range years {
		vooRet := md.VOO[year]
		qqqRet := md.QQQ[year]
		vtiRet := md.VTI[year]
		bndRet := md.BND[year]
		mix6040 := vtiRet*0.6 + bndRet*0.4

		// Only include in average if it's a complete year (not current year)
		if year != fmt.Sprintf("%d", time.Now().Year()) {
			vooSum += vooRet
			qqqSum += qqqRet
			vtiSum += vtiRet
			bndSum += bndRet
			count++
		}

		fmt.Printf("MRKT   %-6s %-10s %-10s %-10s %-10s %-12s\n", year,
			fmt.Sprintf("%.2f%%", vooRet),
			fmt.Sprintf("%.2f%%", qqqRet),
			fmt.Sprintf("%.2f%%", vtiRet),
			fmt.Sprintf("%.2f%%", bndRet),
			fmt.Sprintf("%.2f%%", mix6040))
	}

	if count > 0 {
		avgMix := (vtiSum/float64(count))*0.6 + (bndSum/float64(count))*0.4
		fmt.Println(strings.Repeat("-", 68))
		fmt.Printf("MRKT   %-6s %-10s %-10s %-10s %-10s %-12s\n", "Avg",
			fmt.Sprintf("%.2f%%", vooSum/float64(count)),
			fmt.Sprintf("%.2f%%", qqqSum/float64(count)),
			fmt.Sprintf("%.2f%%", vtiSum/float64(count)),
			fmt.Sprintf("%.2f%%", bndSum/float64(count)),
			fmt.Sprintf("%.2f%%", avgMix))
	}
}
