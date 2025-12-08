<script lang="ts">
  import type { CalculatorInputs, CalculationResults } from '../types';
  import { formatCurrency as formatCurrencyBase, formatPercent } from '../lib/formatter';
  import { getEffectiveLoanValues } from '../lib/calculator';
  import MarketDataTable from './MarketDataTable.svelte';

  export let inputs: CalculatorInputs;
  export let results: CalculationResults;
  export let showFullNumbers = false;

  // Reactive formatter that updates when toggle changes
  $: formatCurrency = (amount: number, forceFullNumbers = false): string => {
    return formatCurrencyBase(amount, forceFullNumbers || showFullNumbers);
  };

  function scrollToSection(event: MouseEvent, targetId: string) {
    event.preventDefault();
    const target = document.getElementById(targetId);
    if (target) {
      target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  }

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
  $: effectiveLoanValues = getEffectiveLoanValues(inputs);
  $: showSellingColumns = inputs.scenario === 'sell_vs_keep' || inputs.includeSelling;
</script>

<div id="results-content" class="space-y-8">
  <!-- Input Parameters Summary -->
  <section class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg font-mono">
    <h2 class="text-2xl font-bold text-light-pink dark:text-monokai-pink mb-6">INPUT PARAMETERS</h2>

    <!-- Scenario -->
    <div class="mb-6">
      <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">SCENARIO</h3>
      <div class="ml-4 space-y-1 text-sm">
        <div><span class="text-light-cyan dark:text-monokai-cyan">Analysis Type:</span> {inputs.scenario === 'buy_vs_rent' ? 'Buy vs Rent' : 'Sell vs Keep'}</div>
      </div>
    </div>

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
          <div><span class="text-light-cyan dark:text-monokai-cyan">Downpayment:</span> {formatCurrency(downpayment, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Annual Tax & Insurance:</span> {formatCurrency(inputs.annualInsurance, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Other Annual Costs:</span> {formatCurrency(inputs.annualTaxes, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Annual Income from Asset:</span> {formatCurrency(inputs.annualIncome, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Appreciation Rate:</span> {formatAppreciationRates(inputs.appreciationRate)}</div>
        </div>
      </div>

      <!-- LOAN Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">LOAN</h3>
        <div class="ml-4 space-y-1 text-sm">
          <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Amount:</span> {formatCurrency(inputs.loanAmount, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Rate:</span> {formatPercent(inputs.loanRate)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Duration:</span> {formatDuration(inputs.loanTerm)}</div>
          {#if inputs.mortgageInterestDeduction > 0}
            <div><span class="text-light-cyan dark:text-monokai-cyan">Mortgage Interest Deduction:</span> {formatPercent(inputs.mortgageInterestDeduction)}</div>
          {/if}
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
          <div><span class="text-light-cyan dark:text-monokai-cyan">Annual Tax & Insurance:</span> {formatCurrency(inputs.annualInsurance, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Other Annual Costs:</span> {formatCurrency(inputs.annualTaxes, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Annual Income from Asset:</span> {formatCurrency(inputs.annualIncome, true)}</div>
          <div><span class="text-light-cyan dark:text-monokai-cyan">Appreciation Rate (if keeping):</span> {formatAppreciationRates(inputs.appreciationRate)}</div>
        </div>
      </div>

      <!-- LOAN Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">LOAN</h3>
        <div class="ml-4 space-y-1 text-sm">
          {#if inputs.includeRefinance}
            <div><span class="text-light-cyan dark:text-monokai-cyan">Include Refinance:</span> Yes</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Payoff Balance:</span> {formatCurrency(inputs.payoffBalance || 0, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Refinance Loan Amount:</span> {formatCurrency(inputs.loanAmount, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Refinance Rate:</span> {formatPercent(inputs.loanRate)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Refinance Term:</span> {formatDuration(inputs.loanTerm)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Closing Costs:</span> {formatCurrency(inputs.closingCosts || 0, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Cash Out:</span> {formatCurrency(effectiveLoanValues.refinanceCashOut, true)}</div>
          {:else if inputs.loanAmount > 0}
            <div><span class="text-light-cyan dark:text-monokai-cyan">Include Refinance:</span> No</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Original Loan Amount:</span> {formatCurrency(inputs.loanAmount, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Remaining Loan Balance:</span> {formatCurrency(effectiveLoanValues.effectiveLoanAmount, true)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Rate:</span> {formatPercent(inputs.loanRate)}</div>
            <div><span class="text-light-cyan dark:text-monokai-cyan">Remaining Loan Term:</span> {formatDuration(inputs.remainingLoanTerm || inputs.loanTerm)}</div>
          {:else}
            <div><span class="text-light-cyan dark:text-monokai-cyan">Loan Status:</span> Fully paid off</div>
          {/if}
          {#if inputs.mortgageInterestDeduction > 0}
            <div><span class="text-light-cyan dark:text-monokai-cyan">Mortgage Interest Deduction:</span> {formatPercent(inputs.mortgageInterestDeduction)}</div>
          {/if}
        </div>
      </div>

      <!-- RENTING (if selling) Section -->
      <div class="mb-6">
        <h3 class="text-light-orange dark:text-monokai-orange font-bold mb-2">RENTING</h3>
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
    <section id="loan-amortization" class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
      <h2 class="section-title">Loan Amortization</h2>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Period</th>
              <th class="text-right">Principal Paid</th>
              <th class="text-right">Interest Paid</th>
              {#if inputs.mortgageInterestDeduction > 0}
                <th class="text-right">Tax Deduction</th>
                <th class="text-right">Eff. Interest</th>
                <th class="text-right">Eff. Loan Pmt ③</th>
              {/if}
              <th class="text-right">Loan Balance</th>
            </tr>
          </thead>
          <tbody>
            {#each results.amortizationTable as row}
              <tr>
                <td class="font-mono">{row.period}</td>
                <td class="text-right font-mono">{formatCurrency(row.principalPaid)}</td>
                <td class="text-right font-mono">{formatCurrency(row.interestPaid)}</td>
                {#if inputs.mortgageInterestDeduction > 0}
                  <td class="text-right font-mono">-{formatCurrency(row.taxDeduction)}</td>
                  <td class="text-right font-mono">{formatCurrency(row.effectiveInterest)}</td>
                  <td class="text-right font-mono">{formatCurrency(row.effectiveLoanPayment)}</td>
                {/if}
                <td class="text-right font-mono">{formatCurrency(row.loanBalance)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="help-text mt-2">
        <p>Note: Monthly payment is fixed. Each payment covers interest on remaining balance, with the rest going to principal.</p>
        <div class="grid grid-cols-[auto_1fr] gap-x-2">
          <span class="text-light-cyan dark:text-monokai-cyan">Principal</span><span>= Cumulative amount paid toward the loan balance.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Interest</span><span>= Cumulative interest paid to the lender.</span>
          {#if inputs.mortgageInterestDeduction > 0}
            <span class="text-light-cyan dark:text-monokai-cyan">Tax Deduction</span><span>= Interest × {formatPercent(inputs.mortgageInterestDeduction)} (your mortgage interest deduction rate).</span>
            <span class="text-light-cyan dark:text-monokai-cyan">Eff. Interest</span><span>= Interest - Tax Deduction.</span>
            <span class="text-light-cyan dark:text-monokai-cyan">Eff. Loan Pmt ③</span><span>= Principal + Eff. Interest (cumulative amount actually paid after tax savings).</span>
          {/if}
          <span class="text-light-cyan dark:text-monokai-cyan">Loan Balance</span><span>= Remaining amount owed on the loan.</span>
        </div>
      </div>
    </section>
  {/if}

  <!-- Expenditure Table (BUY vs RENT only) -->
  {#if results.expenditureTable}
    <section id="expenditure-comparison" class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
      <h2 class="section-title">Total Expenditure Comparison</h2>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Period</th>
              <th class="text-right">{#if inputs.mortgageInterestDeduction > 0}<a href="#loan-amortization" class="underline" on:click={(e) => scrollToSection(e, 'loan-amortization')}>Eff. Loan Pmt ③</a>{:else}Loan Payment{/if}</th>
              <th class="text-right">Costs</th>
              <th class="text-right">Buying Expend.</th>
              <th class="text-right">Renting Expend.</th>
              <th class="text-right"><span class="text-light-pink dark:text-monokai-pink">BUY</span> - <span class="text-light-green dark:text-monokai-green">RENT</span> ②</th>
            </tr>
          </thead>
          <tbody>
            {#each results.expenditureTable as row}
              <tr>
                <td class="font-mono">{row.period}</td>
                <td class="text-right font-mono">{formatCurrency(inputs.mortgageInterestDeduction > 0 ? row.effectiveLoanPayment : row.loanPayment)}</td>
                <td class="text-right font-mono">{formatCurrency(row.costs)}</td>
                <td class="text-right font-mono">{formatCurrency(row.buyingExpenditure)}</td>
                <td class="text-right font-mono">{formatCurrency(row.rentingExpenditure)}</td>
                <td class="text-right font-mono">{formatCurrency(row.difference)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="help-text mt-2">
        <p>Note: All recurring costs (insurance, taxes, rent, HOA, etc.) are inflated annually at {formatPercent(inputs.inflationRate)} rate.</p>
        <div class="grid grid-cols-[auto_1fr] gap-x-2">
          {#if inputs.mortgageInterestDeduction > 0}
            <span class="text-light-cyan dark:text-monokai-cyan">Eff. Loan Pmt ③</span><span>= Annual effective loan payment after tax deduction (from <a href="#loan-amortization" class="underline" on:click={(e) => scrollToSection(e, 'loan-amortization')}>Loan Amortization</a>).</span>
          {:else}
            <span class="text-light-cyan dark:text-monokai-cyan">Loan Payment</span><span>= Annual loan payments for that year (stops after loan term).</span>
          {/if}
          <span class="text-light-cyan dark:text-monokai-cyan">Costs</span><span>= Annual taxes, insurance, and other costs (inflated). Negative means income exceeds costs.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Buying Expend.</span><span>= Cumulative buying costs (downpayment + loan payments + costs).</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Renting Expend.</span><span>= Cumulative renting costs (deposit + rent + annual rent costs).</span>
          <span class="text-light-cyan dark:text-monokai-cyan">BUY - RENT ②</span><span>= Expenditure difference. Negative means renting costs more; this difference is invested as savings.</span>
        </div>
      </div>
    </section>
  {/if}

  <!-- Keep Expenses Breakdown (SELL vs KEEP only) -->
  {#if results.keepExpensesTable}
    <section id="keep-invest-position" class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
      <h2 class="section-title">KEEP Analysis: Invest Position</h2>
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Period</th>
              <th class="text-right">{#if inputs.mortgageInterestDeduction > 0}<a href="#loan-amortization" class="underline" on:click={(e) => scrollToSection(e, 'loan-amortization')}>Eff. Loan Pmt ③</a>{:else}Loan Payment{/if}</th>
              <th class="text-right">Costs</th>
              <th class="text-right">Cumulative Exp</th>
              <th class="text-right">Invest Returns</th>
              <th class="text-right">Investment Val</th>
              <th class="text-right">Net Position ②</th>
            </tr>
          </thead>
          <tbody>
            {#each results.keepExpensesTable as row}
              <tr>
                <td class="font-mono">{row.period}</td>
                <td class="text-right font-mono">{formatCurrency(inputs.mortgageInterestDeduction > 0 ? row.effectiveLoanPayment : row.loanPayment)}</td>
                <td class="text-right font-mono">{formatCurrency(row.taxInsurance + row.otherCosts)}</td>
                <td class="text-right font-mono">{formatCurrency(row.cumulativeExp)}</td>
                <td class="text-right font-mono">{formatCurrency(row.investmentReturns)}</td>
                <td class="text-right font-mono">{formatCurrency(row.investmentVal)}</td>
                <td class="text-right font-mono">{formatCurrency(row.netPosition)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="help-text mt-2">
        <p>Note: This table tracks only cash flow (expenses and income) from the asset, not asset value or equity.</p>
        <div class="grid grid-cols-[auto_1fr] gap-x-2">
          {#if inputs.mortgageInterestDeduction > 0}
            <span class="text-light-cyan dark:text-monokai-cyan">Eff. Loan Pmt ③</span><span>= Annual effective loan payment after tax deduction (from <a href="#loan-amortization" class="underline" on:click={(e) => scrollToSection(e, 'loan-amortization')}>Loan Amortization</a>).</span>
          {:else}
            <span class="text-light-cyan dark:text-monokai-cyan">Loan Payment</span><span>= Loan payments for that year (stops after loan term).</span>
          {/if}
          <span class="text-light-cyan dark:text-monokai-cyan">Costs</span><span>= Tax, insurance, and other annual costs (inflated at {formatPercent(inputs.inflationRate)} annually). Negative means income from asset exceeds costs.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Cumulative Exp</span><span>= Running total of raw expenses.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Invest Returns</span><span>= Cumulative investment returns (at {formatPercent(inputs.investmentReturnRate)} annual return).</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Investment Val</span><span>= Current value of investments (initial cash + returns - withdrawals).</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Net Position ②</span><span>= Investment value minus real out-of-pocket costs.</span>
        </div>
      </div>
    </section>
  {/if}

  <!-- KEEP Sale Proceeds (for sell_vs_keep) or Sale Proceeds (for buy_vs_rent) -->
  <section id="keep-sale-proceeds" class="bg-light-bg-light dark:bg-monokai-bg-light p-6 rounded-lg">
    <h2 class="section-title">{inputs.scenario === 'sell_vs_keep' ? 'KEEP Analysis: Future Sale Proceeds' : 'BUY Analysis: Future Asset Value'}</h2>
    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>Period</th>
            <th class="text-right">{showSellingColumns ? 'Sale Price' : 'Asset Value'}</th>
            {#if showSellingColumns}
              <th class="text-right">Selling Cost</th>
            {/if}
            <th class="text-right"><a href="#loan-amortization" class="hover:underline cursor-pointer" on:click={(e) => scrollToSection(e, 'loan-amortization')}>Loan Payoff</a></th>
            {#if showSellingColumns}
              <th class="text-right">Cap Gains</th>
              <th class="text-right">Tax</th>
            {/if}
            <th class="text-right">{showSellingColumns ? 'Net Proceeds ①' : 'Net Equity'}</th>
          </tr>
        </thead>
        <tbody>
          {#each results.saleProceedsTable as row}
            <tr>
              <td class="font-mono">{row.period}</td>
              <td class="text-right font-mono">{formatCurrency(row.salePrice)}</td>
              {#if showSellingColumns}
                <td class="text-right font-mono">{formatCurrency(row.totalSellingCosts)}</td>
              {/if}
              <td class="text-right font-mono">{formatCurrency(row.loanPayoff)}</td>
              {#if showSellingColumns}
                <td class="text-right font-mono">{formatCurrency(row.capitalGains)}</td>
                <td class="text-right font-mono">{formatCurrency(row.taxOnGains)}</td>
              {/if}
              <td class="text-right font-mono">{formatCurrency(row.netProceeds)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
    <div class="help-text mt-2">
      {#if inputs.scenario === 'sell_vs_keep'}
        <p>Note: Shows net proceeds if you KEEP the asset and sell at each future point. This Net Proceeds value feeds into KEEP Net Worth in the comparison table below.</p>
      {:else if showSellingColumns}
        <p>Note: Appreciation rates are applied year-by-year (compounded).</p>
      {:else}
        <p>Note: Shows asset value and equity without selling costs. Enable "Include Selling Analysis" to see sale proceeds with costs and taxes.</p>
      {/if}
      <div class="grid grid-cols-[auto_1fr] gap-x-2">
        <span class="text-light-cyan dark:text-monokai-cyan">{showSellingColumns ? 'Sale Price' : 'Asset Value'}</span><span>= Compounded property value.</span>
        {#if showSellingColumns}
          <span class="text-light-cyan dark:text-monokai-cyan">Selling Cost</span><span>= Agent commission + staging costs.</span>
        {/if}
        <span class="text-light-cyan dark:text-monokai-cyan">Loan Payoff</span><span>= Remaining loan balance at that time.</span>
        {#if showSellingColumns}
          <span class="text-light-cyan dark:text-monokai-cyan">Cap Gains</span><span>= Sale price - purchase price - selling costs.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Tax</span><span>= Tax on gains exceeding tax-free limit.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Net Proceeds ①</span><span>= Sale price - selling costs - loan payoff - tax.</span>
        {:else}
          <span class="text-light-cyan dark:text-monokai-cyan">Net Equity</span><span>= Asset value - loan payoff.</span>
        {/if}
      </div>
    </div>
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
              <th class="text-right"><a href="#keep-sale-proceeds" class="text-light-pink dark:text-monokai-pink hover:underline cursor-pointer" on:click={(e) => scrollToSection(e, 'keep-sale-proceeds')}>Buying NW ①</a></th>
              <th class="text-right"><a href="#expenditure-comparison" class="text-light-text dark:text-monokai-text hover:underline cursor-pointer" on:click={(e) => scrollToSection(e, 'expenditure-comparison')}>Cum Savings ②</a></th>
              <th class="text-right">Market Return</th>
              <th class="text-right"><span class="text-light-green dark:text-monokai-green">Renting NW</span></th>
              <th class="text-right"><span class="text-light-pink dark:text-monokai-pink">BUY</span> - <span class="text-light-green dark:text-monokai-green">RENT</span></th>
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
                <td class="text-right font-mono" class:text-light-pink={-row.difference > 0} class:dark:text-monokai-pink={-row.difference > 0} class:text-light-green={-row.difference < 0} class:dark:text-monokai-green={-row.difference < 0}>
                  {formatCurrency(-row.difference)}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="help-text mt-2">
        <p>Note: Positive BUY - RENT means buying wins, negative means renting wins.</p>
        <div class="grid grid-cols-[auto_1fr] gap-x-2">
          <span class="text-light-cyan dark:text-monokai-cyan">Asset Value</span><span>= Property value after appreciation.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Buying NW ①</span><span>= Asset value minus remaining loan (or net sale proceeds if selling, from BUY Analysis: Future Asset Value).</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Cum Savings ②</span><span>= Raw cost difference without investment growth (from Total Expenditure Comparison).</span>
          <span class="text-light-cyan dark:text-monokai-cyan">Market Return</span><span>= Investment growth at {formatPercent(inputs.investmentReturnRate)} annual rate.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">RENT Net Worth</span><span>= Cumulative savings + market return.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">BUY - RENT</span><span>= Difference in net worth (positive = buying wins).</span>
        </div>
      </div>
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
              <th class="text-right"><span class="text-light-pink dark:text-monokai-pink">SELL Net Worth</span></th>
              <th class="text-right"><a href="#keep-sale-proceeds" class="text-light-text dark:text-monokai-text hover:underline cursor-pointer" on:click={(e) => scrollToSection(e, 'keep-sale-proceeds')}>KEEP Sale Proceeds ①</a></th>
              <th class="text-right"><a href="#keep-invest-position" class="text-light-text dark:text-monokai-text hover:underline cursor-pointer" on:click={(e) => scrollToSection(e, 'keep-invest-position')}>KEEP Invest Position ②</a></th>
              <th class="text-right"><span class="text-light-green dark:text-monokai-green">KEEP Net Worth</span></th>
              <th class="text-right"><span class="text-light-pink dark:text-monokai-pink">SELL</span> - <span class="text-light-green dark:text-monokai-green">KEEP</span></th>
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
                <td class="text-right font-mono">{formatCurrency(row.keepSaleProceeds)}</td>
                <td class="text-right font-mono">{formatCurrency(row.keepNetPosition)}</td>
                <td class="text-right font-mono text-light-green dark:text-monokai-green">{formatCurrency(row.keepNetWorth)}</td>
                <td class="text-right font-mono" class:text-light-pink={-row.difference > 0} class:dark:text-monokai-pink={-row.difference > 0} class:text-light-green={-row.difference < 0} class:dark:text-monokai-green={-row.difference < 0}>
                  {formatCurrency(-row.difference)}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="help-text mt-2">
        <p>Note: Positive SELL - KEEP means selling wins, negative means keeping wins.</p>
        <div class="grid grid-cols-[auto_1fr] gap-x-2">
          <span class="text-light-cyan dark:text-monokai-cyan">SELL Net Worth</span><span>= Net proceeds from selling now invested at {formatPercent(inputs.investmentReturnRate)} return.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">KEEP Sale Proceeds ①</span><span>= Net proceeds if selling at that future point (from KEEP Analysis: Future Sale Proceeds).</span>
          <span class="text-light-cyan dark:text-monokai-cyan">KEEP Invest Position ②</span><span>= Investment value from income minus real out-of-pocket costs (from KEEP Analysis: Invest Position).</span>
          <span class="text-light-cyan dark:text-monokai-cyan">KEEP Net Worth</span><span>= KEEP Sale Proceeds + KEEP Invest Position.</span>
          <span class="text-light-cyan dark:text-monokai-cyan">SELL - KEEP</span><span>= Difference in net worth (positive = selling wins, negative = keeping wins).</span>
        </div>
      </div>
    </section>
  {/if}
</div>

<div class="mt-8 p-4 bg-light-bg-light dark:bg-monokai-bg-light rounded-lg text-sm text-light-text-muted dark:text-monokai-text-muted">
  <strong>Questions?</strong> Use the "Copy for LLM" button above to copy results, then paste into Claude or Gemini for deeper analysis. (As of Dec '25, ChatGPT tends to get confused with this data.)
</div>

