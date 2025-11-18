package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents a single input field in the form
type FormField struct {
	Key      string
	Label    string
	Help     string
	Input    textinput.Model
	Required bool
	IsToggle bool
	Toggled  bool
}

// FormModel is the bubbletea model for the interactive form
type FormModel struct {
	fieldsMap    map[string]*FormField  // All unique fields by key
	fields       []*FormField           // Flattened array for navigation (points to fieldsMap entries)
	groups       []FieldGroup
	currentField int
	submitted    bool
	values       map[string]string
	err          error
	marketData   *MarketData
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(MonokaiPink).Bold(true)
	blurredStyle = lipgloss.NewStyle().Foreground(MonokaiAdaptiveText)
	cursorStyle  = focusedStyle.Copy()
	helpStyle    = lipgloss.NewStyle().Foreground(MonokaiGrey)
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(MonokaiPink)
	groupStyle   = lipgloss.NewStyle().Bold(true).Foreground(MonokaiOrange)
)

// FieldGroup represents a group of related fields
type FieldGroup struct {
	Name     string
	Fields   []FormField
	Scenario string // "buy_vs_rent", "sell_vs_keep", or "both"
}

// NewFormModel creates a new form with all the input fields organized into groups
func NewFormModel(defaults map[string]string, md *MarketData) FormModel {
	// Determine which scenario is selected (default to BUY vs RENT)
	buyVsRentSelected := true
	sellVsKeepSelected := false
	if val, ok := defaults["scenario_sell_vs_keep"]; ok && (val == "1" || val == "yes" || val == "true") {
		buyVsRentSelected = false
		sellVsKeepSelected = true
	}

	// Create field groups
	groups := []FieldGroup{
		{
			Name:     "SCENARIO SELECTION",
			Scenario: "both",
			Fields: []FormField{
				makeToggleFieldWithValue("scenario_buy_vs_rent", "BUY vs RENT", "Select this scenario to compare buying vs renting", buyVsRentSelected),
				makeToggleFieldWithValue("scenario_sell_vs_keep", "SELL vs KEEP", "Select this scenario to compare selling vs keeping an existing asset", sellVsKeepSelected),
			},
		},
		{
			Name:     "ECONOMIC ASSUMPTIONS",
			Scenario: "both",
			Fields: []FormField{
				makeField("inflation_rate", "Inflation Rate (%)", "Annual inflation for all recurring costs", defaults),
				makeToggleField("include_30year", "Include 30-Year Projections", "Toggle to show 15y, 20y, 30y periods (default: 10y max)", defaults),
			},
		},
		{
			Name:     "BUYING",
			Scenario: "buy_vs_rent",
			Fields: []FormField{
				makeField("purchase_price", "Asset Purchase Price ($)", "Initial purchase price of the asset", defaults),
				makeField("loan_amount", "Loan Amount ($)", "Total mortgage/loan amount", defaults),
				makeField("loan_rate", "Loan Rate (%)", "Annual interest rate (e.g., 6.5)", defaults),
				makeField("loan_term", "Loan Term", "Loan duration (e.g., 5y, 30y)", defaults),
				makeField("annual_insurance", "Annual Tax & Insurance ($)", "Yearly insurance cost", defaults),
				makeField("annual_taxes", "Other Annual Costs ($)", "Taxes, HOA fees, etc.", defaults),
				makeField("monthly_expenses", "Monthly Expenses ($)", "Monthly HOA, utilities, etc.", defaults),
				makeField("appreciation_rate", "Appreciation Rate (%)", "Annual rate (can be negative for depreciation). Comma-separated values apply to first years, last value for all remaining years (e.g., '10,5,3' = 10% yr1, 5% yr2, 3% yr3+)", defaults),
			},
		},
		{
			Name:     "ASSET",
			Scenario: "sell_vs_keep",
			Fields: []FormField{
				makeField("purchase_price", "Asset Purchase Price ($)", "What you originally paid for the asset (for capital gains)", defaults),
				makeField("current_market_value", "Current Market Value ($)", "What the asset is worth today", defaults),
				makeField("remaining_loan_amount", "Remaining Loan Amount ($)", "Current outstanding loan balance", defaults),
				makeField("loan_rate", "Loan Rate (%)", "Annual interest rate on existing loan", defaults),
				makeField("loan_term", "Loan Term", "Original loan duration when started (e.g., 30y)", defaults),
				makeField("remaining_loan_term", "Remaining Loan Term", "Time left on loan (e.g., 25y)", defaults),
				makeField("annual_insurance", "Annual Tax & Insurance ($)", "Yearly costs if keeping", defaults),
				makeField("annual_taxes", "Other Annual Costs ($)", "Taxes, HOA fees, etc. if keeping", defaults),
				makeField("monthly_expenses", "Monthly Expenses ($)", "Monthly costs if keeping", defaults),
				makeField("appreciation_rate", "Appreciation Rate (%)", "Annual rate if keeping. Comma-separated for different years", defaults),
			},
		},
		{
			Name:     "RENTING",
			Scenario: "buy_vs_rent",
			Fields: []FormField{
				makeField("rent_deposit", "Rental Deposit ($)", "Initial rental deposit", defaults),
				makeField("monthly_rent", "Monthly Rent ($)", "Base monthly rent amount", defaults),
				makeField("annual_rent_costs", "Annual Rent Costs ($)", "Yearly rental-related costs", defaults),
				makeField("other_annual_costs", "Other Annual Costs ($)", "Additional yearly costs for renting", defaults),
				makeField("investment_return_rate", "Investment Return Rate (%)", "Expected return on investments. Market averages shown in output", defaults),
			},
		},
		{
			Name:     "INVESTING",
			Scenario: "sell_vs_keep",
			Fields: []FormField{
				makeToggleField("include_renting_sell", "Include Renting Analysis", "Toggle if selling means you'll need to rent", defaults),
				makeField("rent_deposit", "Rental Deposit ($)", "Initial rental deposit if selling", defaults),
				makeField("monthly_rent", "Monthly Rent ($)", "Monthly rent if selling", defaults),
				makeField("annual_rent_costs", "Annual Rent Costs ($)", "Yearly rental costs if selling", defaults),
				makeField("investment_return_rate", "Investment Return Rate (%)", "Expected return on sale proceeds. Market averages shown in output", defaults),
			},
		},
		{
			Name:     "SELLING",
			Scenario: "both",
			Fields: []FormField{
				makeToggleField("include_selling", "Include Selling Analysis", "Toggle to enable/disable selling analysis (BUY vs RENT only)", defaults),
				makeField("agent_commission", "Agent Commission (%)", "Percentage of sale price paid to agents", defaults),
				makeField("staging_costs", "Staging/Selling Costs ($)", "Fixed costs to prepare and sell", defaults),
				makeField("tax_free_limit", "Tax-Free Gains Limit ($)", "Capital gains exempt from tax (250k/500k)", defaults),
				makeField("capital_gains_tax", "Capital Gains Tax Rate (%)", "Long-term capital gains tax rate", defaults),
			},
		},
	}

	// Create fieldsMap to store unique fields by key (shared across scenarios)
	fieldsMap := make(map[string]*FormField)

	// Flatten fields for navigation, ensuring shared keys point to same instance
	var fields []*FormField
	for _, group := range groups {
		for i := range group.Fields {
			field := &group.Fields[i]
			// If this key already exists, use the existing field
			if existingField, exists := fieldsMap[field.Key]; exists {
				fields = append(fields, existingField)
			} else {
				// New field, add to map and fields array
				fieldsMap[field.Key] = field
				fields = append(fields, field)
			}
		}
	}

	// Focus the first field
	if len(fields) > 0 {
		fields[0].Input.Focus()
	}

	return FormModel{
		fieldsMap:    fieldsMap,
		fields:       fields,
		groups:       groups,
		currentField: 0,
		submitted:    false,
		values:       make(map[string]string),
		marketData:   md,
	}
}

func makeField(key, label, help string, defaults map[string]string) FormField {
	ti := textinput.New()
	ti.Placeholder = "0"
	ti.CharLimit = 32
	ti.Width = 30  // Fixed width to prevent jumping
	ti.Prompt = ""  // Disable built-in prompt, we'll use our own caret in the label
	ti.TextStyle = lipgloss.NewStyle().Foreground(MonokaiAdaptiveText)
	ti.Cursor.Style = focusedStyle

	if val, ok := defaults[key]; ok {
		ti.SetValue(val)
	}

	return FormField{
		Key:      key,
		Label:    label,
		Help:     help,
		Input:    ti,
		Required: true,
		IsToggle: false,
	}
}

func makeToggleField(key, label, help string, defaults map[string]string) FormField {
	ti := textinput.New()
	ti.Width = 30

	toggled := false
	if val, ok := defaults[key]; ok {
		toggled = val == "1" || val == "yes" || val == "true"
	}

	return FormField{
		Key:      key,
		Label:    label,
		Help:     help,
		Input:    ti,
		Required: false,
		IsToggle: true,
		Toggled:  toggled,
	}
}

func makeToggleFieldWithValue(key, label, help string, toggled bool) FormField {
	ti := textinput.New()
	ti.Width = 30

	return FormField{
		Key:      key,
		Label:    label,
		Help:     help,
		Input:    ti,
		Required: false,
		IsToggle: true,
		Toggled:  toggled,
	}
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

// isFieldVisible checks if a field at the given index is visible in the current scenario
func (m FormModel) isFieldVisible(fieldIndex int) bool {
	if fieldIndex < 0 || fieldIndex >= len(m.fields) {
		return false
	}

	// Determine current scenario
	selectedScenario := "buy_vs_rent" // default
	if scenarioField, ok := m.fieldsMap["scenario_sell_vs_keep"]; ok && scenarioField.Toggled {
		selectedScenario = "sell_vs_keep"
	}

	// Find which group this field belongs to
	currentIndex := 0
	for _, group := range m.groups {
		for range group.Fields {
			if currentIndex == fieldIndex {
				// Check if this group is visible in current scenario
				return group.Scenario == "both" || group.Scenario == selectedScenario
			}
			currentIndex++
		}
	}

	return false
}

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+t":
			// Toggle between scenarios
			if buyField, ok := m.fieldsMap["scenario_buy_vs_rent"]; ok {
				if sellField, ok := m.fieldsMap["scenario_sell_vs_keep"]; ok {
					// Toggle between the two
					buyField.Toggled = !buyField.Toggled
					sellField.Toggled = !sellField.Toggled
				}
			}
			return m, nil

		case "ctrl+k":
			// Save values and submit
			for _, field := range m.fields {
				if field.IsToggle {
					if field.Toggled {
						m.values[field.Key] = "1"
					} else {
						m.values[field.Key] = "0"
					}
				} else {
					m.values[field.Key] = field.Input.Value()
				}
			}
			m.submitted = true
			return m, tea.Quit

		case "up", "shift+tab":
			// Move to previous visible field
			m.fields[m.currentField].Input.Blur()
			for {
				m.currentField--
				if m.currentField < 0 {
					m.currentField = 0
					break
				}
				if m.isFieldVisible(m.currentField) {
					break
				}
			}
			m.fields[m.currentField].Input.Focus()

		case "down", "tab":
			// Move to next visible field
			m.fields[m.currentField].Input.Blur()
			for {
				m.currentField++
				if m.currentField >= len(m.fields) {
					m.currentField = len(m.fields) - 1
					break
				}
				if m.isFieldVisible(m.currentField) {
					break
				}
			}
			m.fields[m.currentField].Input.Focus()

		case " ", "enter":
			// Toggle if current field is a toggle
			if m.fields[m.currentField].IsToggle {
				currentKey := m.fields[m.currentField].Key

				// Handle mutual exclusivity for scenario toggles
				if currentKey == "scenario_buy_vs_rent" || currentKey == "scenario_sell_vs_keep" {
					// Find both scenario fields and ensure mutual exclusivity
					for i := range m.fields {
						if m.fields[i].Key == "scenario_buy_vs_rent" {
							m.fields[i].Toggled = (currentKey == "scenario_buy_vs_rent")
						} else if m.fields[i].Key == "scenario_sell_vs_keep" {
							m.fields[i].Toggled = (currentKey == "scenario_sell_vs_keep")
						}
					}
				} else {
					// Regular toggle
					m.fields[m.currentField].Toggled = !m.fields[m.currentField].Toggled
				}
				return m, nil
			}
		}
	}

	// Update the focused input field (but not if it's a toggle)
	var cmd tea.Cmd
	if !m.fields[m.currentField].IsToggle {
		m.fields[m.currentField].Input, cmd = m.fields[m.currentField].Input.Update(msg)
	}
	return m, cmd
}

func (m FormModel) View() string {
	if m.submitted {
		return ""
	}

	var b strings.Builder

	// Determine which scenario is currently selected
	selectedScenario := "buy_vs_rent" // default
	for _, field := range m.fields {
		if field.Key == "scenario_sell_vs_keep" && field.Toggled {
			selectedScenario = "sell_vs_keep"
			break
		}
	}

	// Title
	b.WriteString(titleStyle.Render("┌────────────────────────────────────────────────────────────────┐"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("│                   Rent vs Buy Calculator                       │"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("└────────────────────────────────────────────────────────────────┘"))
	b.WriteString("\n\n")

	// Track field index as we render groups
	fieldIndex := 0

	// Render each group
	for groupIdx, group := range m.groups {
		// Skip groups that don't match the selected scenario
		if group.Scenario != "both" && group.Scenario != selectedScenario {
			fieldIndex += len(group.Fields)
			continue
		}
		// Group header
		b.WriteString(groupStyle.Render("  " + group.Name))
		b.WriteString("\n")

		// Render fields in this group (label and input on same line)
		for i := 0; i < len(group.Fields); i++ {
			currentFieldIndex := fieldIndex + i
			// Get field from fieldsMap to ensure we're using shared instances
			groupField := &group.Fields[i]
			field := m.fieldsMap[groupField.Key]

			// Render input or toggle
			var input string
			if field.IsToggle {
				checkbox := "[ ]"
				if field.Toggled {
					checkbox = "[X]"
				}
				input = checkbox
			} else {
				input = field.Input.View()
			}

			// Print label and input on same line with matching colors
			if currentFieldIndex == m.currentField {
				// Focused: entire line is pink with caret
				labelText := fmt.Sprintf("%-50s", "❯ "+field.Label)
				b.WriteString(focusedStyle.Render(labelText))
				if field.IsToggle {
					b.WriteString(focusedStyle.Render(input))
				} else {
					b.WriteString(focusedStyle.Render("> "))
					b.WriteString(focusedStyle.Render(input))
				}
			} else {
				// Not focused: no caret on label, but caret before input value
				labelText := fmt.Sprintf("%-50s", "  "+field.Label)
				b.WriteString(blurredStyle.Render(labelText))
				if field.IsToggle {
					b.WriteString(blurredStyle.Render(input))
				} else {
					b.WriteString(blurredStyle.Render("> "))
					b.WriteString(input)
				}
			}
			b.WriteString("\n")

			// Show market averages after investment return rate field
			if field.Key == "investment_return_rate" && m.marketData != nil && len(m.marketData.VOO) > 0 {
				vooAvg, qqqAvg, vtiAvg, bndAvg, mix6040Avg := calculateMarketAverages(m.marketData)
				if vooAvg > 0 {
					marketInfo := fmt.Sprintf("    Market Averages (10y): VOO %.1f%%, QQQ %.1f%%, VTI %.1f%%, BND %.1f%%, 60/40 %.1f%%",
						vooAvg, qqqAvg, vtiAvg, bndAvg, mix6040Avg)
					b.WriteString(helpStyle.Render(marketInfo))
					b.WriteString("\n")
				}
			}
		}

		// Add spacing between groups (except after last group)
		if groupIdx < len(m.groups)-1 {
			b.WriteString("\n")
		}

		// Update field index for next group
		fieldIndex += len(group.Fields)
	}

	// Show help text for current field at the bottom
	currentField := m.fields[m.currentField]
	b.WriteString("\n")
	// Wrap help text at 80 characters with left padding for indentation
	helpTextStyle := helpStyle.Copy().Width(80).PaddingLeft(2)
	b.WriteString(helpTextStyle.Render(currentField.Help))
	b.WriteString("\n\n")

	// Navigation help
	b.WriteString(helpStyle.Render("  ↑/↓: Navigate  Space/Enter: Toggle  Ctrl+T: Switch Scenario  Ctrl+K: Calculate  Ctrl+C/Esc: Quit"))
	b.WriteString("\n")

	return b.String()
}

// RunInteractiveForm runs the interactive form and returns the values
func RunInteractiveForm(defaults map[string]string, md *MarketData) (map[string]string, error) {
	m := NewFormModel(defaults, md)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	model := finalModel.(FormModel)
	if !model.submitted {
		return nil, fmt.Errorf("form cancelled")
	}

	return model.values, nil
}
