<script lang="ts">
  import { onMount } from 'svelte';
  import type { CalculatorInputs, CalculationResults } from './types';
  import { calculate } from './lib/calculator';
  import TerminalForm from './components/TerminalForm.svelte';
  import ResultsDisplay from './components/ResultsDisplay.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';
  import NumberFormatToggle from './components/NumberFormatToggle.svelte';
  import ShareButton from './components/ShareButton.svelte';
  import CopyReviewButton from './components/CopyReviewButton.svelte';
  import ProjectionToggle from './components/ProjectionToggle.svelte';
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
    annualInsurance: '3K',
    annualTaxes: '5K',
    annualIncome: '0',
    appreciationRate: '3',
    includeRefinance: 'no',
    payoffBalance: '',
    loanAmount: '400K',
    loanRate: '6.5',
    loanTerm: '30y',
    remainingLoanTerm: '',
    closingCosts: '5K',
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
  let reviewCopied = false;
  let showFullNumbers = false;

  // Reactive boolean for 30-year projection toggle
  $: include30Year = formInputs.include30Year === 'yes';

  function handleProjectionToggle() {
    formInputs.include30Year = formInputs.include30Year === 'yes' ? 'no' : 'yes';
    // If on results page, recalculate with new projection setting
    if (showResults && calculatedInputs) {
      calculatedInputs = { ...calculatedInputs, include30Year: formInputs.include30Year === 'yes' };
      results = calculate(calculatedInputs);
    }
  }

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

    // Load number format preference
    const savedNumberFormat = localStorage.getItem('brisk_full_numbers');
    if (savedNumberFormat !== null) {
      showFullNumbers = savedNumberFormat === 'true';
    }

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

  async function handleCopyReview() {
    const resultsElement = document.getElementById('results-content');
    if (!resultsElement) return;

    // Switch to full numbers for LLM review (don't switch back - better for readability)
    showFullNumbers = true;

    // Wait for Svelte to re-render with full numbers
    await new Promise(resolve => setTimeout(resolve, 50));

    const prompt = `Please review these financial calculations from a Buy vs Rent / Sell vs Keep calculator. Verify the math is correct, provide a brief summary of the key findings, and be ready to answer follow-up questions.

---

`;
    const text = prompt + resultsElement.innerText;
    const success = await copyToClipboard(text);

    if (success) {
      shareMessage = 'Results copied! Paste into Claude or Gemini (ChatGPT struggles with calculations).';
      reviewCopied = true;
    } else {
      shareMessage = 'Failed to copy results.';
    }
    // Clear message after 10 seconds
    setTimeout(() => {
      shareMessage = '';
      reviewCopied = false;
    }, 10000);
  }
</script>

<main class="min-h-screen bg-light-bg dark:bg-black text-light-text dark:text-monokai-text p-4 md:p-8 overflow-x-hidden">
  <div class="max-w-7xl mx-auto w-full" class:pt-16={showResults} class:md:pt-14={showResults}>
    <header class="mb-4" class:fixed-header={showResults}>
      <div class="border-2 border-light-border dark:border-monokai-border rounded-lg p-2 bg-light-bg dark:bg-black">
        <div class="flex items-center justify-between">
          <button class="flex items-center gap-1 md:gap-2 text-xs font-mono hover:opacity-70 transition-opacity" on:click={showResults ? handleReset : undefined}>
            <span class="text-light-pink dark:text-monokai-pink">$</span>
            <span class:text-light-cyan={showResults} class:dark:text-monokai-cyan={showResults} class:text-light-text={!showResults} class:dark:text-monokai-text={!showResults}>{showResults ? '../calculator' : './calculator'}</span>
            <span class="hidden md:inline text-light-pink dark:text-monokai-pink font-bold ml-2">BRiSK: Buy v Rent / Sell v Keep</span>
            <span class="hidden lg:inline text-light-text-muted dark:text-monokai-text-muted ml-1">— Make a calculated decision</span>
          </button>
          <div class="flex items-center gap-1 md:gap-2">
            <CopyReviewButton copied={reviewCopied} disabled={!showResults} on:copy={handleCopyReview} />
            <ProjectionToggle {include30Year} on:toggle={handleProjectionToggle} />
            <NumberFormatToggle {showFullNumbers} on:toggle={() => { showFullNumbers = !showFullNumbers; localStorage.setItem('brisk_full_numbers', String(showFullNumbers)); }} />
            <ShareButton copied={shareCopied} on:share={handleShare} />
            <ThemeToggle />
          </div>
        </div>
      </div>
    </header>

    {#if shareMessage}
      <div class="share-message font-mono text-sm text-light-green dark:text-monokai-green bg-light-bg-light dark:bg-monokai-bg-light border border-light-border dark:border-monokai-border rounded px-4 py-2 mb-4">
        {shareMessage}
      </div>
    {/if}

    {#if !showResults}
      <TerminalForm bind:formInputs on:calculate={handleCalculate} on:save={() => showSaveDialog = true} on:load={() => showLoadDialog = true} />
    {:else if results && calculatedInputs}
      <ResultsDisplay inputs={calculatedInputs} {results} {showFullNumbers} />
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

<style>
  .fixed-header {
    position: fixed;
    top: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 100%;
    max-width: 80rem; /* max-w-7xl */
    z-index: 50;
    padding: 0.5rem 1rem;
    @apply bg-light-bg dark:bg-black;
  }

  @media (min-width: 768px) {
    .fixed-header {
      padding: 0.5rem 2rem;
    }
  }
</style>
