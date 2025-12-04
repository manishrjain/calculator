<script lang="ts">
  import { onMount } from 'svelte';
  import type { CalculatorInputs, CalculationResults } from './types';
  import { calculate } from './lib/calculator';
  import TerminalForm from './components/TerminalForm.svelte';
  import ResultsDisplay from './components/ResultsDisplay.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';
  import ShareButton from './components/ShareButton.svelte';
  import SaveDialog from './components/SaveDialog.svelte';
  import LoadDialog from './components/LoadDialog.svelte';
  import { theme } from './lib/theme';
  import { saveConfig, loadConfig } from './lib/storage';
  import { encodeInputsToURL, decodeInputsFromURL, clearURLParams, copyToClipboard } from './lib/share';

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
  let showSaveDialog = false;
  let showLoadDialog = false;
  let shareMessage = '';
  let shareCopied = false;

  function handleGlobalKeyDown(event: KeyboardEvent) {
    // Handle Ctrl+Shift+S to share
    if (event.ctrlKey && event.shiftKey && event.key === 'S') {
      event.preventDefault();
      handleShare();
      return;
    }

    // Handle Ctrl+S to save
    if (event.ctrlKey && event.key === 's') {
      event.preventDefault();
      showSaveDialog = true;
      return;
    }

    // Handle Ctrl+O to load
    if (event.ctrlKey && event.key === 'o') {
      event.preventDefault();
      showLoadDialog = true;
      return;
    }

    // Handle Escape key to go back from results or close dialogs
    if (event.key === 'Escape') {
      if (showSaveDialog) {
        showSaveDialog = false;
        return;
      }
      if (showLoadDialog) {
        showLoadDialog = false;
        return;
      }
      if (showResults) {
        event.preventDefault();
        handleReset();
      }
    }
  }

  // Helper to normalize boolean values
  function normalizeBoolean(val: any) {
    if (typeof val === 'boolean') return val ? 'yes' : 'no';
    if (val === 'true') return 'yes';
    if (val === 'false') return 'no';
    return val;
  }

  function normalizeInputs(inputs: Record<string, any>) {
    if (inputs.includeSelling !== undefined) {
      inputs.includeSelling = normalizeBoolean(inputs.includeSelling);
    }
    if (inputs.includeRentingSell !== undefined) {
      inputs.includeRentingSell = normalizeBoolean(inputs.includeRentingSell);
    }
    if (inputs.include30Year !== undefined) {
      inputs.include30Year = normalizeBoolean(inputs.include30Year);
    }
    return inputs;
  }

  onMount(() => {
    // Initialize theme
    theme.initialize();

    // Check for shared URL params first (takes priority)
    const urlInputs = decodeInputsFromURL();
    if (urlInputs) {
      normalizeInputs(urlInputs);
      formInputs = { ...formInputs, ...urlInputs };
      // Clear URL params after loading
      clearURLParams();
    } else {
      // Load saved inputs from localStorage
      const saved = localStorage.getItem('rentobuy_inputs');
      if (saved) {
        try {
          const loadedInputs = JSON.parse(saved);
          normalizeInputs(loadedInputs);
          formInputs = { ...formInputs, ...loadedInputs };
        } catch (e) {
          console.error('Failed to load saved inputs:', e);
        }
      }
    }

    // Add global keyboard handler
    window.addEventListener('keydown', handleGlobalKeyDown);

    return () => {
      window.removeEventListener('keydown', handleGlobalKeyDown);
    };
  });

  // Update body overflow and scroll behavior
  $: {
    if (typeof document !== 'undefined') {
      // Always allow scrolling so header can scroll away
      document.body.style.overflow = 'auto';

      if (showResults) {
        // Scroll to bottom after results are rendered
        setTimeout(() => {
          window.scrollTo({
            top: document.body.scrollHeight,
            behavior: 'smooth'
          });
        }, 100);
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

  function handleSave(event: CustomEvent<{ name: string }>) {
    const { name } = event.detail;
    saveConfig(name, formInputs);
    showSaveDialog = false;
  }

  function handleLoad(event: CustomEvent<{ name: string; data: Record<string, any> }>) {
    const { data } = event.detail;
    normalizeInputs(data);
    formInputs = { ...formInputs, ...data };
    showLoadDialog = false;
  }

  async function handleShare() {
    const url = encodeInputsToURL(formInputs);
    const success = await copyToClipboard(url);
    if (success) {
      shareMessage = 'Link copied to clipboard!';
      shareCopied = true;
    } else {
      shareMessage = 'Failed to copy. URL: ' + url;
    }
    // Clear message after 3 seconds
    setTimeout(() => {
      shareMessage = '';
      shareCopied = false;
    }, 3000);
  }
</script>

<main class="min-h-screen bg-light-bg dark:bg-black text-light-text dark:text-monokai-text p-4 md:p-8 overflow-x-hidden">
  <div class="max-w-7xl mx-auto w-full">
    <header class="mb-4 md:mb-8">
      <div class="border-2 border-light-border dark:border-monokai-border rounded-lg p-2 md:p-4 bg-light-bg dark:bg-black">
        <div class="flex items-center justify-between mb-1 md:mb-2">
          <div class="flex items-center gap-1 md:gap-2 text-xs font-mono">
            <span class="text-light-pink dark:text-monokai-pink">$</span>
            <span class="text-light-text dark:text-monokai-text">./calculator</span>
          </div>
          <div class="flex items-center gap-1 md:gap-2">
            <ShareButton copied={shareCopied} on:share={handleShare} />
            <ThemeToggle />
          </div>
        </div>
        <h1 class="text-lg md:text-2xl font-bold text-light-orange dark:text-monokai-orange font-mono">
          BRiSK Calculator: Buy v Rent / Sell v Keep
        </h1>
        <div class="mt-1 md:mt-2 text-xs text-light-text-muted dark:text-monokai-text-muted">
          Make a calculated decision
        </div>
      </div>
    </header>

    {#if shareMessage}
      <div class="share-message font-mono text-sm text-light-green dark:text-monokai-green bg-light-bg-light dark:bg-monokai-bg-light border border-light-border dark:border-monokai-border rounded px-4 py-2 mb-4">
        {shareMessage}
      </div>
    {/if}

    {#if !showResults}
      <TerminalForm bind:formInputs on:calculate={handleCalculate} />
    {:else}
      <div class="mb-6">
        <button class="terminal-back-button font-mono" on:click={handleReset}>
          <span class="text-light-pink dark:text-monokai-pink">$</span> cd .. && ./calculator
        </button>
      </div>
      {#if results && calculatedInputs}
        <ResultsDisplay inputs={calculatedInputs} {results} />
      {/if}
    {/if}
  </div>
</main>

<SaveDialog
  isOpen={showSaveDialog}
  on:close={() => showSaveDialog = false}
  on:save={handleSave}
/>

<LoadDialog
  isOpen={showLoadDialog}
  on:close={() => showLoadDialog = false}
  on:load={handleLoad}
/>
