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
var debugMode bool

// Global arrays for monthly costs
var monthlyBuyingCosts []float64
var monthlyRentingCosts []float64
var remainingLoanBalance []float64
var cumulativePrincipalPaid []float64
var cumulativeInterestPaid []float64

const inputsFile = ".rentobuy_inputs.json"

func main() {
	// Parse command line flags
	flag.BoolVar(&useDefaults, "defaults", false, "Use all previously saved default values without prompting")
	flag.BoolVar(&debugMode, "debug", false, "Show detailed debugging information including principal paid off")
	flag.Parse()

	// Load previous inputs
	savedDefaults = loadInputs()
	currentInputs = make(map[string]string)

	// Check if we have defaults when --defaults flag is used
	if useDefaults && len(savedDefaults) == 0 {
		fmt.Println("Error: --defaults flag used but no saved defaults found. Run without the flag first.")
		return
	}

	// Get inflation rate first
	fmt.Println("\n--- Economic Assumptions ---")
	inflationRate, err := getFloatInputWithDefault("inflation_rate", "Enter annual inflation rate (e.g., 3 for 3%): ")
	if err != nil {
		fmt.Println("Invalid inflation rate")
		return
	}
	fmt.Println("Note: All recurring costs (insurance, taxes, HOA, etc.) will increase annually at this rate.")

	// Get purchase price
	purchasePrice, err := getFloatInputWithDefault("purchase_price", "\nEnter purchase price: $")
	if err != nil {
		fmt.Println("Invalid purchase price")
		return
	}

	// Get downpayment
	downpayment, err := getFloatInputWithDefault("downpayment", "Enter downpayment: $")
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
		annualRate, err = getFloatInputWithDefault("loan_rate", "Enter loan rate percentage (e.g., 6.5 for 6.5%): ")
		if err != nil {
			fmt.Println("Invalid loan rate")
			return
		}

		// Get loan duration
		totalMonths, err = getStringInputAndParseWithDefault("loan_duration", "Enter loan duration (e.g., 5y6m for 5 years 6 months): ", parseDuration)
		if err != nil {
			fmt.Println("Invalid duration format:", err)
			return
		}

		// Calculate monthly payment for buying
		monthlyRate = annualRate / 100 / 12
		monthlyLoanPayment = calculateMonthlyPayment(loanAmount, monthlyRate, totalMonths)
	}

	// Get annual expenses
	fmt.Println("\n--- Annual Fixed Expenses ---")
	annualInsurance, err := getFloatInputWithDefault("annual_insurance", "Enter annual insurance cost: $")
	if err != nil {
		fmt.Println("Invalid insurance amount")
		return
	}

	annualTaxes, err := getFloatInputWithDefault("annual_taxes", "Enter other annual costs (taxes, etc.): $")
	if err != nil {
		fmt.Println("Invalid taxes amount")
		return
	}

	totalAnnualExpenses := annualInsurance + annualTaxes

	// Get monthly expenses
	fmt.Println("\n--- Monthly Fixed Expenses ---")
	monthlyExpenses, err := getFloatInputWithDefault("monthly_expenses", "Enter monthly fixed expenses (HOA, utilities, etc.): $")
	if err != nil {
		fmt.Println("Invalid monthly expenses")
		return
	}

	// Get asset appreciation rate
	fmt.Println("\n--- Asset Appreciation ---")
	appreciationRate, err := getFloatInputWithDefault("appreciation_rate", "Enter annual appreciation rate (e.g., 3 for 3%, -2 for -2% depreciation): ")
	if err != nil {
		fmt.Println("Invalid appreciation rate")
		return
	}

	// Get renting parameters
	fmt.Println("\n--- Renting Parameters ---")
	rentDeposit, err := getFloatInputWithDefault("rent_deposit", "Enter rental deposit: $")
	if err != nil {
		fmt.Println("Invalid deposit amount")
		return
	}

	monthlyRent, err := getFloatInputWithDefault("monthly_rent", "Enter monthly rent: $")
	if err != nil {
		fmt.Println("Invalid monthly rent")
		return
	}

	annualRentCosts, err := getFloatInputWithDefault("annual_rent_costs", "Enter annual rent costs: $")
	if err != nil {
		fmt.Println("Invalid annual rent costs")
		return
	}

	otherAnnualCosts, err := getFloatInputWithDefault("other_annual_costs", "Enter other annual costs: $")
	if err != nil {
		fmt.Println("Invalid other annual costs")
		return
	}

	investmentReturnRate, err := getFloatInputWithDefault("investment_return_rate", "Enter investment return rate (e.g., 7 for 7%): ")
	if err != nil {
		fmt.Println("Invalid investment return rate")
		return
	}

	// Save inputs for next time
	saveInputs(currentInputs)

	// Calculate monthly recurring expenses
	monthlyRecurringExpenses := (totalAnnualExpenses / 12) + monthlyExpenses
	totalMonthlyBuyingCost := monthlyLoanPayment + monthlyRecurringExpenses

	// Calculate monthly cost for renting
	monthlyRentingExpenses := (annualRentCosts / 12) + (otherAnnualCosts / 12)
	totalMonthlyRentingCost := monthlyRent + monthlyRentingExpenses

	// Populate global cost arrays for projections (360 months = 30 years max)
	populateMonthlyCosts(360, monthlyLoanPayment, monthlyRecurringExpenses, totalMonths, totalMonthlyRentingCost, loanAmount, monthlyRate, inflationRate)

	// Display results
	fmt.Println("\n=== Buying Details ===")
	if loanAmount > 0 {
		fmt.Printf("Loan amount: %s\n", formatCurrency(loanAmount))
		fmt.Printf("Annual interest rate: %.2f%%\n", annualRate)
		fmt.Printf("Loan duration: %s months\n", formatNumber(totalMonths))
		fmt.Printf("\nMonthly loan payment: %s\n", formatCurrency(monthlyLoanPayment))
	} else {
		fmt.Printf("Loan amount: %s (no loan needed)\n", formatCurrency(loanAmount))
	}
	fmt.Printf("Monthly recurring expenses: %s\n", formatCurrency(monthlyRecurringExpenses))
	fmt.Printf("Total monthly buying cost: %s\n", formatCurrency(totalMonthlyBuyingCost))

	fmt.Println("\n=== Renting Details ===")
	fmt.Printf("Rental deposit: %s\n", formatCurrency(rentDeposit))
	fmt.Printf("Monthly rent: %s\n", formatCurrency(monthlyRent))
	fmt.Printf("Monthly rent expenses: %s\n", formatCurrency(monthlyRentingExpenses))
	fmt.Printf("Total monthly renting cost: %s\n", formatCurrency(totalMonthlyRentingCost))

	// Display projections
	fmt.Println("\n=== Total Expenditure Comparison ===")
	displayExpenditureTable(downpayment, totalMonths, rentDeposit)

	if debugMode && loanAmount > 0 {
		fmt.Println("\n=== Loan Amortization Details ===")
		displayAmortizationTable(loanAmount, totalMonths)
	}

	fmt.Println("\n=== Net Worth Projections: Buy vs Rent ===")
	displayComparisonTable(purchasePrice, downpayment, appreciationRate, totalMonths,
		rentDeposit, investmentReturnRate)
}

// getInputWithDefault prompts the user and reads string input with default value support
// Uses the key to retrieve the appropriate default value from savedDefaults
func getInputWithDefault(key, prompt string) string {
	defaultValue := savedDefaults[key]

	// If --defaults flag is set, automatically use the default
	if useDefaults {
		if defaultValue != "" {
			fmt.Printf("%s%s (using default)\n", prompt, defaultValue)
			currentInputs[key] = defaultValue
			return defaultValue
		}
		// No default available but flag is set - this shouldn't happen due to check in main
		fmt.Printf("%s(no default available)\n", prompt)
		currentInputs[key] = ""
		return ""
	}

	if defaultValue != "" {
		fmt.Printf("%s[%s] ", prompt, defaultValue)
	} else {
		fmt.Print(prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// If empty and we have a default, use the default
	if input == "" && defaultValue != "" {
		input = defaultValue
	}

	currentInputs[key] = input
	return input
}

// getFloatInputWithDefault prompts with a default and converts to float
func getFloatInputWithDefault(key, prompt string) (float64, error) {
	input := getInputWithDefault(key, prompt)
	value, err := parseAmount(input)
	return value, err
}

// getStringInputAndParseWithDefault prompts with a default and applies parser
func getStringInputAndParseWithDefault(key, prompt string, parser func(string) (int, error)) (int, error) {
	input := getInputWithDefault(key, prompt)
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

// getStringInputAndParse prompts the user and applies a parser function
func getStringInputAndParse(prompt string, parser func(string) (int, error)) (int, error) {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return parser(strings.TrimSpace(input))
}

// formatCurrency formats a number as currency with commas and 2 decimal places
func formatCurrency(amount float64) string {
	// Handle negative numbers
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	// Format with 2 decimal places
	formatted := fmt.Sprintf("%.2f", amount)
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
func getPeriods(loanDuration int) []struct {
	label  string
	months int
} {
	// Define standard periods with proper spacing for alignment
	standardPeriods := []struct {
		label  string
		months int
	}{
		{"  1 year", 12},
		{"  2 years", 24},
		{"  3 years", 36},
		{"  4 years", 48},
		{"  5 years", 60},
		{"  6 years", 72},
		{"  7 years", 84},
		{"  8 years", 96},
		{"  9 years", 108},
		{" 10 years", 120},
		{" 15 years", 180},
		{" 20 years", 240},
		{" 30 years", 360},
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
		loanTermLabel = fmt.Sprintf("X %d years", years)
	} else {
		years := loanDuration / 12
		months := loanDuration % 12
		loanTermLabel = fmt.Sprintf("X %dy %dm", years, months)
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

// displayAmortizationTable displays loan amortization details
func displayAmortizationTable(loanAmount float64, loanDuration int) {
	periods := getPeriods(loanDuration)

	// Print table header
	fmt.Printf("\n%-20s %-20s %-20s %-20s\n", "Period", "Principal Paid", "Interest Paid", "Loan Balance")
	fmt.Println(strings.Repeat("-", 80))

	// Print each row
	for _, period := range periods {
		monthIndex := period.months - 1
		if monthIndex >= len(remainingLoanBalance) {
			monthIndex = len(remainingLoanBalance) - 1
		}

		principalPaid := cumulativePrincipalPaid[monthIndex]
		interestPaid := cumulativeInterestPaid[monthIndex]
		loanBalance := remainingLoanBalance[monthIndex]

		fmt.Printf("%-20s %-20s %-20s %-20s\n",
			period.label,
			formatCurrency(principalPaid),
			formatCurrency(interestPaid),
			formatCurrency(loanBalance),
		)
	}
}

// displayExpenditureTable displays total expenditure for buying vs renting
// Uses global monthlyBuyingCosts and monthlyRentingCosts arrays
func displayExpenditureTable(downpayment float64, loanDuration int, rentDeposit float64) {
	periods := getPeriods(loanDuration)

	// Print table header
	fmt.Printf("\n%-20s %-25s %-25s %-25s\n", "Period", "Buying Expenditure", "Renting Expenditure", "Difference")
	fmt.Println(strings.Repeat("-", 95))

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

		fmt.Printf("%-20s %-25s %-25s %-25s\n",
			period.label,
			formatCurrency(buyingExpenditure),
			formatCurrency(rentingExpenditure),
			formatCurrency(difference),
		)
	}
}

// displayComparisonTable displays buy vs rent net worth projections side-by-side
// Uses global monthlyBuyingCosts and monthlyRentingCosts arrays
func displayComparisonTable(purchasePrice, downpayment, appreciationRate float64, loanDuration int,
	rentDeposit, investmentReturnRate float64) {
	periods := getPeriods(loanDuration)

	// Print table header
	fmt.Printf("\n%-15s %-18s %-18s %-18s %-18s %-18s\n", "Period", "Asset Value", "Buying Net Worth", "Cumulative Savings", "Renting Net Worth", "RENT - BUY")
	fmt.Println(strings.Repeat("-", 110))

	// Print each row
	for _, period := range periods {
		assetValue, _, buyingNetWorth := calculateNetWorth(
			period.months, purchasePrice, downpayment, appreciationRate,
		)

		rentingNetWorth := calculateRentingNetWorth(
			period.months, downpayment, rentDeposit, investmentReturnRate,
		)

		// Calculate cumulative savings (without investment growth)
		cumulativeSavings := downpayment - rentDeposit
		for i := 0; i < period.months; i++ {
			cumulativeSavings += monthlyBuyingCosts[i] - monthlyRentingCosts[i]
		}

		difference := rentingNetWorth - buyingNetWorth

		fmt.Printf("%-15s %-18s %-18s %-18s %-18s %-18s\n",
			period.label,
			formatCurrency(assetValue),
			formatCurrency(buyingNetWorth),
			formatCurrency(cumulativeSavings),
			formatCurrency(rentingNetWorth),
			formatCurrency(difference),
		)
	}
}

// displayNetWorthTable displays net worth projections in a table format
// Uses global monthlyBuyingCosts array
func displayNetWorthTable(purchasePrice, downpayment, appreciationRate float64, loanDuration int) {
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
			period.months, purchasePrice, downpayment, appreciationRate,
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
func calculateNetWorth(months int, purchasePrice, downpayment, appreciationRate float64) (float64, float64, float64) {
	// Calculate years for appreciation (asset continues to appreciate beyond loan term)
	years := float64(months) / 12.0

	// Calculate asset value with appreciation/depreciation
	assetValue := purchasePrice * math.Pow(1+(appreciationRate/100), years)

	// Calculate total expenditure by summing monthly costs from array
	totalExpenditure := downpayment
	for i := 0; i < months; i++ {
		totalExpenditure += monthlyBuyingCosts[i]
	}

	// Get remaining loan balance at this point
	monthIndex := months - 1
	if monthIndex >= len(remainingLoanBalance) {
		monthIndex = len(remainingLoanBalance) - 1
	}
	loanBalance := remainingLoanBalance[monthIndex]

	// Calculate net worth: asset value minus what you still owe
	netWorth := assetValue - loanBalance

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
