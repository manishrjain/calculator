import type { CalculatorInputs, CalculationResults, Period } from '../types';

export function calculateMonthlyPayment(
  principal: number,
  monthlyRate: number,
  months: number
): number {
  if (monthlyRate === 0) {
    return principal / months;
  }
  const factor = Math.pow(1 + monthlyRate, months);
  return (principal * (monthlyRate * factor)) / (factor - 1);
}

export function getPeriods(loanDuration: number, include30Year: boolean): Period[] {
  const basePeriods: Period[] = [
    { label: '  1y', months: 12 },
    { label: '  2y', months: 24 },
    { label: '  3y', months: 36 },
    { label: '  4y', months: 48 },
    { label: '  5y', months: 60 },
    { label: '  6y', months: 72 },
    { label: '  7y', months: 84 },
    { label: '  8y', months: 96 },
    { label: '  9y', months: 108 },
    { label: ' 10y', months: 120 },
  ];

  const extendedPeriods: Period[] = [
    { label: ' 15y', months: 180 },
    { label: ' 20y', months: 240 },
    { label: ' 30y', months: 360 },
  ];

  let standardPeriods = basePeriods;
  const maxMonths = include30Year ? 360 : 120;

  if (include30Year) {
    standardPeriods = [...basePeriods, ...extendedPeriods];
  }

  const periods: Period[] = [];
  let loanTermLabel = '';
  let includeLoanTerm = false;

  // Only include loan term if it's within the allowed range
  if (loanDuration > 0 && loanDuration % 12 === 0 && loanDuration <= maxMonths) {
    const years = loanDuration / 12;
    loanTermLabel = `X ${years}y`;
    includeLoanTerm = true;
  }

  let inserted = false;
  for (const period of standardPeriods) {
    if (includeLoanTerm && !inserted && loanDuration < period.months) {
      periods.push({ label: loanTermLabel, months: loanDuration });
      inserted = true;
    }

    if (period.months === loanDuration && includeLoanTerm) {
      periods.push({ label: loanTermLabel, months: loanDuration });
      inserted = true;
    } else {
      periods.push(period);
    }
  }

  if (includeLoanTerm && !inserted) {
    periods.push({ label: loanTermLabel, months: loanDuration });
  }

  return periods;
}

export function populateMonthlyCosts(inputs: CalculatorInputs): {
  monthlyBuyingCosts: number[];
  monthlyRentingCosts: number[];
  remainingLoanBalance: number[];
  cumulativePrincipalPaid: number[];
  cumulativeInterestPaid: number[];
} {
  const maxMonths = 360;
  const monthlyBuyingCosts: number[] = new Array(maxMonths);
  const monthlyRentingCosts: number[] = new Array(maxMonths);
  const remainingLoanBalance: number[] = new Array(maxMonths);
  const cumulativePrincipalPaid: number[] = new Array(maxMonths);
  const cumulativeInterestPaid: number[] = new Array(maxMonths);

  const totalAnnualExpenses = inputs.annualInsurance + inputs.annualTaxes;
  let monthlyRecurringExpenses = totalAnnualExpenses / 12 + inputs.monthlyExpenses;

  const monthlyRate = inputs.loanRate / 100 / 12;
  const monthlyLoanPayment =
    inputs.loanAmount > 0
      ? calculateMonthlyPayment(inputs.loanAmount, monthlyRate, inputs.loanTerm)
      : 0;

  let currentRentingCost =
    inputs.monthlyRent + inputs.annualRentCosts / 12 + inputs.otherAnnualCosts / 12;
  let currentRecurringExpenses = monthlyRecurringExpenses;

  let currentBalance = inputs.loanAmount;
  let totalPrincipalPaid = 0;
  let totalInterestPaid = 0;

  for (let i = 0; i < maxMonths; i++) {
    if (i > 0 && i % 12 === 0) {
      currentRentingCost *= 1 + inputs.inflationRate / 100;
      currentRecurringExpenses *= 1 + inputs.inflationRate / 100;
    }

    monthlyRentingCosts[i] = currentRentingCost;

    if (i < inputs.loanTerm) {
      monthlyBuyingCosts[i] = monthlyLoanPayment + currentRecurringExpenses;

      const interestPayment = currentBalance * monthlyRate;
      const principalPayment = monthlyLoanPayment - interestPayment;
      currentBalance -= principalPayment;

      totalPrincipalPaid += principalPayment;
      totalInterestPaid += interestPayment;

      remainingLoanBalance[i] = currentBalance;
      cumulativePrincipalPaid[i] = totalPrincipalPaid;
      cumulativeInterestPaid[i] = totalInterestPaid;
    } else {
      monthlyBuyingCosts[i] = currentRecurringExpenses;
      remainingLoanBalance[i] = 0;
      cumulativePrincipalPaid[i] = totalPrincipalPaid;
      cumulativeInterestPaid[i] = totalInterestPaid;
    }
  }

  return {
    monthlyBuyingCosts,
    monthlyRentingCosts,
    remainingLoanBalance,
    cumulativePrincipalPaid,
    cumulativeInterestPaid,
  };
}

export function calculateKeepInvestmentTracking(
  monthlyBuyingCosts: number[],
  investmentReturnRate: number,
  maxMonths: number
): {
  monthlyKeepInvestmentValue: number[];
  monthlyKeepRealCosts: number[];
  monthlyKeepNetPosition: number[];
} {
  const monthlyKeepInvestmentValue: number[] = new Array(maxMonths);
  const monthlyKeepRealCosts: number[] = new Array(maxMonths);
  const monthlyKeepNetPosition: number[] = new Array(maxMonths);

  let investmentValue = 0;
  let totalRealCosts = 0;
  const monthlyInvestmentRate = investmentReturnRate / 100 / 12;

  for (let i = 0; i < maxMonths; i++) {
    const monthlyCost = monthlyBuyingCosts[i];

    if (monthlyCost < 0) {
      investmentValue += -monthlyCost;
    } else if (monthlyCost > 0) {
      if (investmentValue >= monthlyCost) {
        investmentValue -= monthlyCost;
      } else {
        const deficit = monthlyCost - investmentValue;
        investmentValue = 0;
        totalRealCosts += deficit;
      }
    }

    investmentValue *= 1 + monthlyInvestmentRate;

    monthlyKeepInvestmentValue[i] = investmentValue;
    monthlyKeepRealCosts[i] = totalRealCosts;
    monthlyKeepNetPosition[i] = investmentValue - totalRealCosts;
  }

  return {
    monthlyKeepInvestmentValue,
    monthlyKeepRealCosts,
    monthlyKeepNetPosition,
  };
}

export function calculateAssetValue(
  startingPrice: number,
  months: number,
  appreciationRates: number[]
): number {
  let assetValue = startingPrice;
  const years = Math.floor(months / 12);
  const remainingMonths = months % 12;

  for (let year = 0; year < years; year++) {
    const rateIndex = Math.min(year, appreciationRates.length - 1);
    assetValue *= 1 + appreciationRates[rateIndex] / 100;
  }

  if (remainingMonths > 0) {
    const rateIndex = Math.min(years, appreciationRates.length - 1);
    const partialYearFactor = Math.pow(
      1 + appreciationRates[rateIndex] / 100,
      remainingMonths / 12
    );
    assetValue *= partialYearFactor;
  }

  return assetValue;
}

export function calculateSaleProceeds(
  inputs: CalculatorInputs,
  months: number,
  remainingLoanBalance: number[]
): {
  salePrice: number;
  totalSellingCosts: number;
  loanPayoff: number;
  capitalGains: number;
  taxOnGains: number;
  netProceeds: number;
} {
  const startingPrice = inputs.currentMarketValue || inputs.purchasePrice;
  const salePrice = calculateAssetValue(startingPrice, months, inputs.appreciationRate);

  const agentFee = salePrice * (inputs.agentCommission / 100);
  const totalSellingCosts = agentFee + inputs.stagingCosts;

  const monthIndex = Math.min(months - 1, remainingLoanBalance.length - 1);
  const loanPayoff = remainingLoanBalance[monthIndex];

  const capitalGains = salePrice - inputs.purchasePrice - totalSellingCosts;

  const years = Math.floor(months / 12);
  const taxFreeLimitIndex = Math.max(0, Math.min(years - 1, inputs.taxFreeLimits.length - 1));
  const taxFreeLimit = inputs.taxFreeLimits[taxFreeLimitIndex] || 0;

  const taxableGains = Math.max(0, capitalGains - taxFreeLimit);
  const taxOnGains = taxableGains * (inputs.capitalGainsTax / 100);

  const netProceeds = salePrice - totalSellingCosts - loanPayoff - taxOnGains;

  return {
    salePrice,
    totalSellingCosts,
    loanPayoff,
    capitalGains,
    taxOnGains,
    netProceeds,
  };
}

export function calculateRentingNetWorth(
  inputs: CalculatorInputs,
  months: number,
  monthlyBuyingCosts: number[],
  monthlyRentingCosts: number[]
): number {
  const downpayment = inputs.purchasePrice - inputs.loanAmount;
  let investmentValue = downpayment - inputs.rentDeposit;
  const monthlyInvestmentRate = inputs.investmentReturnRate / 100 / 12;

  for (let i = 0; i < months; i++) {
    const monthlySavings = monthlyBuyingCosts[i] - monthlyRentingCosts[i];
    investmentValue += monthlySavings;
    investmentValue *= 1 + monthlyInvestmentRate;
  }

  const recoverableDeposit = inputs.rentDeposit * 0.75;
  return investmentValue + recoverableDeposit;
}

export function calculate(inputs: CalculatorInputs): CalculationResults {
  const costs = populateMonthlyCosts(inputs);
  const periods = getPeriods(inputs.loanTerm, inputs.include30Year);

  const keepTracking = calculateKeepInvestmentTracking(
    costs.monthlyBuyingCosts,
    inputs.investmentReturnRate,
    360
  );

  // Calculate sale proceeds for all periods
  const saleProceedsTable = periods.map((period) => {
    const proceeds = calculateSaleProceeds(inputs, period.months, costs.remainingLoanBalance);
    return {
      period: 'SALE ' + period.label,
      ...proceeds,
    };
  });

  const results: CalculationResults = {
    ...costs,
    ...keepTracking,
    periods,
    saleProceedsTable,
  };

  // Scenario-specific calculations
  if (inputs.scenario === 'buy_vs_rent') {
    // Amortization table
    if (inputs.loanAmount > 0) {
      results.amortizationTable = periods.map((period) => {
        const monthIndex = Math.min(period.months - 1, costs.remainingLoanBalance.length - 1);
        return {
          period: 'LOAN ' + period.label,
          principalPaid: costs.cumulativePrincipalPaid[monthIndex],
          interestPaid: costs.cumulativeInterestPaid[monthIndex],
          loanBalance: costs.remainingLoanBalance[monthIndex],
        };
      });
    }

    // Expenditure table
    results.expenditureTable = periods.map((period) => {
      const downpayment = inputs.purchasePrice - inputs.loanAmount;
      let buyingExpenditure = downpayment;
      let rentingExpenditure = inputs.rentDeposit;

      for (let i = 0; i < period.months; i++) {
        buyingExpenditure += costs.monthlyBuyingCosts[i];
        rentingExpenditure += costs.monthlyRentingCosts[i];
      }

      return {
        period: 'EXP ' + period.label,
        buyingExpenditure,
        rentingExpenditure,
        difference: buyingExpenditure - rentingExpenditure,
      };
    });

    // Comparison table
    results.comparisonTable = periods.map((period) => {
      const assetValue = calculateAssetValue(
        inputs.purchasePrice,
        period.months,
        inputs.appreciationRate
      );

      const downpayment = inputs.purchasePrice - inputs.loanAmount;
      const monthIndex = Math.min(period.months - 1, costs.remainingLoanBalance.length - 1);
      const loanBalance = costs.remainingLoanBalance[monthIndex];

      let buyingNetWorth: number;
      if (inputs.includeSelling) {
        const { netProceeds } = calculateSaleProceeds(
          inputs,
          period.months,
          costs.remainingLoanBalance
        );
        buyingNetWorth = netProceeds;
      } else {
        buyingNetWorth = assetValue - loanBalance;
      }

      const rentingNetWorth = calculateRentingNetWorth(
        inputs,
        period.months,
        costs.monthlyBuyingCosts,
        costs.monthlyRentingCosts
      );

      let cumulativeSavings = downpayment - inputs.rentDeposit;
      for (let i = 0; i < period.months; i++) {
        cumulativeSavings += costs.monthlyBuyingCosts[i] - costs.monthlyRentingCosts[i];
      }

      const recoverableDeposit = inputs.rentDeposit * 0.75;
      const marketReturn = rentingNetWorth - cumulativeSavings - recoverableDeposit;

      return {
        period: 'NET ' + period.label,
        assetValue,
        buyingNetWorth,
        cumulativeSavings,
        marketReturn,
        rentingNetWorth,
        difference: rentingNetWorth - buyingNetWorth,
      };
    });
  } else {
    // sell_vs_keep scenario
    results.sellVsKeepTable = periods.map((period) => {
      const monthIndex = Math.min(period.months - 1, keepTracking.monthlyKeepNetPosition.length - 1);
      const keepNetPosition = keepTracking.monthlyKeepNetPosition[monthIndex];

      const { netProceeds } = calculateSaleProceeds(
        inputs,
        period.months,
        costs.remainingLoanBalance
      );
      const keepNetWorth = netProceeds + keepNetPosition;

      // Calculate sell net worth (selling now and investing)
      const currentMarketValue = inputs.currentMarketValue || inputs.purchasePrice;
      const salePrice = currentMarketValue;
      const agentFee = salePrice * (inputs.agentCommission / 100);
      const totalSellingCosts = agentFee + inputs.stagingCosts;
      const loanPayoff = inputs.loanAmount;
      const capitalGains = salePrice - inputs.purchasePrice - totalSellingCosts;
      const taxFreeLimit = inputs.taxFreeLimits[0] || 0;
      const taxableGains = Math.max(0, capitalGains - taxFreeLimit);
      const taxOnGains = taxableGains * (inputs.capitalGainsTax / 100);
      const netProceedsNow = salePrice - totalSellingCosts - loanPayoff - taxOnGains;

      let sellNetWorth: number;
      let sellCumulativeExpenses: number | undefined;

      if (inputs.includeRentingSell) {
        let investmentValue = netProceedsNow - inputs.rentDeposit;
        const monthlyInvestmentRate = inputs.investmentReturnRate / 100 / 12;

        for (let i = 0; i < period.months; i++) {
          investmentValue -= costs.monthlyRentingCosts[i];
          investmentValue *= 1 + monthlyInvestmentRate;
        }

        const recoverableDeposit = inputs.rentDeposit * 0.75;
        sellNetWorth = investmentValue + recoverableDeposit;

        // Calculate cumulative rental expenses
        let cumulativeRentExpenses = inputs.rentDeposit;
        for (let i = 0; i < period.months; i++) {
          cumulativeRentExpenses += costs.monthlyRentingCosts[i];
        }
        cumulativeRentExpenses -= recoverableDeposit;
        sellCumulativeExpenses = cumulativeRentExpenses;
      } else {
        let investmentValue = netProceedsNow;
        const monthlyInvestmentRate = inputs.investmentReturnRate / 100 / 12;

        for (let i = 0; i < period.months; i++) {
          investmentValue *= 1 + monthlyInvestmentRate;
        }

        sellNetWorth = investmentValue;
      }

      return {
        period: 'NET ' + period.label,
        sellCumulativeExpenses,
        sellNetWorth,
        keepNetPosition,
        keepNetProceeds: keepNetWorth,
        difference: keepNetWorth - sellNetWorth,
      };
    });
  }

  return results;
}
