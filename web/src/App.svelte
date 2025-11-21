<script lang="ts">
  import { onMount } from 'svelte';
  import type { CalculatorInputs, CalculationResults } from './types';
  import { calculate } from './lib/calculator';
  import TerminalForm from './components/TerminalForm.svelte';
  import ResultsDisplay from './components/ResultsDisplay.svelte';

  // String versions for form binding
  let formInputs = {
    scenario: 'buy_vs_rent',
    inflationRate: '3',
    investmentReturnRate: '10',
    include30Year: 'no',
    purchasePrice: '500K',
    currentMarketValue: '',
    loanAmount: '400K',
    loanRate: '6.5',
    loanTerm: '30y',
    remainingLoanTerm: '',
    annualInsurance: '3K',
    annualTaxes: '5K',
    monthlyExpenses: '500',
    appreciationRate: '3',
    rentDeposit: '5K',
    monthlyRent: '3K',
    annualRentCosts: '1K',
    otherAnnualCosts: '500',
    includeSelling: 'no',
    includeRentingSell: 'no',
    agentCommission: '6',
    stagingCosts: '10K',
    taxFreeLimits: '250K',
    capitalGainsTax: '20',
  };

  let results: CalculationResults | null = null;
  let calculatedInputs: CalculatorInputs | null = null;
  let showResults = false;

  onMount(() => {
    // Load saved inputs from localStorage
    const saved = localStorage.getItem('rentobuy_inputs');
    if (saved) {
      try {
        const loadedInputs = JSON.parse(saved);
        // Normalize boolean values to 'yes'/'no' strings
        const normalizeBoolean = (val: any) => {
          if (typeof val === 'boolean') return val ? 'yes' : 'no';
          if (val === 'true') return 'yes';
          if (val === 'false') return 'no';
          return val;
        };
        if (loadedInputs.includeSelling !== undefined) {
          loadedInputs.includeSelling = normalizeBoolean(loadedInputs.includeSelling);
        }
        if (loadedInputs.includeRentingSell !== undefined) {
          loadedInputs.includeRentingSell = normalizeBoolean(loadedInputs.includeRentingSell);
        }
        if (loadedInputs.include30Year !== undefined) {
          loadedInputs.include30Year = normalizeBoolean(loadedInputs.include30Year);
        }
        formInputs = { ...formInputs, ...loadedInputs };
      } catch (e) {
        console.error('Failed to load saved inputs:', e);
      }
    }
  });

  // Update body overflow based on whether we're showing results
  $: {
    if (typeof document !== 'undefined') {
      if (showResults) {
        document.body.style.overflow = 'auto';
      } else {
        document.body.style.overflow = 'hidden';
      }
    }
  }

  function handleCalculate(event: CustomEvent) {
    try {
      const inputs: CalculatorInputs = event.detail;
      calculatedInputs = inputs;
      results = calculate(inputs);
      showResults = true;
      // Save form inputs to localStorage
      localStorage.setItem('rentobuy_inputs', JSON.stringify(formInputs));
    } catch (error) {
      console.error('Calculation error:', error);
      alert('Error calculating results. Please check your inputs.');
    }
  }

  function handleReset() {
    showResults = false;
    results = null;
  }
</script>

<main class="min-h-screen bg-black text-monokai-text p-4 md:p-8">
  <div class="max-w-7xl mx-auto">
    <header class="mb-8">
      <div class="border-2 border-monokai-border rounded-lg p-4 bg-black">
        <div class="flex items-center gap-2 mb-2 text-xs font-mono">
          <span class="text-monokai-pink">$</span>
          <span class="text-monokai-text">./calculator</span>
        </div>
        <h1 class="text-2xl font-bold text-monokai-orange font-mono">
          BRSK Calculator: Buy v Rent / Sell v Keep
        </h1>
        <div class="mt-2 text-xs text-monokai-text-muted">
          Make a calculated decision
        </div>
      </div>
    </header>

    {#if !showResults}
      <TerminalForm bind:formInputs on:calculate={handleCalculate} />
    {:else}
      <div class="mb-6">
        <button class="terminal-back-button font-mono" on:click={handleReset}>
          <span class="text-monokai-pink">$</span> cd .. && ./calculator
        </button>
      </div>
      {#if results && calculatedInputs}
        <ResultsDisplay inputs={calculatedInputs} {results} />
      {/if}
    {/if}
  </div>
</main>
