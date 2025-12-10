<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { CalculatorInputs, ScenarioType } from '../types';
  import { parseAmount, parseDuration, parseAppreciationRates } from '../lib/formatter';

  export let formInputs: any;

  const dispatch = createEventDispatcher();

  function convertFormInputsToCalculatorInputs(): CalculatorInputs {
    return {
      scenario: formInputs.scenario as ScenarioType,
      inflationRate: parseFloat(formInputs.inflationRate) || 0,
      investmentReturnRate: parseFloat(formInputs.investmentReturnRate) || 0,
      projectionYears: parseInt(formInputs.projectionYears) || 10,
      purchasePrice: parseAmount(formInputs.purchasePrice.toString()),
      currentMarketValue: formInputs.currentMarketValue ? parseAmount(formInputs.currentMarketValue.toString()) : undefined,
      loanAmount: parseAmount(formInputs.loanAmount.toString()),
      loanRate: parseFloat(formInputs.loanRate) || 0,
      loanTerm: parseDuration(formInputs.loanTerm),
      remainingLoanTerm: formInputs.remainingLoanTerm ? parseDuration(formInputs.remainingLoanTerm) : undefined,
      mortgageInterestDeduction: parseFloat(formInputs.mortgageInterestDeduction) || 0,
      annualInsurance: parseAmount(formInputs.annualInsurance.toString()),
      annualTaxes: parseAmount(formInputs.annualTaxes.toString()),
      annualIncome: parseAmount(formInputs.annualIncome.toString()),
      appreciationRate: parseAppreciationRates(formInputs.appreciationRate),
      rentDeposit: parseAmount(formInputs.rentDeposit.toString()),
      monthlyRent: parseAmount(formInputs.monthlyRent.toString()),
      annualRentCosts: parseAmount(formInputs.annualRentCosts.toString()),
      otherAnnualCosts: parseAmount(formInputs.otherAnnualCosts.toString()),
      includeSelling: formInputs.includeSelling,
      includeRentingSell: formInputs.includeRentingSell,
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
</script>

<form on:submit|preventDefault={handleSubmit} class="space-y-6 max-w-5xl mx-auto">
  <!-- Scenario Selection -->
  <section class="form-section">
    <h2 class="group-title">
      Scenario Selection
    </h2>
    <div class="flex gap-4 flex-wrap">
      <label class="flex-1 min-w-[200px]">
        <input
          type="radio"
          name="scenario"
          value="buy_vs_rent"
          bind:group={formInputs.scenario}
          class="peer sr-only"
        />
        <div class="px-6 py-4 rounded-lg border-2 border-monokai-border cursor-pointer transition-all duration-200
                    peer-checked:border-monokai-pink peer-checked:bg-monokai-pink/10
                    hover:border-monokai-pink/50 hover:bg-monokai-bg-light">
          <div class="font-bold text-lg text-monokai-text">BUY vs RENT</div>
          <div class="text-xs text-monokai-text-muted mt-1">Compare purchasing vs renting</div>
        </div>
      </label>
      <label class="flex-1 min-w-[200px]">
        <input
          type="radio"
          name="scenario"
          value="sell_vs_keep"
          bind:group={formInputs.scenario}
          class="peer sr-only"
        />
        <div class="px-6 py-4 rounded-lg border-2 border-monokai-border cursor-pointer transition-all duration-200
                    peer-checked:border-monokai-cyan peer-checked:bg-monokai-cyan/10
                    hover:border-monokai-cyan/50 hover:bg-monokai-bg-light">
          <div class="font-bold text-lg text-monokai-text">SELL vs KEEP</div>
          <div class="text-xs text-monokai-text-muted mt-1">Analyze selling vs keeping</div>
        </div>
      </label>
    </div>
  </section>

  <!-- Economic Assumptions -->
  <section class="form-section">
    <h2 class="group-title">
      Economic Assumptions
    </h2>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-2">
        <label class="label">
          Inflation Rate (%)
        </label>
        <input
          type="text"
          bind:value={formInputs.inflationRate}
          class="input-field w-full font-mono text-lg"
          placeholder="3"
        />
        <p class="help-text">Annual inflation for all recurring costs</p>
      </div>
      <div class="space-y-2">
        <label class="label">
          Investment Return Rate (%)
        </label>
        <input
          type="text"
          bind:value={formInputs.investmentReturnRate}
          class="input-field w-full font-mono text-lg"
          placeholder="10"
        />
        <p class="help-text">Expected return on investments</p>
      </div>
      <div class="md:col-span-2">
        <label class="flex items-center gap-3 cursor-pointer p-4 rounded-lg border-2 border-monokai-border hover:border-monokai-cyan/50 transition-all">
          <input
            type="checkbox"
            bind:checked={formInputs.include30Year}
            class="w-5 h-5 rounded border-monokai-border text-monokai-cyan focus:ring-monokai-cyan focus:ring-2"
          />
          <div>
            <span class="font-semibold text-monokai-text">Include 30-Year Projections</span>
            <p class="help-text mt-0">Show 15y, 20y, 30y periods (default: 10y max)</p>
          </div>
        </label>
      </div>
    </div>
  </section>

  <!-- Buying/Asset Section -->
  <section class="form-section">
    <h2 class="group-title">
      {formInputs.scenario === 'sell_vs_keep' ? 'Asset Details' : 'Buying Details'}
    </h2>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-2">
        <label class="label">
          {formInputs.scenario === 'sell_vs_keep' ? 'Original Purchase Price' : 'Asset Purchase Price'}
        </label>
        <input
          type="text"
          bind:value={formInputs.purchasePrice}
          class="input-field w-full font-mono text-lg"
          placeholder="500K or 500000"
        />
        <p class="help-text">
          {formInputs.scenario === 'sell_vs_keep' ? 'What you originally paid (for capital gains)' : 'Initial purchase price. Use K/M suffix (e.g., 500K)'}
        </p>
      </div>

      {#if formInputs.scenario === 'sell_vs_keep'}
        <div class="space-y-2">
          <label class="label">
            Current Market Value
          </label>
          <input
            type="text"
            bind:value={formInputs.currentMarketValue}
            class="input-field w-full font-mono text-lg"
            placeholder="2.2M"
          />
          <p class="help-text">What the asset is worth today</p>
        </div>
      {/if}

      <div class="space-y-2">
        <label class="label">
          {formInputs.scenario === 'sell_vs_keep' ? 'Original Loan Amount' : 'Loan Amount'}
        </label>
        <input
          type="text"
          bind:value={formInputs.loanAmount}
          class="input-field w-full font-mono text-lg"
          placeholder="400K"
        />
        <p class="help-text">Total mortgage/loan amount</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Loan Rate (%)
        </label>
        <input
          type="text"
          bind:value={formInputs.loanRate}
          class="input-field w-full font-mono text-lg"
          placeholder="6.5"
        />
        <p class="help-text">Annual interest rate</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Loan Term
        </label>
        <input
          type="text"
          bind:value={formInputs.loanTerm}
          placeholder="30y or 360m"
          class="input-field w-full font-mono text-lg"
        />
        <p class="help-text">Duration (e.g., 30y, 15y)</p>
      </div>

      {#if formInputs.scenario === 'sell_vs_keep'}
        <div class="space-y-2">
          <label class="label">
            Remaining Loan Term
          </label>
          <input
            type="text"
            bind:value={formInputs.remainingLoanTerm}
            placeholder="25y"
            class="input-field w-full font-mono text-lg"
          />
          <p class="help-text">Time left on loan</p>
        </div>
      {/if}

      <div class="space-y-2">
        <label class="label">
          Annual Tax & Insurance
        </label>
        <input
          type="text"
          bind:value={formInputs.annualInsurance}
          class="input-field w-full font-mono text-lg"
          placeholder="3K"
        />
        <p class="help-text">Yearly insurance cost</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Other Annual Costs
        </label>
        <input
          type="text"
          bind:value={formInputs.annualTaxes}
          class="input-field w-full font-mono text-lg"
          placeholder="5K"
        />
        <p class="help-text">Maintenance, utilities, HOA, etc.</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Annual Income from Asset
        </label>
        <input
          type="text"
          bind:value={formInputs.annualIncome}
          class="input-field w-full font-mono text-lg"
          placeholder="0"
        />
        <p class="help-text">Rental income or other income from the asset</p>
      </div>

      <div class="md:col-span-2 space-y-2">
        <label class="label">
          Appreciation Rate (%)
        </label>
        <input
          type="text"
          bind:value={formInputs.appreciationRate}
          placeholder="3 or 5,3,2"
          class="input-field w-full font-mono text-lg"
        />
        <p class="help-text">
          Annual rate (comma-separated for different years, e.g., '10,5,3' = 10% yr1, 5% yr2, 3% yr3+)
        </p>
      </div>
    </div>
  </section>

  <!-- Renting Section -->
  <section class="form-section">
    <h2 class="group-title">
      {formInputs.scenario === 'sell_vs_keep' ? 'Investing (if selling)' : 'Renting'}
    </h2>

    {#if formInputs.scenario === 'sell_vs_keep'}
      <div class="mb-6">
        <label class="flex items-center gap-3 cursor-pointer p-4 rounded-lg border-2 border-monokai-border hover:border-monokai-cyan/50 transition-all">
          <input
            type="checkbox"
            bind:checked={formInputs.includeRentingSell}
            class="w-5 h-5 rounded border-monokai-border text-monokai-cyan focus:ring-monokai-cyan focus:ring-2"
          />
          <div>
            <span class="font-semibold text-monokai-text">Include Renting Analysis</span>
            <p class="help-text mt-0">Toggle if selling means you'll need to rent</p>
          </div>
        </label>
      </div>
    {/if}

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-2">
        <label class="label">
          Rental Deposit
        </label>
        <input
          type="text"
          bind:value={formInputs.rentDeposit}
          class="input-field w-full font-mono text-lg"
          placeholder="5K"
        />
        <p class="help-text">Initial rental deposit</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Monthly Rent
        </label>
        <input
          type="text"
          bind:value={formInputs.monthlyRent}
          class="input-field w-full font-mono text-lg"
          placeholder="3K"
        />
        <p class="help-text">Base monthly rent amount</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Annual Rent Costs
        </label>
        <input
          type="text"
          bind:value={formInputs.annualRentCosts}
          class="input-field w-full font-mono text-lg"
          placeholder="1K"
        />
        <p class="help-text">Yearly rental-related costs</p>
      </div>

      {#if formInputs.scenario === 'buy_vs_rent'}
        <div class="space-y-2">
          <label class="label">
            Other Annual Costs
          </label>
          <input
            type="text"
            bind:value={formInputs.otherAnnualCosts}
            class="input-field w-full font-mono text-lg"
            placeholder="500"
          />
          <p class="help-text">Additional yearly costs for renting</p>
        </div>
      {/if}
    </div>
  </section>

  <!-- Selling Section -->
  <section class="form-section">
    <h2 class="group-title">
      Selling
    </h2>

    {#if formInputs.scenario === 'buy_vs_rent'}
      <div class="mb-6">
        <label class="flex items-center gap-3 cursor-pointer p-4 rounded-lg border-2 border-monokai-border hover:border-monokai-cyan/50 transition-all">
          <input
            type="checkbox"
            bind:checked={formInputs.includeSelling}
            class="w-5 h-5 rounded border-monokai-border text-monokai-cyan focus:ring-monokai-cyan focus:ring-2"
          />
          <div>
            <span class="font-semibold text-monokai-text">Include Selling Analysis</span>
            <p class="help-text mt-0">Toggle to enable/disable selling analysis</p>
          </div>
        </label>
      </div>
    {/if}

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="space-y-2">
        <label class="label">
          Agent Commission (%)
        </label>
        <input
          type="text"
          bind:value={formInputs.agentCommission}
          class="input-field w-full font-mono text-lg"
          placeholder="6"
        />
        <p class="help-text">Percentage of sale price</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Staging/Selling Costs
        </label>
        <input
          type="text"
          bind:value={formInputs.stagingCosts}
          class="input-field w-full font-mono text-lg"
          placeholder="10K"
        />
        <p class="help-text">Fixed costs to prepare and sell</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Tax-Free Gains Limit
        </label>
        <input
          type="text"
          bind:value={formInputs.taxFreeLimits}
          placeholder="250K or 500K,0"
          class="input-field w-full font-mono text-lg"
        />
        <p class="help-text">Capital gains exempt from tax (comma-separated for different years)</p>
      </div>

      <div class="space-y-2">
        <label class="label">
          Capital Gains Tax Rate (%)
        </label>
        <input
          type="text"
          bind:value={formInputs.capitalGainsTax}
          class="input-field w-full font-mono text-lg"
          placeholder="20"
        />
        <p class="help-text">Long-term capital gains tax rate</p>
      </div>
    </div>
  </section>

  <!-- Submit Button -->
  <div class="flex justify-center pt-4">
    <button type="submit" class="btn-primary font-mono">
      $ ./calculate --run
    </button>
  </div>
</form>
