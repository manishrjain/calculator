package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var reader = bufio.NewReader(os.Stdin)
var savedDefaults map[string]string
var currentInputs map[string]string
var useDefaults bool
var fullNumbers bool

// Global arrays for monthly costs
var monthlyBuyingCosts []float64
var monthlyRentingCosts []float64
var remainingLoanBalance []float64
var cumulativePrincipalPaid []float64
var cumulativeInterestPaid []float64
var appreciationRates []float64 // Annual appreciation rates

// Config holds all input parameters
type Config struct {
	// Economic
	inflationRate float64
	include30Year float64

	// Buying
	purchasePrice float64
	downpayment   float64
	loanAmount    float64
	annualRate    float64
	totalMonths   int
	monthlyRate   float64
	monthlyLoanPayment float64
	annualInsurance float64
	annualTaxes     float64
	monthlyExpenses float64
	totalMonthlyBuyingCost float64

	// Renting
	rentDeposit            float64
	monthlyRent            float64
	annualRentCosts        float64
	otherAnnualCosts       float64
	investmentReturnRate   float64
	totalMonthlyRentingCost float64

	// Selling
	includeSelling   float64
	agentCommission  float64
	stagingCosts     float64
	taxFreeLimit     float64
	capitalGainsTax  float64
}

var config Config

const inputsFile = ".rentobuy_inputs.json"

func main() {
	// Clear screen
	// fmt.Print("\033[H\033[2J")

	// Parse command line flags
	flag.BoolVar(&useDefaults, "defaults", false, "Use all previously saved default values without prompting")
	flag.BoolVar(&fullNumbers, "full-numbers", false, "Display full numbers instead of compact K/M format")
	flag.Parse()

	// Update market data (blocking to ensure we have it for display)
	marketData, err := updateMarketData()
	if err != nil {
		fmt.Println("Warning: Could not fetch market data:", err)
		// Continue anyway with empty market data
		marketData = &MarketData{
			VOO: make(map[string]float64),
			QQQ: make(map[string]float64),
			VTI: make(map[string]float64),
			BND: make(map[string]float64),
		}
	}

	// Load previous inputs
	savedDefaults = loadInputs()
	currentInputs = make(map[string]string)

	// If not using defaults, show interactive form
	if !useDefaults {
		values, err := RunInteractiveForm(savedDefaults, marketData)
		if err != nil {
			fmt.Println("Form cancelled or error:", err)
			return
		}
		currentInputs = values
		// Save the inputs for next time
		saveInputs(currentInputs)
	} else {
		// Check if we have defaults when --defaults flag is used
		if len(savedDefaults) == 0 {
			fmt.Println("Error: --defaults flag used but no saved defaults found. Run without the flag first.")
			return
		}
		// Use saved defaults
		currentInputs = savedDefaults
	}

	// Parse all inputs from currentInputs
	inflationRate, err := getFloatValue("inflation_rate")
	if err != nil {
		fmt.Println("Invalid inflation rate")
		return
	}

	// Get 30-year projection toggle
	include30Year, err := getFloatValue("include_30year")
	if err != nil {
		include30Year = 0 // Default to 10-year projections only
	}

	purchasePrice, err := getFloatValue("purchase_price")
	if err != nil || purchasePrice == 0 {
		fmt.Println("Invalid purchase price - cannot be zero")
		return
	}

	loanAmount, err := getFloatValue("loan_amount")
	if err != nil {
		fmt.Println("Invalid loan amount")
		return
	}

	// Calculate downpayment
	downpayment := purchasePrice - loanAmount

	var annualRate float64
	var totalMonths int
	var monthlyRate float64
	var monthlyLoanPayment float64

	if loanAmount <= 0 {
		fmt.Println("\nNo loan needed. Purchase can be made with downpayment.")
		annualRate = 0
		totalMonths = 0
		monthlyRate = 0
		monthlyLoanPayment = 0
	} else {
		// Get loan rate
		annualRate, err = getFloatValue("loan_rate")
		if err != nil {
			fmt.Println("Invalid loan rate")
			return
		}

		// Get loan duration
		totalMonths, err = getIntValue("loan_duration", parseDuration)
		if err != nil {
			fmt.Println("Invalid duration format:", err)
			return
		}

		// Calculate monthly payment for buying
		monthlyRate = annualRate / 100 / 12
		monthlyLoanPayment = calculateMonthlyPayment(loanAmount, monthlyRate, totalMonths)
	}

	// Get all remaining values
	annualInsurance, err := getFloatValue("annual_insurance")
	if err != nil {
		fmt.Println("Invalid insurance amount")
		return
	}

	annualTaxes, err := getFloatValue("annual_taxes")
	if err != nil {
		fmt.Println("Invalid taxes amount")
		return
	}

	totalAnnualExpenses := annualInsurance + annualTaxes

	monthlyExpenses, err := getFloatValue("monthly_expenses")
	if err != nil {
		fmt.Println("Invalid monthly expenses")
		return
	}

	// Parse appreciation rates (comma-separated)
	appreciationRateStr := currentInputs["appreciation_rate"]
	appreciationRates, err = parseAppreciationRates(appreciationRateStr)
	if err != nil {
		fmt.Println("Invalid appreciation rate:", err)
		return
	}

	rentDeposit, err := getFloatValue("rent_deposit")
	if err != nil {
		fmt.Println("Invalid deposit amount")
		return
	}

	monthlyRent, err := getFloatValue("monthly_rent")
	if err != nil {
		fmt.Println("Invalid monthly rent")
		return
	}

	annualRentCosts, err := getFloatValue("annual_rent_costs")
	if err != nil {
		fmt.Println("Invalid annual rent costs")
		return
	}

	otherAnnualCosts, err := getFloatValue("other_annual_costs")
	if err != nil {
		fmt.Println("Invalid other annual costs")
		return
	}

	investmentReturnRate, err := getFloatValue("investment_return_rate")
	if err != nil {
		fmt.Println("Invalid investment return rate")
		return
	}

	// Get selling analysis parameters
	includeSelling, err := getFloatValue("include_selling")
	if err != nil {
		includeSelling = 0 // Default to not including selling analysis
	}

	var agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax float64
	if includeSelling > 0 {
		agentCommission, err = getFloatValue("agent_commission")
		if err != nil {
			fmt.Println("Invalid agent commission")
			return
		}

		stagingCosts, err = getFloatValue("staging_costs")
		if err != nil {
			fmt.Println("Invalid staging costs")
			return
		}

		taxFreeLimit, err = getFloatValue("tax_free_limit")
		if err != nil {
			fmt.Println("Invalid tax-free limit")
			return
		}

		capitalGainsTax, err = getFloatValue("capital_gains_tax")
		if err != nil {
			fmt.Println("Invalid capital gains tax rate")
			return
		}
	}

	// Calculate monthly recurring expenses
	monthlyRecurringExpenses := (totalAnnualExpenses / 12) + monthlyExpenses
	totalMonthlyBuyingCost := monthlyLoanPayment + monthlyRecurringExpenses

	// Calculate monthly cost for renting
	monthlyRentingExpenses := (annualRentCosts / 12) + (otherAnnualCosts / 12)
	totalMonthlyRentingCost := monthlyRent + monthlyRentingExpenses

	// Populate global config struct
	config = Config{
		inflationRate:              inflationRate,
		include30Year:              include30Year,
		purchasePrice:              purchasePrice,
		downpayment:                downpayment,
		loanAmount:                 loanAmount,
		annualRate:                 annualRate,
		totalMonths:                totalMonths,
		monthlyRate:                monthlyRate,
		monthlyLoanPayment:         monthlyLoanPayment,
		annualInsurance:            annualInsurance,
		annualTaxes:                annualTaxes,
		monthlyExpenses:            monthlyExpenses,
		totalMonthlyBuyingCost:     totalMonthlyBuyingCost,
		rentDeposit:                rentDeposit,
		monthlyRent:                monthlyRent,
		annualRentCosts:            annualRentCosts,
		otherAnnualCosts:           otherAnnualCosts,
		investmentReturnRate:       investmentReturnRate,
		totalMonthlyRentingCost:    totalMonthlyRentingCost,
		includeSelling:             includeSelling,
		agentCommission:            agentCommission,
		stagingCosts:               stagingCosts,
		taxFreeLimit:               taxFreeLimit,
		capitalGainsTax:            capitalGainsTax,
	}

	// Populate global cost arrays for projections (360 months = 30 years max)
	populateMonthlyCosts(360, monthlyLoanPayment, monthlyRecurringExpenses, totalMonths, totalMonthlyRentingCost, loanAmount, monthlyRate, inflationRate)

	// Display input parameters
	displayInputParameters(marketData)

	// Display market data after input parameters
	displayMarketData(marketData)

	// Display projections
	displayExpenditureTable()

	if config.loanAmount > 0 {
		displayAmortizationTable()
	}

	if config.includeSelling > 0 {
		displaySaleProceeds()
	}

	displayComparisonTable()
}

// getFloatValue gets a float value from currentInputs
func getFloatValue(key string) (float64, error) {
	input := currentInputs[key]
	value, err := parseAmount(input)
	return value, err
}

// getIntValue gets an int value from currentInputs with a parser
func getIntValue(key string, parser func(string) (int, error)) (int, error) {
	input := currentInputs[key]
	value, err := parser(input)
	return value, err
}

// loadInputs loads previously saved inputs from file
func loadInputs() map[string]string {
	data, err := os.ReadFile(inputsFile)
	if err != nil {
		return make(map[string]string)
	}

	var inputs map[string]string
	err = json.Unmarshal(data, &inputs)
	if err != nil {
		return make(map[string]string)
	}

	return inputs
}

// saveInputs saves current inputs to file for next run
func saveInputs(inputs map[string]string) {
	data, err := json.Marshal(inputs)
	if err != nil {
		return
	}

	os.WriteFile(inputsFile, data, 0644)
}

// parseAmount parses currency amounts with k, M, B suffixes
// Returns 0 for empty input
// Also handles % sign (strips it out)
func parseAmount(input string) (float64, error) {
	input = strings.ToLower(strings.TrimSpace(input))

	// Handle empty input - default to 0
	if input == "" {
		return 0, nil
	}

	// Remove % sign if present (for percentage inputs like "-10%")
	input = strings.TrimSuffix(input, "%")
	input = strings.TrimSpace(input)

	// Check for suffix
	multiplier := 1.0
	numStr := input

	if strings.HasSuffix(input, "k") {
		multiplier = 1000.0
		numStr = strings.TrimSuffix(input, "k")
	} else if strings.HasSuffix(input, "m") {
		multiplier = 1000000.0
		numStr = strings.TrimSuffix(input, "m")
	} else if strings.HasSuffix(input, "b") {
		multiplier = 1000000000.0
		numStr = strings.TrimSuffix(input, "b")
	}

	// Parse the numeric part
	value, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
	if err != nil {
		return 0, err
	}

	return value * multiplier, nil
}

// parseAppreciationRates parses comma-separated appreciation rates
// Returns array where each entry corresponds to a year, with the last entry applying to all future years
func parseAppreciationRates(input string) ([]float64, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return []float64{0}, nil
	}

	// Split by comma
	parts := strings.Split(input, ",")
	rates := make([]float64, 0, len(parts))

	for _, part := range parts {
		rate, err := parseAmount(part)
		if err != nil {
			return nil, fmt.Errorf("invalid rate '%s': %v", strings.TrimSpace(part), err)
		}
		rates = append(rates, rate)
	}

	if len(rates) == 0 {
		return []float64{0}, nil
	}

	return rates, nil
}

// getStringInputAndParse prompts the user and applies a parser function
func getStringInputAndParse(prompt string, parser func(string) (int, error)) (int, error) {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return parser(strings.TrimSpace(input))
}

// displayTable displays a formatted table with title and optional notes
func displayTable(title string, rows [][]string, notes string, highlightLastRow bool) {
	re := lipgloss.NewRenderer(os.Stdout)

	// Title style
	titleStyle := re.NewStyle().Foreground(MonokaiPink).Bold(true)

	// Table styles
	headerStyle := re.NewStyle().Padding(0, 1).Foreground(MonokaiCyan).Bold(true)
	rowStyle := re.NewStyle().Padding(0, 1).Foreground(MonokaiAdaptiveText)

	// Print title
	fmt.Println()
	fmt.Println(titleStyle.Render(title))

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(MonokaiBorder)).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			if row == 0 || (highlightLastRow && row == len(rows)-1) {
				// Header row and optionally last row
				style = headerStyle
			} else {
				style = rowStyle
			}

			// Right-align all number columns (col > 0)
			if col > 0 {
				style = style.Align(lipgloss.Right)
			}

			return style
		})

	fmt.Println(t)

	// Print notes if provided
	if notes != "" {
		noteStyle := re.NewStyle().Width(100).Italic(true).Foreground(MonokaiGrey).PaddingLeft(2)
		fmt.Println(noteStyle.Render(notes))
	}
}

// formatCurrency formats a number as currency with K/M suffixes (compact) or full format
func formatCurrency(amount float64) string {
	// Handle negative numbers
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	// If fullNumbers flag is set, use full format with dollar sign and commas
	if fullNumbers {
		// Format with 1 decimal place (automatically rounds)
		formatted := fmt.Sprintf("%.1f", amount)
		parts := strings.Split(formatted, ".")

		// Add commas to the integer part
		intPart := parts[0]
		var result strings.Builder
		for i, digit := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result.WriteRune(',')
			}
			result.WriteRune(digit)
		}

		return fmt.Sprintf("%s$%s.%s", sign, result.String(), parts[1])
	}

	// Default: compact format with K/M suffixes, no dollar sign (automatically rounds)
	var formatted string
	if amount >= 1000000 {
		// Millions
		formatted = fmt.Sprintf("%.1fM", amount/1000000)
	} else if amount >= 1000 {
		// Thousands
		formatted = fmt.Sprintf("%.1fK", amount/1000)
	} else {
		// Less than 1000
		formatted = fmt.Sprintf("%.1f", amount)
	}

	return sign + formatted
}

// formatNumber formats an integer with commas
func formatNumber(num int) string {
	numStr := strconv.Itoa(num)
	var result strings.Builder

	for i, digit := range numStr {
		if i > 0 && (len(numStr)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	return result.String()
}

// parseDuration parses duration strings like "5y6m", "30y", "6m"
func parseDuration(duration string) (int, error) {
	duration = strings.ToLower(duration)
	years := 0
	months := 0

	// Find 'y' for years
	yIndex := strings.Index(duration, "y")
	if yIndex != -1 {
		yearStr := duration[:yIndex]
		var err error
		years, err = strconv.Atoi(yearStr)
		if err != nil {
			return 0, fmt.Errorf("invalid year format")
		}
		duration = duration[yIndex+1:]
	}

	// Find 'm' for months
	mIndex := strings.Index(duration, "m")
	if mIndex != -1 {
		monthStr := duration[:mIndex]
		var err error
		months, err = strconv.Atoi(monthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid month format")
		}
	}

	totalMonths := years*12 + months
	if totalMonths <= 0 {
		return 0, fmt.Errorf("duration must be greater than 0")
	}

	return totalMonths, nil
}

// calculateMonthlyPayment calculates the monthly payment using the amortization formula
// M = P * [r(1+r)^n] / [(1+r)^n - 1]
func calculateMonthlyPayment(principal, monthlyRate float64, months int) float64 {
	if monthlyRate == 0 {
		return principal / float64(months)
	}

	factor := math.Pow(1+monthlyRate, float64(months))
	monthlyPayment := principal * (monthlyRate * factor) / (factor - 1)
	return monthlyPayment
}

// getPeriods returns the list of time periods to display in tables
func getPeriods(loanDuration int, include30Year bool) []struct {
	label  string
	months int
} {
	// Define base periods (always included)
	basePeriods := []struct {
		label  string
		months int
	}{
		{"  1y", 12},
		{"  2y", 24},
		{"  3y", 36},
		{"  4y", 48},
		{"  5y", 60},
		{"  6y", 72},
		{"  7y", 84},
		{"  8y", 96},
		{"  9y", 108},
		{" 10y", 120},
	}

	// Extended periods (only if include30Year is true)
	extendedPeriods := []struct {
		label  string
		months int
	}{
		{" 15y", 180},
		{" 20y", 240},
		{" 30y", 360},
	}

	// Build standard periods based on include30Year flag
	standardPeriods := basePeriods
	if include30Year {
		standardPeriods = append(standardPeriods, extendedPeriods...)
	}

	// Build the final list of periods, inserting loan term if needed
	periods := []struct {
		label  string
		months int
	}{}

	// Create loan term label with X prefix
	var loanTermLabel string
	if loanDuration%12 == 0 {
		years := loanDuration / 12
		loanTermLabel = fmt.Sprintf("X %dy", years)
	} else {
		years := loanDuration / 12
		months := loanDuration % 12
		loanTermLabel = fmt.Sprintf("X %dy%dm", years, months)
	}

	inserted := false
	for _, period := range standardPeriods {
		// Insert loan term before the first period that's longer
		if !inserted && loanDuration < period.months && loanDuration > 0 {
			periods = append(periods, struct {
				label  string
				months int
			}{loanTermLabel, loanDuration})
			inserted = true
		}

		// Skip if this period matches the loan duration
		if period.months == loanDuration {
			periods = append(periods, struct {
				label  string
				months int
			}{loanTermLabel, loanDuration})
			inserted = true
		} else {
			periods = append(periods, period)
		}
	}

	// If loan term is longer than all standard periods, add it at the end
	if !inserted && loanDuration > 0 {
		periods = append(periods, struct {
			label  string
			months int
		}{loanTermLabel, loanDuration})
	}

	return periods
}

// displayInputParameters displays all input parameters in grouped format
func displayInputParameters(md *MarketData) {
	re := lipgloss.NewRenderer(os.Stdout)
	titleStyle := re.NewStyle().Foreground(MonokaiPink).Bold(true)
	labelStyle := re.NewStyle().Foreground(MonokaiCyan)
	groupStyle := re.NewStyle().Foreground(MonokaiOrange).Bold(true)

	fmt.Println()
	fmt.Println(titleStyle.Render("INPUT PARAMETERS"))

	fmt.Println()
	fmt.Println(groupStyle.Render("ECONOMIC ASSUMPTIONS"))
	fmt.Printf("  %s: %.2f%%\n", labelStyle.Render("Inflation Rate"), config.inflationRate)

	fmt.Println()
	fmt.Println(groupStyle.Render("BUYING"))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Asset Purchase Price"), formatCurrency(config.purchasePrice))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Loan Amount"), formatCurrency(config.loanAmount))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Downpayment"), formatCurrency(config.downpayment))
	fmt.Printf("  %s: %.2f%%\n", labelStyle.Render("Loan Rate"), config.annualRate)

	// Format loan duration
	loanDurationStr := ""
	if config.totalMonths%12 == 0 {
		loanDurationStr = fmt.Sprintf("%dy", config.totalMonths/12)
	} else {
		loanDurationStr = fmt.Sprintf("%d months", config.totalMonths)
	}
	fmt.Printf("  %s: %s\n", labelStyle.Render("Loan Duration"), loanDurationStr)
	fmt.Printf("  %s: %s\n", labelStyle.Render("Annual Tax & Insurance"), formatCurrency(config.annualInsurance))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Other Annual Costs"), formatCurrency(config.annualTaxes))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Monthly Expenses"), formatCurrency(config.monthlyExpenses))

	// Format appreciation rates
	appreciationRateStr := ""
	if len(appreciationRates) == 1 {
		appreciationRateStr = fmt.Sprintf("%.2f%% (all years)", appreciationRates[0])
	} else {
		rateStrs := make([]string, len(appreciationRates))
		for i, rate := range appreciationRates {
			if i == len(appreciationRates)-1 {
				rateStrs[i] = fmt.Sprintf("%.2f%% (year %d+)", rate, i+1)
			} else {
				rateStrs[i] = fmt.Sprintf("%.2f%% (year %d)", rate, i+1)
			}
		}
		appreciationRateStr = strings.Join(rateStrs, ", ")
	}
	fmt.Printf("  %s: %s\n", labelStyle.Render("Appreciation Rate"), appreciationRateStr)
	fmt.Printf("  %s: %s\n", labelStyle.Render("Total Monthly Cost"), formatCurrency(config.totalMonthlyBuyingCost))

	fmt.Println()
	fmt.Println(groupStyle.Render("RENTING"))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Rental Deposit"), formatCurrency(config.rentDeposit))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Monthly Rent"), formatCurrency(config.monthlyRent))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Annual Rent Costs"), formatCurrency(config.annualRentCosts))
	fmt.Printf("  %s: %s\n", labelStyle.Render("Other Annual Costs"), formatCurrency(config.otherAnnualCosts))
	fmt.Printf("  %s: %.2f%%\n", labelStyle.Render("Investment Return Rate"), config.investmentReturnRate)

	// Display market averages under investment return rate
	if md != nil && len(md.VOO) > 0 {
		vooAvg, qqqAvg, vtiAvg, bndAvg, mix6040Avg := calculateMarketAverages(md)
		if vooAvg > 0 {
			fmt.Printf("    Market Averages (10y): VOO %.1f%%, QQQ %.1f%%, VTI %.1f%%, BND %.1f%%, 60/40 %.1f%%\n",
				vooAvg, qqqAvg, vtiAvg, bndAvg, mix6040Avg)
		}
	}

	fmt.Printf("  %s: %s\n", labelStyle.Render("Total Monthly Cost"), formatCurrency(config.totalMonthlyRentingCost))

	if config.includeSelling > 0 {
		fmt.Println()
		fmt.Println(groupStyle.Render("SELLING"))
		fmt.Printf("  %s: Yes\n", labelStyle.Render("Include Selling Analysis"))
		fmt.Printf("  %s: %.2f%%\n", labelStyle.Render("Agent Commission"), config.agentCommission)
		fmt.Printf("  %s: %s\n", labelStyle.Render("Staging/Selling Costs"), formatCurrency(config.stagingCosts))
		fmt.Printf("  %s: %s\n", labelStyle.Render("Tax-Free Gains Limit"), formatCurrency(config.taxFreeLimit))
		fmt.Printf("  %s: %.2f%%\n", labelStyle.Render("Capital Gains Tax Rate"), config.capitalGainsTax)
	} else {
		fmt.Println()
		fmt.Println(groupStyle.Render("SELLING"))
		fmt.Printf("  %s: No\n", labelStyle.Render("Include Selling Analysis"))
	}
}

// displayAmortizationTable displays loan amortization details
func displayAmortizationTable() {
	periods := getPeriods(config.totalMonths, config.include30Year > 0)

	// Build table rows (header + data)
	rows := [][]string{
		{"Period", "Principal Paid", "Interest Paid", "Loan Balance"},
	}

	// Build each data row
	for _, period := range periods {
		monthIndex := period.months - 1
		if monthIndex >= len(remainingLoanBalance) {
			monthIndex = len(remainingLoanBalance) - 1
		}

		principalPaid := cumulativePrincipalPaid[monthIndex]
		interestPaid := cumulativeInterestPaid[monthIndex]
		loanBalance := remainingLoanBalance[monthIndex]

		rows = append(rows, []string{
			"LOAN " + period.label,
			formatCurrency(principalPaid),
			formatCurrency(interestPaid),
			formatCurrency(loanBalance),
		})
	}

	notes := "Note: Monthly payment is fixed. Each payment covers interest on remaining balance, with the rest going to principal. Early payments are mostly interest."
	displayTable("LOAN AMORTIZATION DETAILS", rows, notes, false)
}

// displayExpenditureTable displays total expenditure for buying vs renting
// Uses global monthlyBuyingCosts and monthlyRentingCosts arrays
func displayExpenditureTable() {
	periods := getPeriods(config.totalMonths, config.include30Year > 0)

	// Build table rows (header + data)
	rows := [][]string{
		{"Period", "Buying Expend.", "Renting Expend.", "Difference"},
	}

	// Add data rows
	for _, period := range periods {
		// Calculate total buying expenditure (downpayment + all monthly costs)
		buyingExpenditure := config.downpayment
		for i := 0; i < period.months; i++ {
			buyingExpenditure += monthlyBuyingCosts[i]
		}

		// Calculate total renting expenditure (deposit + all monthly costs)
		rentingExpenditure := config.rentDeposit
		for i := 0; i < period.months; i++ {
			rentingExpenditure += monthlyRentingCosts[i]
		}

		difference := buyingExpenditure - rentingExpenditure

		rows = append(rows, []string{
			"EXP " + period.label,
			formatCurrency(buyingExpenditure),
			formatCurrency(rentingExpenditure),
			formatCurrency(difference),
		})
	}

	notes := fmt.Sprintf("Note: All recurring costs (insurance, taxes, rent, HOA, etc.) are inflated annually at %.1f%% rate.", config.inflationRate)
	displayTable("TOTAL EXPENDITURE COMPARISON", rows, notes, false)
}

// displayComparisonTable displays buy vs rent net worth projections side-by-side
// Uses global monthlyBuyingCosts and monthlyRentingCosts arrays
func displayComparisonTable() {
	periods := getPeriods(config.totalMonths, config.include30Year > 0)

	// Build table rows (header + data)
	rows := [][]string{
		{"Period", "Asset Value", "Buying NW", "Cum Savings", "Market Return", "Renting NW", "RENT - BUY"},
	}

	// Build each data row
	for _, period := range periods {
		assetValue, _, buyingNetWorth := calculateNetWorth(period.months)

		rentingNetWorth := calculateRentingNetWorth(period.months)

		// Calculate cumulative savings (without investment growth)
		cumulativeSavings := config.downpayment - config.rentDeposit
		for i := 0; i < period.months; i++ {
			cumulativeSavings += monthlyBuyingCosts[i] - monthlyRentingCosts[i]
		}

		// Calculate market return (investment growth portion only)
		recoverableDeposit := config.rentDeposit * 0.75
		marketReturn := rentingNetWorth - cumulativeSavings - recoverableDeposit

		difference := rentingNetWorth - buyingNetWorth

		rows = append(rows, []string{
			"NET " + period.label,
			formatCurrency(assetValue),
			formatCurrency(buyingNetWorth),
			formatCurrency(cumulativeSavings),
			formatCurrency(marketReturn),
			formatCurrency(rentingNetWorth),
			formatCurrency(difference),
		})
	}

	// Build note text with conditional buying NW explanation
	noteText := fmt.Sprintf("Note: 'Cum Savings' = Cumulative Savings track raw difference in costs (Buying - Renting) without investment growth. See Total Expenditure Comparison.\n\n'Market Return' = investment growth using monthly dollar-cost averaging at %.0f%% annual rate. Each month's savings are invested immediately and compounded monthly. This models realistic investing behavior (not lump sum at year start), so effective return < annual rate for short periods.\n\n'Renting NW' = Cumul. Savings + Market Return + 75%% recoverable deposit. ", config.investmentReturnRate)
	if config.includeSelling > 0 {
		noteText += "'Buying NW' = Net proceeds after selling (sale price - selling costs - loan payoff - taxes). "
	} else {
		noteText += "'Buying NW' = Asset value - remaining loan balance. "
	}
	noteText += "'RENT - BUY': Positive values mean renting wins, negative values mean buying wins."

	displayTable("NET WORTH PROJECTIONS: BUY VS RENT", rows, noteText, false)
}

// calculateSaleProceeds calculates the net proceeds from selling at a given time
func calculateSaleProceeds(months int) (salePrice, totalSellingCosts, loanPayoff, capitalGains, taxOnGains, netProceeds float64) {
	// Calculate asset value (sale price) by compounding appreciation rates
	salePrice = config.purchasePrice
	years := months / 12
	remainingMonths := months % 12

	// Apply each year's rate
	for year := 0; year < years; year++ {
		rateIndex := year
		if rateIndex >= len(appreciationRates) {
			rateIndex = len(appreciationRates) - 1
		}
		salePrice *= (1 + appreciationRates[rateIndex]/100)
	}

	// Apply partial year if there are remaining months
	if remainingMonths > 0 {
		rateIndex := years
		if rateIndex >= len(appreciationRates) {
			rateIndex = len(appreciationRates) - 1
		}
		partialYearFactor := math.Pow(1+appreciationRates[rateIndex]/100, float64(remainingMonths)/12.0)
		salePrice *= partialYearFactor
	}

	// Calculate agent commission
	agentFee := salePrice * (config.agentCommission / 100)

	// Combine agent commission and staging costs
	totalSellingCosts = agentFee + config.stagingCosts

	// Get remaining loan balance
	monthIndex := months - 1
	if monthIndex >= len(remainingLoanBalance) {
		monthIndex = len(remainingLoanBalance) - 1
	}
	loanPayoff = remainingLoanBalance[monthIndex]

	// Calculate capital gains
	capitalGains = salePrice - config.purchasePrice

	// Calculate taxable gains (after exemption)
	taxableGains := math.Max(0, capitalGains-config.taxFreeLimit)

	// Calculate tax on gains
	taxOnGains = taxableGains * (config.capitalGainsTax / 100)

	// Calculate net proceeds
	netProceeds = salePrice - totalSellingCosts - loanPayoff - taxOnGains

	return
}

// displaySaleProceeds displays the proceeds from selling the property at various periods
func displaySaleProceeds() {
	periods := getPeriods(config.totalMonths, config.include30Year > 0)

	// Build table rows (header + data)
	rows := [][]string{
		{"Period", "Sale Price", "Selling Cost", "Loan Payoff", "Cap Gains", "Tax", "Net Proceeds"},
	}

	// Build each data row
	for _, period := range periods {
		salePrice, totalSellingCosts, loanPayoff, capitalGains, taxOnGains, netProceeds := calculateSaleProceeds(period.months)

		rows = append(rows, []string{
			"SALE " + period.label,
			formatCurrency(salePrice),
			formatCurrency(totalSellingCosts),
			formatCurrency(loanPayoff),
			formatCurrency(capitalGains),
			formatCurrency(taxOnGains),
			formatCurrency(netProceeds),
		})
	}

	notes := "Note: Appreciation rates are applied year-by-year (compounded). If multiple rates are specified (e.g., '-20,-10,-5'), first rate applies to year 1, second to year 2, etc. The last rate applies to all remaining years. Sale price = compounded property value."
	displayTable("SALE PROCEEDS ANALYSIS", rows, notes, false)
}

// displayNetWorthTable displays net worth projections in a table format
// Uses global monthlyBuyingCosts array
func displayNetWorthTable(purchasePrice, downpayment float64, loanDuration int, includeSelling float64,
	agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax float64) {
	// Define standard periods
	standardPeriods := []struct {
		label  string
		months int
	}{
		{"1 year", 12},
		{"3 years", 36},
		{"5 years", 60},
		{"10 years", 120},
		{"20 years", 240},
		{"30 years", 360},
	}

	// Build the final list of periods, inserting loan term if needed
	periods := []struct {
		label  string
		months int
	}{}

	loanTermLabel := fmt.Sprintf("Loan term (%d years)", loanDuration/12)
	if loanDuration%12 != 0 {
		years := loanDuration / 12
		months := loanDuration % 12
		loanTermLabel = fmt.Sprintf("Loan term (%dy %dm)", years, months)
	}

	inserted := false
	for _, period := range standardPeriods {
		// Insert loan term before the first period that's longer
		if !inserted && loanDuration < period.months && loanDuration > 0 {
			periods = append(periods, struct {
				label  string
				months int
			}{loanTermLabel, loanDuration})
			inserted = true
		}

		// Skip if this period matches the loan duration
		if period.months == loanDuration {
			periods = append(periods, struct {
				label  string
				months int
			}{loanTermLabel, loanDuration})
			inserted = true
		} else {
			periods = append(periods, period)
		}
	}

	// If loan term is longer than all standard periods, add it at the end
	if !inserted && loanDuration > 0 {
		periods = append(periods, struct {
			label  string
			months int
		}{loanTermLabel, loanDuration})
	}

	// Print table header
	fmt.Printf("\n%-20s %-20s %-20s %-20s\n", "Period", "Asset Value", "Total Expenditure", "Net Worth")
	fmt.Println(strings.Repeat("-", 80))

	// Print each row
	for _, period := range periods {
		assetValue, totalExpenditure, netWorth := calculateNetWorth(period.months)

		fmt.Printf("%-20s %-20s %-20s %-20s\n",
			period.label,
			formatCurrency(assetValue),
			formatCurrency(totalExpenditure),
			formatCurrency(netWorth),
		)
	}
}

// calculateNetWorth calculates the asset value, total expenditure, and net worth for a given time period
// Uses the global monthlyBuyingCosts and remainingLoanBalance arrays
func calculateNetWorth(months int) (float64, float64, float64) {
	// Calculate asset value by compounding each year's appreciation rate
	assetValue := config.purchasePrice
	years := months / 12
	remainingMonths := months % 12

	// Apply each year's rate
	for year := 0; year < years; year++ {
		rateIndex := year
		if rateIndex >= len(appreciationRates) {
			rateIndex = len(appreciationRates) - 1 // Use last rate for all future years
		}
		assetValue *= (1 + appreciationRates[rateIndex]/100)
	}

	// Apply partial year if there are remaining months
	if remainingMonths > 0 {
		rateIndex := years
		if rateIndex >= len(appreciationRates) {
			rateIndex = len(appreciationRates) - 1
		}
		partialYearFactor := math.Pow(1+appreciationRates[rateIndex]/100, float64(remainingMonths)/12.0)
		assetValue *= partialYearFactor
	}

	// Calculate total expenditure by summing monthly costs from array
	totalExpenditure := config.downpayment
	for i := 0; i < months; i++ {
		totalExpenditure += monthlyBuyingCosts[i]
	}

	// Calculate net worth
	var netWorth float64
	if config.includeSelling > 0 {
		// If selling is enabled, use net proceeds after selling costs
		_, _, _, _, _, netProceeds := calculateSaleProceeds(months)
		netWorth = netProceeds
	} else {
		// Otherwise, just asset value minus loan balance
		monthIndex := months - 1
		if monthIndex >= len(remainingLoanBalance) {
			monthIndex = len(remainingLoanBalance) - 1
		}
		loanBalance := remainingLoanBalance[monthIndex]
		netWorth = assetValue - loanBalance
	}

	return assetValue, totalExpenditure, netWorth
}

// populateMonthlyCosts fills global arrays with monthly costs for buying and renting
func populateMonthlyCosts(maxMonths int, monthlyLoanPayment, monthlyRecurringExpenses float64, loanDuration int, monthlyRentingCost, loanAmount, monthlyRate, inflationRate float64) {
	monthlyBuyingCosts = make([]float64, maxMonths)
	monthlyRentingCosts = make([]float64, maxMonths)
	remainingLoanBalance = make([]float64, maxMonths)
	cumulativePrincipalPaid = make([]float64, maxMonths)
	cumulativeInterestPaid = make([]float64, maxMonths)

	// Calculate current rental cost with annual increases
	currentRentingCost := monthlyRentingCost

	// Track current recurring expenses (will increase with inflation)
	currentRecurringExpenses := monthlyRecurringExpenses

	// Track remaining loan balance
	currentBalance := loanAmount
	totalPrincipalPaid := 0.0
	totalInterestPaid := 0.0

	for i := 0; i < maxMonths; i++ {
		// Apply inflation to all costs at the start of each year (except the first month)
		if i > 0 && i%12 == 0 {
			currentRentingCost *= (1 + inflationRate/100)
			currentRecurringExpenses *= (1 + inflationRate/100)
		}

		// Set renting cost for this month
		monthlyRentingCosts[i] = currentRentingCost

		// Buying cost: loan payment stops after loan duration, but recurring expenses continue
		if i < loanDuration {
			monthlyBuyingCosts[i] = monthlyLoanPayment + currentRecurringExpenses

			// Calculate interest for this month
			interestPayment := currentBalance * monthlyRate
			// Principal payment is the remainder
			principalPayment := monthlyLoanPayment - interestPayment
			// Reduce the balance
			currentBalance -= principalPayment

			// Track cumulative amounts
			totalPrincipalPaid += principalPayment
			totalInterestPaid += interestPayment

			// Store remaining balance after this payment
			remainingLoanBalance[i] = currentBalance
			cumulativePrincipalPaid[i] = totalPrincipalPaid
			cumulativeInterestPaid[i] = totalInterestPaid
		} else {
			// After loan is paid off, only recurring expenses remain
			monthlyBuyingCosts[i] = currentRecurringExpenses
			remainingLoanBalance[i] = 0
			cumulativePrincipalPaid[i] = totalPrincipalPaid
			cumulativeInterestPaid[i] = totalInterestPaid
		}
	}
}

// calculateRentingNetWorth calculates net worth for the renting scenario
// Uses month-by-month calculation: investment grows from downpayment + monthly savings
func calculateRentingNetWorth(months int) float64 {
	// Start with downpayment minus deposit as initial investment
	investmentValue := config.downpayment - config.rentDeposit
	monthlyInvestmentRate := config.investmentReturnRate / 100 / 12

	// For each month: calculate savings, add to investment, grow investment
	for i := 0; i < months; i++ {
		// Monthly savings = buying cost - renting cost
		monthlySavings := monthlyBuyingCosts[i] - monthlyRentingCosts[i]

		// Add savings to investment
		investmentValue += monthlySavings

		// Apply monthly growth
		investmentValue *= (1 + monthlyInvestmentRate)
	}

	// Add back 75% of deposit (recoverable)
	recoverableDeposit := config.rentDeposit * 0.75

	return investmentValue + recoverableDeposit
}
