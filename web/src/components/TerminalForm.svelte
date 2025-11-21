<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import type { CalculatorInputs, ScenarioType } from '../types';
  import { parseAmount, parseDuration, parseAppreciationRates } from '../lib/formatter';

  export let formInputs: any;

  const dispatch = createEventDispatcher();

  interface FormField {
    key: string;
    label: string;
    help: string;
    placeholder: string;
    visible: () => boolean;
    disabled?: () => boolean;
    isHeader?: boolean;
    headerText?: string;
    toggleValues?: string[];  // For fields that can toggle between two values
  }

  let fields: FormField[] = [];
  let currentFieldIndex = 0;
  let inputRefs: HTMLInputElement[] = [];

  $: {
    // Rebuild fields list when scenario changes
    fields = [
      { key: 'header_scenario', label: '', help: '', placeholder: '', visible: () => true, isHeader: true, headerText: 'SCENARIO SELECTION' },
      { key: 'scenario', label: 'Scenario', help: 'Select buy_vs_rent to compare buying vs renting, or sell_vs_keep to compare selling vs keeping an existing asset', placeholder: 'buy_vs_rent', visible: () => true, toggleValues: ['buy_vs_rent', 'sell_vs_keep'] },

      { key: 'header_economic', label: '', help: '', placeholder: '', visible: () => true, isHeader: true, headerText: 'ECONOMIC ASSUMPTIONS' },
      { key: 'inflationRate', label: 'Inflation Rate (%)', help: 'Annual inflation for all recurring costs', placeholder: '3', visible: () => true },
      { key: 'investmentReturnRate', label: 'Investment Return (%)', help: 'Expected return on investments. Market averages shown below', placeholder: '10', visible: () => true },
      { key: 'include30Year', label: '30-Year Projections', help: 'Toggle to show 15y, 20y, 30y periods (default: 10y max)', placeholder: 'no', visible: () => true, toggleValues: ['yes', 'no'] },

      { key: 'header_asset', label: '', help: '', placeholder: '', visible: () => true, isHeader: true, headerText: formInputs.scenario === 'sell_vs_keep' ? 'ASSET' : 'BUYING' },
      { key: 'purchasePrice', label: formInputs.scenario === 'sell_vs_keep' ? 'Original Purchase Price' : 'Asset Purchase Price', help: formInputs.scenario === 'sell_vs_keep' ? 'What you originally paid for the asset (for capital gains)' : 'Initial purchase price of the asset', placeholder: '500K', visible: () => true },
      { key: 'currentMarketValue', label: 'Current Market Value', help: 'What the asset is worth today', placeholder: '2.2M', visible: () => formInputs.scenario === 'sell_vs_keep' },
      { key: 'loanAmount', label: formInputs.scenario === 'sell_vs_keep' ? 'Original Loan Amount' : 'Loan Amount', help: formInputs.scenario === 'sell_vs_keep' ? 'The original loan amount when purchased (we\'ll calculate remaining balance)' : 'Total mortgage/loan amount', placeholder: '400K', visible: () => true },
      { key: 'loanRate', label: 'Loan Rate (%)', help: formInputs.scenario === 'sell_vs_keep' ? 'Annual interest rate on existing loan' : 'Annual interest rate (e.g., 6.5)', placeholder: '6.5', visible: () => true },
      { key: 'loanTerm', label: 'Loan Term', help: formInputs.scenario === 'sell_vs_keep' ? 'Original loan duration when started (e.g., 30y)' : 'Loan duration (e.g., 5y, 30y)', placeholder: '30y', visible: () => true },
      { key: 'remainingLoanTerm', label: 'Remaining Loan Term', help: 'Time left on loan (e.g., 25y)', placeholder: '25y', visible: () => formInputs.scenario === 'sell_vs_keep' },
      { key: 'annualInsurance', label: 'Annual Tax & Insurance', help: formInputs.scenario === 'sell_vs_keep' ? 'Yearly costs if keeping' : 'Yearly insurance cost', placeholder: '3K', visible: () => true },
      { key: 'annualTaxes', label: 'Other Annual Costs', help: formInputs.scenario === 'sell_vs_keep' ? 'Taxes, HOA fees, etc. if keeping' : 'Maintenance costs, etc.', placeholder: '5K', visible: () => true },
      { key: 'monthlyExpenses', label: 'Monthly Expenses', help: formInputs.scenario === 'sell_vs_keep' ? 'Monthly costs if keeping' : 'Monthly expenses. Typically include utilities, HOA, etc. Can be negative if earning income, e.g., -4K.', placeholder: '500', visible: () => true },
      { key: 'appreciationRate', label: 'Appreciation Rate (%)', help: formInputs.scenario === 'sell_vs_keep' ? 'Annual rate if keeping. Comma-separated for different years' : 'Annual rate (can be negative for depreciation). Comma-separated values apply to first years, last value for all remaining years (e.g., \'10,5,3\' = 10% yr1, 5% yr2, 3% yr3+)', placeholder: '3', visible: () => true },

      { key: 'header_renting', label: '', help: '', placeholder: '', visible: () => true, isHeader: true, headerText: formInputs.scenario === 'sell_vs_keep' ? 'INVESTING' : 'RENTING' },
      { key: 'includeRentingSell', label: 'Include Renting Analysis', help: 'Toggle if selling means you\'ll need to rent', placeholder: 'no', visible: () => formInputs.scenario === 'sell_vs_keep', toggleValues: ['yes', 'no'] },
      { key: 'rentDeposit', label: 'Rental Deposit', help: formInputs.scenario === 'sell_vs_keep' ? 'Initial rental deposit if selling' : 'Initial rental deposit', placeholder: '5K', visible: () => true, disabled: () => formInputs.scenario === 'sell_vs_keep' && formInputs.includeRentingSell !== 'yes' },
      { key: 'monthlyRent', label: 'Monthly Rent', help: formInputs.scenario === 'sell_vs_keep' ? 'Monthly rent if selling' : 'Base monthly rent amount', placeholder: '3K', visible: () => true, disabled: () => formInputs.scenario === 'sell_vs_keep' && formInputs.includeRentingSell !== 'yes' },
      { key: 'annualRentCosts', label: 'Annual Rent Costs', help: formInputs.scenario === 'sell_vs_keep' ? 'Yearly rental costs if selling' : 'Yearly rental-related costs', placeholder: '1K', visible: () => true, disabled: () => formInputs.scenario === 'sell_vs_keep' && formInputs.includeRentingSell !== 'yes' },
      { key: 'otherAnnualCosts', label: 'Other Annual Costs', help: 'Additional yearly costs for renting', placeholder: '500', visible: () => formInputs.scenario === 'buy_vs_rent' },

      { key: 'header_selling', label: '', help: '', placeholder: '', visible: () => true, isHeader: true, headerText: 'SELLING' },
      { key: 'includeSelling', label: 'Include Selling Analysis', help: 'Toggle to enable/disable selling analysis', placeholder: 'no', visible: () => true, toggleValues: ['yes', 'no'] },
      { key: 'agentCommission', label: 'Agent Commission (%)', help: 'Percentage of sale price paid to agents', placeholder: '6', visible: () => true, disabled: () => formInputs.includeSelling !== 'yes' },
      { key: 'stagingCosts', label: 'Staging/Selling Costs', help: 'Fixed costs to prepare and sell', placeholder: '10K', visible: () => true, disabled: () => formInputs.includeSelling !== 'yes' },
      { key: 'taxFreeLimits', label: 'Tax-Free Gains Limit', help: 'Capital gains exempt from tax. Comma-separated for different years (e.g., \'500K,0K\' = 500K year 1, 0 year 2+)', placeholder: '250K', visible: () => true, disabled: () => formInputs.includeSelling !== 'yes' },
      { key: 'capitalGainsTax', label: 'Capital Gains Tax (%)', help: 'Long-term capital gains tax rate', placeholder: '20', visible: () => true, disabled: () => formInputs.includeSelling !== 'yes' },
    ];
  }

  $: visibleFields = fields.filter(f => f.visible());
  $: currentHelpText = visibleFields[currentFieldIndex]?.help || '';

  function handleKeyDown(event: KeyboardEvent) {
    const currentField = visibleFields[currentFieldIndex];

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      moveToNextField();
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      moveToPreviousField();
    } else if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
      // Toggle value for fields with toggleValues
      if (currentField && currentField.toggleValues && currentField.toggleValues.length === 2) {
        event.preventDefault();
        toggleFieldValue(currentField);
      }
    } else if (event.key === 'Enter') {
      event.preventDefault();
      if (event.ctrlKey) {
        handleSubmit();
      } else {
        // Regular Enter just moves to next field
        moveToNextField();
      }
    } else if (event.key === 'Tab') {
      event.preventDefault();
      if (event.shiftKey) {
        moveToPreviousField();
      } else {
        moveToNextField();
      }
    }
  }

  function toggleFieldValue(field: FormField) {
    if (!field.toggleValues || field.toggleValues.length !== 2) return;

    const currentValue = formInputs[field.key];
    const [value1, value2] = field.toggleValues;

    // Toggle to the other value
    formInputs[field.key] = currentValue === value1 ? value2 : value1;
  }

  function moveToNextField() {
    let nextIndex = currentFieldIndex + 1;
    // Skip headers and disabled fields
    while (nextIndex < visibleFields.length && (visibleFields[nextIndex].isHeader || (visibleFields[nextIndex].disabled && visibleFields[nextIndex].disabled()))) {
      nextIndex++;
    }
    if (nextIndex < visibleFields.length) {
      currentFieldIndex = nextIndex;
      focusCurrentField();
    }
  }

  function moveToPreviousField() {
    let prevIndex = currentFieldIndex - 1;
    // Skip headers and disabled fields
    while (prevIndex >= 0 && (visibleFields[prevIndex].isHeader || (visibleFields[prevIndex].disabled && visibleFields[prevIndex].disabled()))) {
      prevIndex--;
    }
    if (prevIndex >= 0) {
      currentFieldIndex = prevIndex;
      focusCurrentField();
    }
  }

  function focusCurrentField() {
    setTimeout(() => {
      const input = inputRefs[currentFieldIndex];
      if (input) {
        input.focus({ preventScroll: true });
        input.select();
        // Scroll the field into view within the terminal-content container
        input.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }
    }, 0);
  }

  function handleFieldFocus(index: number) {
    currentFieldIndex = index;
  }

  function convertFormInputsToCalculatorInputs(): CalculatorInputs {
    return {
      scenario: formInputs.scenario as ScenarioType,
      inflationRate: parseFloat(formInputs.inflationRate) || 0,
      investmentReturnRate: parseFloat(formInputs.investmentReturnRate) || 0,
      include30Year: formInputs.include30Year === 'yes' || formInputs.include30Year === true,
      purchasePrice: parseAmount(formInputs.purchasePrice.toString()),
      currentMarketValue: formInputs.currentMarketValue ? parseAmount(formInputs.currentMarketValue.toString()) : undefined,
      loanAmount: parseAmount(formInputs.loanAmount.toString()),
      loanRate: parseFloat(formInputs.loanRate) || 0,
      loanTerm: parseDuration(formInputs.loanTerm),
      remainingLoanTerm: formInputs.remainingLoanTerm ? parseDuration(formInputs.remainingLoanTerm) : undefined,
      annualInsurance: parseAmount(formInputs.annualInsurance.toString()),
      annualTaxes: parseAmount(formInputs.annualTaxes.toString()),
      monthlyExpenses: parseAmount(formInputs.monthlyExpenses.toString()),
      appreciationRate: parseAppreciationRates(formInputs.appreciationRate),
      rentDeposit: parseAmount(formInputs.rentDeposit.toString()),
      monthlyRent: parseAmount(formInputs.monthlyRent.toString()),
      annualRentCosts: parseAmount(formInputs.annualRentCosts.toString()),
      otherAnnualCosts: parseAmount(formInputs.otherAnnualCosts.toString()),
      includeSelling: formInputs.includeSelling === 'yes' || formInputs.includeSelling === true,
      includeRentingSell: formInputs.includeRentingSell === 'yes' || formInputs.includeRentingSell === true,
      agentCommission: parseFloat(formInputs.agentCommission) || 0,
      stagingCosts: parseAmount(formInputs.stagingCosts.toString()),
      taxFreeLimits: parseAppreciationRates(formInputs.taxFreeLimits),
      capitalGainsTax: parseFloat(formInputs.capitalGainsTax) || 0,
    };
  }

  function handleSubmit() {
    try {
      const inputs = convertFormInputsToCalculatorInputs();
      dispatch('calculate', inputs);
    } catch (error) {
      alert(`Invalid input: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  function handleContainerClick() {
    // Refocus current field when clicking anywhere in the container
    focusCurrentField();
  }

  function handleGlobalKeyDown(event: KeyboardEvent) {
    // Capture arrow keys globally to prevent page scrolling
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp' || event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
      // If no input is focused, focus the current field
      const activeElement = document.activeElement;
      const isInputFocused = inputRefs.some(ref => ref === activeElement);
      if (!isInputFocused) {
        event.preventDefault();
        focusCurrentField();
      }
    }
  }

  onMount(() => {
    // Find first non-header, non-disabled field to focus
    let firstFieldIndex = 0;
    while (firstFieldIndex < visibleFields.length &&
           (visibleFields[firstFieldIndex].isHeader ||
            (visibleFields[firstFieldIndex].disabled && visibleFields[firstFieldIndex].disabled()))) {
      firstFieldIndex++;
    }
    currentFieldIndex = firstFieldIndex;
    focusCurrentField();

    // Add global keyboard handler
    window.addEventListener('keydown', handleGlobalKeyDown);

    return () => {
      window.removeEventListener('keydown', handleGlobalKeyDown);
    };
  });
</script>

<div class="terminal-container font-mono" on:click={handleContainerClick}>
  <div class="terminal-content">
    <form on:submit|preventDefault={handleSubmit} class="space-y-1">
    {#each visibleFields as field, index}
      {#if field.isHeader}
        <div class="section-header">
          <span class="text-monokai-orange">{field.headerText}</span>
        </div>
      {:else}
        <div class="terminal-field" class:focused={index === currentFieldIndex} class:disabled={field.disabled && field.disabled()}>
          <div class="flex items-center gap-2">
            <div class="field-label w-72 flex-shrink-0">
              <span class="text-monokai-pink">{index === currentFieldIndex ? '>' : ' '}</span>
              <span class="ml-2" class:text-monokai-pink={index === currentFieldIndex} class:text-monokai-text={index !== currentFieldIndex}>{field.label}:</span>
            </div>
            <div class="field-input flex-1 min-w-0">
              <input
                type="text"
                bind:value={formInputs[field.key]}
                bind:this={inputRefs[index]}
                on:keydown={handleKeyDown}
                on:focus={() => handleFieldFocus(index)}
                placeholder={field.placeholder}
                class="terminal-input w-full"
                disabled={field.disabled && field.disabled()}
              />
            </div>
          </div>
        </div>
      {/if}
    {/each}

    </form>
  </div>

  <!-- Help Text Section - Fixed at bottom -->
  <div class="help-section">
    <div class="help-header">
      <div class="help-nav">
        <span class="text-monokai-cyan">↑↓</span> arrows to move | <span class="text-monokai-cyan">Ctrl+Enter</span> to <button type="button" on:click={handleSubmit} class="calculate-link">calculate</button>
      </div>
      <div class="help-field-counter">
        Field <span class="text-monokai-pink">{currentFieldIndex + 1}</span>/<span class="text-monokai-cyan">{visibleFields.length}</span>
      </div>
    </div>
    <div class="help-content">{currentHelpText || 'Navigate through fields using arrow keys'}</div>
  </div>
</div>

<style>
  .terminal-container {
    background: #000;
    padding: 1rem;
    border: 2px solid #2d2d2d;
    border-radius: 0.5rem;
    display: flex;
    flex-direction: column;
    height: 80vh;
    max-height: 80vh;
  }

  .terminal-content {
    flex: 1;
    overflow-y: auto;
    padding-right: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .terminal-content::-webkit-scrollbar {
    width: 6px;
  }

  .terminal-content::-webkit-scrollbar-track {
    background: #0a0a0a;
  }

  .terminal-content::-webkit-scrollbar-thumb {
    background: #2d2d2d;
    border-radius: 4px;
  }

  .terminal-content::-webkit-scrollbar-thumb:hover {
    background: #3d3d3d;
  }

  .terminal-field {
    padding: 0.125rem 0.75rem;
    transition: all 0.15s;
    border-left: 2px solid transparent;
    font-size: 0.8rem;
  }

  .terminal-field.focused {
    background-color: #0a0a0a;
    border-left-color: #FF6188;
  }

  .terminal-field.disabled {
    opacity: 0.4;
    pointer-events: none;
  }

  .field-label {
    font-size: 0.8rem;
  }

  .terminal-input {
    background: transparent;
    border: none;
    outline: none;
    color: #FCFCFA;
    font-family: 'Fira Code', monospace;
    font-size: 0.8rem;
    padding: 0.1rem 0;
  }

  .terminal-input::placeholder {
    color: #5c5c5c;
    font-style: italic;
  }

  .section-header {
    margin-top: 0.5rem;
    margin-bottom: 0.125rem;
    padding: 0.125rem 0.75rem;
    font-weight: bold;
    font-size: 0.75rem;
    letter-spacing: 0.05em;
    border-bottom: 1px solid #2d2d2d;
  }

  .help-section {
    padding: 0.5rem 1rem;
    background-color: #0a0a0a;
    border-top: 1px solid #2d2d2d;
    flex-shrink: 0;
  }

  .help-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.375rem;
    padding-bottom: 0.25rem;
    border-bottom: 1px solid #2d2d2d;
  }

  .help-nav {
    color: #FCFCFA;
    font-size: 0.75rem;
  }

  .help-field-counter {
    color: #939293;
    font-size: 0.7rem;
  }

  .help-content {
    color: #939293;
    font-size: 0.75rem;
    line-height: 1.4;
  }

  .calculate-link {
    background: none;
    border: none;
    color: #78DCE8;
    text-decoration: none;
    cursor: pointer;
    padding: 0;
    font-family: inherit;
    font-size: inherit;
  }
</style>
