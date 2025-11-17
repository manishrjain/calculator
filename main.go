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

const inputsFile = ".rentobuy_inputs.json"

func main() {
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
	if err != nil {
		fmt.Println("Invalid purchase price")
		return
	}

	downpayment, err := getFloatValue("downpayment")
	if err != nil {
		fmt.Println("Invalid downpayment")
		return
	}

	// Calculate loan amount
	loanAmount := purchasePrice - downpayment

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

	// Populate global cost arrays for projections (360 months = 30 years max)
	populateMonthlyCosts(360, monthlyLoanPayment, monthlyRecurringExpenses, totalMonths, totalMonthlyRentingCost, loanAmount, monthlyRate, inflationRate)

	// Display input parameters
	displayInputParameters(inflationRate, purchasePrice, downpayment, loanAmount, annualRate, totalMonths,
		annualInsurance, annualTaxes, monthlyExpenses, totalMonthlyBuyingCost,
		rentDeposit, monthlyRent, annualRentCosts, otherAnnualCosts, investmentReturnRate, totalMonthlyRentingCost,
		includeSelling, agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax, marketData)

	// Display market data after input parameters
	displayMarketData(marketData)

	// Display projections
	fmt.Println("\n=== Total Expenditure Comparison ===")
	displayExpenditureTable(downpayment, totalMonths, rentDeposit, include30Year)
	fmt.Println("\nNote: All recurring costs (insurance, taxes, rent, HOA, etc.) are inflated annually at", inflationRate, "% rate.")

	if loanAmount > 0 {
		fmt.Println("\n=== Loan Amortization Details ===")
		displayAmortizationTable(loanAmount, totalMonths, include30Year)
		fmt.Println("\nNote: Monthly payment is fixed. Each payment covers interest on remaining balance,")
		fmt.Println("      with the rest going to principal. Early payments are mostly interest.")
	}

	if includeSelling > 0 {
		fmt.Println("\n=== Sale Proceeds Analysis ===")
		displaySaleProceeds(purchasePrice, downpayment, totalMonths,
			agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax, include30Year)
		fmt.Println("\nNote: Appreciation rates are applied year-by-year (compounded). If multiple rates are")
		fmt.Println("      specified (e.g., '-20,-10,-5'), first rate applies to year 1, second to year 2, etc.")
		fmt.Println("      The last rate applies to all remaining years. Sale price = compounded property value.")
	}

	fmt.Println("\n=== Net Worth Projections: Buy vs Rent ===")
	displayComparisonTable(purchasePrice, downpayment, totalMonths,
		rentDeposit, investmentReturnRate, include30Year, includeSelling,
		agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax)
	fmt.Println("\nNote: 'Cumul. Savings' = raw difference in costs (Buying - Renting) without investment growth.")
	fmt.Println("      See Total Expenditure Comparison.")
	fmt.Println("      'Market Return' = investment growth using monthly dollar-cost averaging at", investmentReturnRate, "% annual rate.")
	fmt.Println("      Each month's savings are invested immediately and compounded monthly. This models realistic")
	fmt.Println("      investing behavior (not lump sum at year start), so effective return < annual rate for short periods.")
	fmt.Println("      'Renting NW' = Cumul. Savings + Market Return + 75% recoverable deposit.")
	if includeSelling > 0 {
		fmt.Println("      'Buying NW' = Net proceeds after selling (sale price - selling costs - loan payoff - taxes).")
	} else {
		fmt.Println("      'Buying NW' = Asset value - remaining loan balance.")
	}
	fmt.Println("      'RENT - BUY': Positive values mean renting wins, negative values mean buying wins.")
	fmt.Println()
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
func displayInputParameters(inflationRate, purchasePrice, downpayment, loanAmount, annualRate float64, loanDuration int,
	annualInsurance, annualTaxes, monthlyExpenses, totalMonthlyBuyingCost,
	rentDeposit, monthlyRent, annualRentCosts, otherAnnualCosts, investmentReturnRate, totalMonthlyRentingCost,
	includeSelling, agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax float64, md *MarketData) {

	fmt.Println("\n=== INPUT PARAMETERS ===")

	fmt.Println("\nECONOMIC ASSUMPTIONS")
	fmt.Printf("  Inflation Rate: %.2f%%\n", inflationRate)

	fmt.Println("\nBUYING")
	fmt.Printf("  Purchase Price: %s\n", formatCurrency(purchasePrice))
	fmt.Printf("  Downpayment: %s\n", formatCurrency(downpayment))
	fmt.Printf("  Loan Amount: %s\n", formatCurrency(loanAmount))
	fmt.Printf("  Loan Rate: %.2f%%\n", annualRate)

	// Format loan duration
	loanDurationStr := ""
	if loanDuration%12 == 0 {
		loanDurationStr = fmt.Sprintf("%dy", loanDuration/12)
	} else {
		loanDurationStr = fmt.Sprintf("%d months", loanDuration)
	}
	fmt.Printf("  Loan Duration: %s\n", loanDurationStr)
	fmt.Printf("  Annual Tax & Insurance: %s\n", formatCurrency(annualInsurance))
	fmt.Printf("  Other Annual Costs: %s\n", formatCurrency(annualTaxes))
	fmt.Printf("  Monthly Expenses: %s\n", formatCurrency(monthlyExpenses))

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
	fmt.Printf("  Appreciation Rate: %s\n", appreciationRateStr)
	fmt.Printf("  Total Monthly Cost: %s\n", formatCurrency(totalMonthlyBuyingCost))

	fmt.Println("\nRENTING")
	fmt.Printf("  Rental Deposit: %s\n", formatCurrency(rentDeposit))
	fmt.Printf("  Monthly Rent: %s\n", formatCurrency(monthlyRent))
	fmt.Printf("  Annual Rent Costs: %s\n", formatCurrency(annualRentCosts))
	fmt.Printf("  Other Annual Costs: %s\n", formatCurrency(otherAnnualCosts))
	fmt.Printf("  Investment Return Rate: %.2f%%\n", investmentReturnRate)

	// Display market averages under investment return rate
	if md != nil && len(md.VOO) > 0 {
		vooAvg, qqqAvg, vtiAvg, bndAvg, mix6040Avg := calculateMarketAverages(md)
		if vooAvg > 0 {
			fmt.Printf("    Market Averages (10y): VOO %.1f%%, QQQ %.1f%%, VTI %.1f%%, BND %.1f%%, 60/40 %.1f%%\n",
				vooAvg, qqqAvg, vtiAvg, bndAvg, mix6040Avg)
		}
	}

	fmt.Printf("  Total Monthly Cost: %s\n", formatCurrency(totalMonthlyRentingCost))

	if includeSelling > 0 {
		fmt.Println("\nSELLING")
		fmt.Printf("  Include Selling Analysis: Yes\n")
		fmt.Printf("  Agent Commission: %.2f%%\n", agentCommission)
		fmt.Printf("  Staging/Selling Costs: %s\n", formatCurrency(stagingCosts))
		fmt.Printf("  Tax-Free Gains Limit: %s\n", formatCurrency(taxFreeLimit))
		fmt.Printf("  Capital Gains Tax Rate: %.2f%%\n", capitalGainsTax)
	} else {
		fmt.Println("\nSELLING")
		fmt.Printf("  Include Selling Analysis: No\n")
	}
}

// displayAmortizationTable displays loan amortization details
func displayAmortizationTable(loanAmount float64, loanDuration int, include30Year float64) {
	periods := getPeriods(loanDuration, include30Year > 0)

	// Print table header
	fmt.Printf("\n%-12s %-15s %-15s %-15s\n", "Period", "Principal Paid", "Interest Paid", "Loan Balance")
	fmt.Println(strings.Repeat("-", 57))

	// Print each row
	for _, period := range periods {
		monthIndex := period.months - 1
		if monthIndex >= len(remainingLoanBalance) {
			monthIndex = len(remainingLoanBalance) - 1
		}

		principalPaid := cumulativePrincipalPaid[monthIndex]
		interestPaid := cumulativeInterestPaid[monthIndex]
		loanBalance := remainingLoanBalance[monthIndex]

		fmt.Printf("%-12s %-15s %-15s %-15s\n",
			"LOAN "+period.label,
			formatCurrency(principalPaid),
			formatCurrency(interestPaid),
			formatCurrency(loanBalance),
		)
	}
}

// displayExpenditureTable displays total expenditure for buying vs renting
// Uses global monthlyBuyingCosts and monthlyRentingCosts arrays
func displayExpenditureTable(downpayment float64, loanDuration int, rentDeposit float64, include30Year float64) {
	periods := getPeriods(loanDuration, include30Year > 0)

	// Print table header
	fmt.Printf("\n%-12s %-18s %-18s %-15s\n", "Period", "Buying Expend.", "Renting Expend.", "Difference")
	fmt.Println(strings.Repeat("-", 63))

	// Print each row
	for _, period := range periods {
		// Calculate total buying expenditure (downpayment + all monthly costs)
		buyingExpenditure := downpayment
		for i := 0; i < period.months; i++ {
			buyingExpenditure += monthlyBuyingCosts[i]
		}

		// Calculate total renting expenditure (deposit + all monthly costs)
		rentingExpenditure := rentDeposit
		for i := 0; i < period.months; i++ {
			rentingExpenditure += monthlyRentingCosts[i]
		}

		difference := buyingExpenditure - rentingExpenditure

		fmt.Printf("%-12s %-18s %-18s %-15s\n",
			"EXP "+period.label,
			formatCurrency(buyingExpenditure),
			formatCurrency(rentingExpenditure),
			formatCurrency(difference),
		)
	}
}

// displayComparisonTable displays buy vs rent net worth projections side-by-side
// Uses global monthlyBuyingCosts and monthlyRentingCosts arrays
func displayComparisonTable(purchasePrice, downpayment float64, loanDuration int,
	rentDeposit, investmentReturnRate float64, include30Year float64, includeSelling float64,
	agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax float64) {
	periods := getPeriods(loanDuration, include30Year > 0)

	// Print table header
	fmt.Printf("\n%-12s %-13s %-12s %-13s %-13s %-12s %-12s\n", "Period", "Asset Value", "Buying NW", "Cumul. Savings", "Market Return", "Renting NW", "RENT - BUY")
	fmt.Println(strings.Repeat("-", 87))

	// Print each row
	for _, period := range periods {
		assetValue, _, buyingNetWorth := calculateNetWorth(
			period.months, purchasePrice, downpayment, includeSelling,
			agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax,
		)

		rentingNetWorth := calculateRentingNetWorth(
			period.months, downpayment, rentDeposit, investmentReturnRate,
		)

		// Calculate cumulative savings (without investment growth)
		cumulativeSavings := downpayment - rentDeposit
		for i := 0; i < period.months; i++ {
			cumulativeSavings += monthlyBuyingCosts[i] - monthlyRentingCosts[i]
		}

		// Calculate market return (investment growth portion only)
		recoverableDeposit := rentDeposit * 0.75
		marketReturn := rentingNetWorth - cumulativeSavings - recoverableDeposit

		difference := rentingNetWorth - buyingNetWorth

		fmt.Printf("%-12s %-13s %-12s %-13s %-13s %-12s %-12s\n",
			"NET "+period.label,
			formatCurrency(assetValue),
			formatCurrency(buyingNetWorth),
			formatCurrency(cumulativeSavings),
			formatCurrency(marketReturn),
			formatCurrency(rentingNetWorth),
			formatCurrency(difference),
		)
	}
}

// calculateSaleProceeds calculates the net proceeds from selling at a given time
func calculateSaleProceeds(months int, purchasePrice float64, agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax float64) (salePrice, totalSellingCosts, loanPayoff, capitalGains, taxOnGains, netProceeds float64) {
	// Calculate asset value (sale price) by compounding appreciation rates
	salePrice = purchasePrice
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
	agentFee := salePrice * (agentCommission / 100)

	// Combine agent commission and staging costs
	totalSellingCosts = agentFee + stagingCosts

	// Get remaining loan balance
	monthIndex := months - 1
	if monthIndex >= len(remainingLoanBalance) {
		monthIndex = len(remainingLoanBalance) - 1
	}
	loanPayoff = remainingLoanBalance[monthIndex]

	// Calculate capital gains
	capitalGains = salePrice - purchasePrice

	// Calculate taxable gains (after exemption)
	taxableGains := math.Max(0, capitalGains-taxFreeLimit)

	// Calculate tax on gains
	taxOnGains = taxableGains * (capitalGainsTax / 100)

	// Calculate net proceeds
	netProceeds = salePrice - totalSellingCosts - loanPayoff - taxOnGains

	return
}

// displaySaleProceeds displays the proceeds from selling the property at various periods
func displaySaleProceeds(purchasePrice, downpayment float64, loanDuration int,
	agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax float64, include30Year float64) {
	periods := getPeriods(loanDuration, include30Year > 0)

	// Print table header
	fmt.Printf("\n%-12s %-13s %-13s %-13s %-13s %-13s %-13s\n",
		"Period", "Sale Price", "Selling Cost", "Loan Payoff", "Capital Gain", "Tax on Gain", "Net Proceeds")
	fmt.Println(strings.Repeat("-", 90))

	// Print each row
	for _, period := range periods {
		salePrice, totalSellingCosts, loanPayoff, capitalGains, taxOnGains, netProceeds := calculateSaleProceeds(
			period.months, purchasePrice, agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax)

		fmt.Printf("%-12s %-13s %-13s %-13s %-13s %-13s %-13s\n",
			"SALE "+period.label,
			formatCurrency(salePrice),
			formatCurrency(totalSellingCosts),
			formatCurrency(loanPayoff),
			formatCurrency(capitalGains),
			formatCurrency(taxOnGains),
			formatCurrency(netProceeds),
		)
	}
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
		assetValue, totalExpenditure, netWorth := calculateNetWorth(
			period.months, purchasePrice, downpayment, includeSelling,
			agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax,
		)

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
func calculateNetWorth(months int, purchasePrice, downpayment float64, includeSelling float64,
	agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax float64) (float64, float64, float64) {
	// Calculate asset value by compounding each year's appreciation rate
	assetValue := purchasePrice
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
	totalExpenditure := downpayment
	for i := 0; i < months; i++ {
		totalExpenditure += monthlyBuyingCosts[i]
	}

	// Calculate net worth
	var netWorth float64
	if includeSelling > 0 {
		// If selling is enabled, use net proceeds after selling costs
		_, _, _, _, _, netProceeds := calculateSaleProceeds(
			months, purchasePrice, agentCommission, stagingCosts, taxFreeLimit, capitalGainsTax)
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
func calculateRentingNetWorth(months int, downpayment, rentDeposit, investmentReturnRate float64) float64 {
	// Start with downpayment minus deposit as initial investment
	investmentValue := downpayment - rentDeposit
	monthlyInvestmentRate := investmentReturnRate / 100 / 12

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
	recoverableDeposit := rentDeposit * 0.75

	return investmentValue + recoverableDeposit
}
