<script lang="ts">
  import type { CalculatorInputs, CalculationResults } from '../types';
  import { formatCurrency as formatCurrencyBase, formatPercent } from '../lib/formatter';
  import MarketDataTable from './MarketDataTable.svelte';

  export let inputs: CalculatorInputs;
  export let results: CalculationResults;

  let showFullNumbers = false;

  // Reactive formatter that updates when toggle changes
  $: formatCurrency = (amount: number, forceFullNumbers = false): string => {
    return formatCurrencyBase(amount, forceFullNumbers || showFullNumbers);
  };

  function formatDuration(months: number): string {
    if (months % 12 === 0) {
      return `${months / 12}y`;
    }
    const years = Math.floor(months / 12);
    const remainingMonths = months % 12;
    if (years === 0) {
      return `${remainingMonths}m`;
    }
    return `${years}y ${remainingMonths}m`;
  }

  function formatAppreciationRates(rates: number[]): string {
    if (rates.length === 1) {
      return `${formatPercent(rates[0])} (all years)`;
    }
    return rates.map((rate, i) => {
      const year = i + 1;
      const suffix = i === rates.length - 1 ? `year ${year}+` : `year ${year}`;
      return `${formatPercent(rate)} (${suffix})`;
    }).join(', ');
  }

  // formatTaxFreeLimits needs to be reactive too
  $: formatTaxFreeLimits = (limits: number[]): string => {
    if (limits.length === 1) {
      return `${formatCurrency(limits[0], true)} (all years)`;
    }
    return limits.map((limit, i) => {
      const year = i + 1;
      const suffix = i === limits.length - 1 ? `year ${year}+` : `year ${year}`;
      return `${formatCurrency(limit, true)} (${suffix})`;
    }).join(', ');
  };

  $: downpayment = inputs.purchasePrice - inputs.loanAmount;
</script>

<div class="space-y-8">
  <!-- Number Format Toggle -->
  <div class="flex justify-end font-mono">
    <label class="flex items-center gap-3 cursor-pointer">
      <span class="text-sm text-light-text-muted dark:text-monokai-text-muted">Compact</span>
      <div class="toggle-switch">
        <input type="checkbox" bind:checked={showFullNumbers} />
        <span class="toggle-slider"></span>
      </div>
      <span class="text-sm text-light-text-muted dark:text-monokai-text-muted">Full numbers</span>
    </label>
  </div>

  <!-- Input Parameters Summary -->
  <section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg font-mono">
    <h2 class="text-2xl font-bold text-light-pink dark:text-monokai-pink mb-6">
      {inputs.scenario === 'buy_vs_rent' ? 'INPUT PARAMETERS' : 'INPUT PARAMETERS - SELL VS KEEP'}
    </h2>

    <!-- Economic Assumptions -->
    <div class="mb-6">
      <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">ECONOMIC ASSUMPTIONS</h3>
      <div class="ml-4 space-y-1 text-sm">
        <div><span class="text-light-cyan dark:text-monokai-cyan">Inflation Rate:</span> {formatPercent(inputs.inflationRate)}</div>
        <div><span class="text-light-cyan dark:text-monokai-cyan">Investment Return Rate:</span> {formatPercent(inputs.investmentReturnRate)}</div>
      </div>
    </div>

    {#if inputs.scenario === 'buy_vs_rent'}
      <!-- BUYING Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">BUYING</h3>
        <div class="ml-4 space-y-1 text-sm">
          <div><span class="text-light-cyan dark:text-monokai-cyan">Asset Purchase Price:</span> {formatCurrency(inputs.purchasePrice, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Amount:</span> {formatCurrency(inputs.loanAmount, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Downpayment:</span> {formatCurrency(downpayment, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Rate:</span> {formatPercent(inputs.loanRate)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Duration:</span> {formatDuration(inputs.loanTerm)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Annual Tax & Insurance:</span> {formatCurrency(inputs.annualInsurance, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Other Annual Costs:</span> {formatCurrency(inputs.annualTaxes, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Monthly Expenses:</span> {formatCurrency(inputs.monthlyExpenses, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Appreciation Rate:</span> {formatAppreciationRates(inputs.appreciationRate)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Total Monthly Cost:</span> {formatCurrency(results.monthlyBuyingCosts[0], true)}</div>
        </div>
      </div>

      <!-- RENTING Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">RENTING</h3>
        <div class="ml-4 space-y-1 text-sm">
          <div><span class="text-light-cyan dark:text-monokai-cyan">Rental Deposit:</span> {formatCurrency(inputs.rentDeposit, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Monthly Rent:</span> {formatCurrency(inputs.monthlyRent, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Annual Rent Costs:</span> {formatCurrency(inputs.annualRentCosts, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Other Annual Costs:</span> {formatCurrency(inputs.otherAnnualCosts, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Total Monthly Cost:</span> {formatCurrency(results.monthlyRentingCosts[0], true)}</div>
        </div>
      </div>

      <!-- SELLING Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">SELLING</h3>
        {#if inputs.includeSelling}
          <div class="ml-4 space-y-1 text-sm">
            <div><span class="text-light-cyan dark:text-monokai-cyan">Include Selling Analysis:</span> Yes</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Agent Commission:</span> {formatPercent(inputs.agentCommission)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Staging/Selling Costs:</span> {formatCurrency(inputs.stagingCosts, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Tax-Free Gains Limit:</span> {formatTaxFreeLimits(inputs.taxFreeLimits)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Capital Gains Tax Rate:</span> {formatPercent(inputs.capitalGainsTax)}</div>
          </div>
        {:else}
          <div class="ml-4 text-sm">
            <div><span class="text-light-cyan dark:text-monokai-cyan">Include Selling Analysis:</span> No</div>
          </div>
        {/if}
      </div>
    {:else}
      <!-- SELL VS KEEP: ASSET Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">ASSET</h3>
        <div class="ml-4 space-y-1 text-sm">
          <div><span class="text-light-cyan dark:text-monokai-cyan">Original Purchase Price:</span> {formatCurrency(inputs.purchasePrice, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Current Market Value:</span> {formatCurrency(inputs.currentMarketValue || 0, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Current Equity:</span> {formatCurrency(downpayment, true)}</div>
          {#if inputs.loanAmount > 0}
            <div><span class="text-light-cyan dark:text-monokai-cyan">Remaining Loan Balance:</span> {formatCurrency(inputs.loanAmount, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Rate:</span> {formatPercent(inputs.loanRate)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Remaining Loan Term:</span> {formatDuration(inputs.remainingLoanTerm || inputs.loanTerm)}</div>
          {:else}
            <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Status:</span> Fully paid off</div>
          {/if}
          <div><span class="text-light-cyan dark:text-monokai-cyan">Annual Tax & Insurance:</span> {formatCurrency(inputs.annualInsurance, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Other Annual Costs:</span> {formatCurrency(inputs.annualTaxes, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Monthly Expenses:</span> {formatCurrency(inputs.monthlyExpenses, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Appreciation Rate (if keeping):</span> {formatAppreciationRates(inputs.appreciationRate)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Total Monthly Cost (if keeping):</span> {formatCurrency(results.monthlyBuyingCosts[0], true)}</div>
        </div>
      </div>

      <!-- INVESTING (if selling) Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">INVESTING (if selling)</h3>
        {#if inputs.includeRentingSell}
          <div class="ml-4 space-y-1 text-sm">
            <div><span class="text-light-cyan dark:text-monokai-cyan">Include Renting Analysis:</span> Yes</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Rental Deposit:</span> {formatCurrency(inputs.rentDeposit, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Monthly Rent:</span> {formatCurrency(inputs.monthlyRent, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Annual Rent Costs:</span> {formatCurrency(inputs.annualRentCosts, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Total Monthly Renting Cost:</span> {formatCurrency(results.monthlyRentingCosts[0], true)}</div>
          </div>
        {:else}
          <div class="ml-4 text-sm">
            <div><span class="text-light-cyan dark:text-monokai-cyan">Include Renting Analysis:</span> No</div>
          </div>
        {/if}
      </div>

      <!-- SELLING COSTS Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">SELLING COSTS</h3>
        <div class="ml-4 space-y-1 text-sm">
          <div><span class="text-light-cyan dark:text-monokai-cyan">Agent Commission:</span> {formatPercent(inputs.agentCommission)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Staging/Selling Costs:</span> {formatCurrency(inputs.stagingCosts, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Tax-Free Gains Limit:</span> {formatTaxFreeLimits(inputs.taxFreeLimits)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Capital Gains Tax Rate:</span> {formatPercent(inputs.capitalGainsTax)}</div>
        </div>
      </div>
    {/if}
  </section>

  <!-- Market Data Reference -->
  <MarketDataTable />

  <!-- Amortization Table -->
  {#if results.amortizationTable && inputs.loanAmount > 0}
    <section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
      <h2 class="section-title">Loan Amortization Details</h2>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Period</th>
              <th class="text-right">Principal Paid</th>
              <th class="text-right">Interest Paid</th>
              <th class="text-right">Loan Balance</th>
            </tr>
          </thead>
          <tbody>
            {#each results.amortizationTable as row}
              <tr>
                <td class="font-mono">{row.period}</td>
                <td class="text-right font-mono">{formatCurrency(row.principalPaid)}</td>
                <td class="text-right font-mono">{formatCurrency(row.interestPaid)}</td>
                <td class="text-right font-mono">{formatCurrency(row.loanBalance)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <p class="help-text mt-2">
        Note: Monthly payment is fixed. Each payment covers interest on remaining balance, with the rest going to principal.
      </p>
    </section>
  {/if}

  <!-- Expenditure Table (BUY vs RENT only) -->
  {#if results.expenditureTable}
    <section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
      <h2 class="section-title">Total Expenditure Comparison</h2>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Period</th>
              <th class="text-right">Buying Expend.</th>
              <th class="text-right">Renting Expend.</th>
              <th class="text-right">Difference</th>
            </tr>
          </thead>
          <tbody>
            {#each results.expenditureTable as row}
              <tr>
                <td class="font-mono">{row.period}</td>
                <td class="text-right font-mono">{formatCurrency(row.buyingExpenditure)}</td>
                <td class="text-right font-mono">{formatCurrency(row.rentingExpenditure)}</td>
                <td class="text-right font-mono">{formatCurrency(row.difference)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <p class="help-text mt-2">
        Note: All recurring costs (insurance, taxes, rent, HOA, etc.) are inflated annually at {formatPercent(inputs.inflationRate)} rate.
      </p>
    </section>
  {/if}

  <!-- Keep Expenses Breakdown (SELL vs KEEP only) -->
  {#if results.keepExpensesTable}
    <section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
      <h2 class="section-title">Keep Expenses Breakdown</h2>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Period</th>
              <th class="text-right">Loan Payment</th>
              <th class="text-right">Tax/Insurance</th>
              <th class="text-right">Other Costs</th>
              <th class="text-right">Cumulative Exp</th>
              <th class="text-right">Investment Val</th>
              <th class="text-right">Net Position</th>
            </tr>
          </thead>
          <tbody>
            {#each results.keepExpensesTable as row}
              <tr>
                <td class="font-mono">{row.period}</td>
                <td class="text-right font-mono">{formatCurrency(row.loanPayment)}</td>
                <td class="text-right font-mono">{formatCurrency(row.taxInsurance)}</td>
                <td class="text-right font-mono">{formatCurrency(row.otherCosts)}</td>
                <td class="text-right font-mono">{formatCurrency(row.cumulativeExp)}</td>
                <td class="text-right font-mono">{formatCurrency(row.investmentVal)}</td>
                <td class="text-right font-mono">{formatCurrency(row.netPosition)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <p class="help-text mt-2">
        Note: Shows annual expenses for the specific year of each period. 'Loan Payment' = Loan payments for that year (stops after loan term).
        'Tax/Insurance' = Annual tax & insurance (inflated at {formatPercent(inputs.inflationRate)} annually).
        'Other Costs' = Other annual costs + monthly expenses (inflated).
        'Cumulative Exp' = Running total of raw expenses.
        'Investment Val' = Value of invested income (compounded at {formatPercent(inputs.investmentReturnRate)} return).
        'Net Position' = Investment value minus real out-of-pocket costs.
      </p>
    </section>
  {/if}

  <!-- Sale Proceeds -->
  <section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
    <h2 class="section-title">Sale Proceeds Analysis</h2>
    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>Period</th>
            <th class="text-right">Sale Price</th>
            <th class="text-right">Selling Cost</th>
            <th class="text-right">Loan Payoff</th>
            <th class="text-right">Cap Gains</th>
            <th class="text-right">Tax</th>
            <th class="text-right">Net Proceeds</th>
          </tr>
        </thead>
        <tbody>
          {#each results.saleProceedsTable as row}
            <tr>
              <td class="font-mono">{row.period}</td>
              <td class="text-right font-mono">{formatCurrency(row.salePrice)}</td>
              <td class="text-right font-mono">{formatCurrency(row.totalSellingCosts)}</td>
              <td class="text-right font-mono">{formatCurrency(row.loanPayoff)}</td>
              <td class="text-right font-mono">{formatCurrency(row.capitalGains)}</td>
              <td class="text-right font-mono">{formatCurrency(row.taxOnGains)}</td>
              <td class="text-right font-mono">{formatCurrency(row.netProceeds)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
    <p class="help-text mt-2">
      Note: Appreciation rates are applied year-by-year (compounded). Sale price = compounded property value.
    </p>
  </section>

  <!-- Comparison Table (BUY vs RENT) -->
  {#if results.comparisonTable}
    <section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
      <h2 class="section-title">Net Worth Projections: BUY vs RENT</h2>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Period</th>
              <th class="text-right">Asset Value</th>
              <th class="text-right text-light-pink dark:text-monokai-pink">Buying NW</th>
              <th class="text-right">Cum Savings</th>
              <th class="text-right">Market Return</th>
              <th class="text-right text-light-green dark:text-monokai-green">Renting NW</th>
              <th class="text-right"><span class="text-light-green dark:text-monokai-green">RENT</span> - <span class="text-light-pink dark:text-monokai-pink">BUY</span></th>
            </tr>
          </thead>
          <tbody>
            {#each results.comparisonTable as row}
              <tr>
                <td class="font-mono">{row.period}</td>
                <td class="text-right font-mono">{formatCurrency(row.assetValue)}</td>
                <td class="text-right font-mono text-light-pink dark:text-monokai-pink">{formatCurrency(row.buyingNetWorth)}</td>
                <td class="text-right font-mono">{formatCurrency(row.cumulativeSavings)}</td>
                <td class="text-right font-mono">{formatCurrency(row.marketReturn)}</td>
                <td class="text-right font-mono text-light-green dark:text-monokai-green">{formatCurrency(row.rentingNetWorth)}</td>
                <td class="text-right font-mono" class:text-light-green={row.difference > 0} class:dark:text-monokai-green={row.difference > 0} class:text-light-pink={row.difference < 0} class:dark:text-monokai-pink={row.difference < 0}>
                  {formatCurrency(row.difference)}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <p class="help-text mt-2">
        Note: 'Cum Savings' = raw cost difference without investment growth.
        'Market Return' = investment growth at {formatPercent(inputs.investmentReturnRate)} annual rate.
        Positive RENT - BUY means renting wins, negative means buying wins.
      </p>
    </section>
  {/if}

  <!-- Sell vs Keep Comparison -->
  {#if results.sellVsKeepTable}
    <section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
      <h2 class="section-title">Net Worth Projections: SELL vs KEEP</h2>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Period</th>
              {#if inputs.includeRentingSell}
                <th class="text-right">SELL Cum. Exp</th>
              {/if}
              <th class="text-right text-light-pink dark:text-monokai-pink">SELL Net Worth</th>
              <th class="text-right">KEEP Net Position</th>
              <th class="text-right text-light-green dark:text-monokai-green">KEEP Net Proceeds</th>
              <th class="text-right"><span class="text-light-green dark:text-monokai-green">KEEP</span> - <span class="text-light-pink dark:text-monokai-pink">SELL</span></th>
            </tr>
          </thead>
          <tbody>
            {#each results.sellVsKeepTable as row}
              <tr>
                <td class="font-mono">{row.period}</td>
                {#if row.sellCumulativeExpenses !== undefined}
                  <td class="text-right font-mono">{formatCurrency(row.sellCumulativeExpenses)}</td>
                {/if}
                <td class="text-right font-mono text-light-pink dark:text-monokai-pink">{formatCurrency(row.sellNetWorth)}</td>
                <td class="text-right font-mono">{formatCurrency(row.keepNetPosition)}</td>
                <td class="text-right font-mono text-light-green dark:text-monokai-green">{formatCurrency(row.keepNetProceeds)}</td>
                <td class="text-right font-mono" class:text-light-green={row.difference > 0} class:dark:text-monokai-green={row.difference > 0} class:text-light-pink={row.difference < 0} class:dark:text-monokai-pink={row.difference < 0}>
                  {formatCurrency(row.difference)}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <p class="help-text mt-2">
        Note: 'SELL Net Worth' = Net proceeds from selling now invested at {formatPercent(inputs.investmentReturnRate)} return.
        'KEEP Net Position' = Investment value from income minus real costs.
        Positive KEEP - SELL means keeping wins, negative means selling wins.
      </p>
    </section>
  {/if}
</div>

