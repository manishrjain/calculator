export type ScenarioType = 'buy_vs_rent' | 'sell_vs_keep';

export interface CalculatorInputs {
  // Scenario selection
  scenario: ScenarioType;

  // Economic assumptions
  inflationRate: number;
  investmentReturnRate: number;
  include30Year: boolean;

  // Buying/Asset
  purchasePrice: number;
  currentMarketValue?: number; // Only for sell_vs_keep
  annualInsurance: number;
  annualTaxes: number;
  annualIncome: number; // Annual income from asset (e.g., rental income)
  appreciationRate: number[]; // Array for different years

  // Loan
  loanAmount: number;
  loanRate: number;
  loanTerm: number; // in months
  remainingLoanTerm?: number; // Only for sell_vs_keep, in months
  includeRefinance?: boolean; // Only for sell_vs_keep
  payoffBalance?: number; // Only for sell_vs_keep with refinance
  closingCosts?: number; // Only for sell_vs_keep with refinance
  mortgageInterestDeduction: number; // Effective tax rate for mortgage interest deduction (0 to skip)

  // Renting
  rentDeposit: number;
  monthlyRent: number;
  annualRentCosts: number;
  otherAnnualCosts: number;

  // Selling
  includeSelling: boolean;
  includeRentingSell?: boolean; // Only for sell_vs_keep
  agentCommission: number;
  stagingCosts: number;
  taxFreeLimits: number[]; // Array for different years
  capitalGainsTax: number;
}

export interface Period {
  label: string;
  months: number;
}

export interface AmortizationRow {
  period: string;
  principalPaid: number;
  interestPaid: number;
  taxDeduction: number; // Tax savings from mortgage interest deduction
  effectiveInterest: number; // Interest paid after tax deduction
  effectiveLoanPayment: number; // Principal + Effective Interest (what you actually pay)
  loanBalance: number;
}

export interface ExpenditureRow {
  period: string;
  loanPayment: number;
  taxDeduction: number; // Tax savings from mortgage interest deduction
  effectiveLoanPayment: number; // Loan payment after tax deduction
  costs: number;
  buyingExpenditure: number;
  rentingExpenditure: number;
  difference: number;
}

export interface SaleProceedsRow {
  period: string;
  salePrice: number;
  totalSellingCosts: number;
  loanPayoff: number;
  capitalGains: number;
  taxOnGains: number;
  netProceeds: number;
}

export interface ComparisonRow {
  period: string;
  assetValue: number;
  buyingNetWorth: number;
  cumulativeSavings: number;
  marketReturn: number;
  rentingNetWorth: number;
  difference: number;
}

export interface SellVsKeepRow {
  period: string;
  sellCumulativeExpenses?: number;
  sellNetWorth: number;
  keepSaleProceeds: number;
  keepNetPosition: number;
  keepNetWorth: number;
  difference: number;
}

export interface KeepExpensesRow {
  period: string;
  loanPayment: number;
  taxDeduction: number; // Tax savings from mortgage interest deduction
  effectiveLoanPayment: number; // Loan payment after tax deduction
  taxInsurance: number;
  otherCosts: number;
  cumulativeExp: number;
  investmentVal: number;
  investmentReturns: number;
  netPosition: number;
}

export interface CalculationResults {
  // Common arrays for calculations
  monthlyBuyingCosts: number[];
  monthlyRentingCosts: number[];
  remainingLoanBalance: number[];
  cumulativePrincipalPaid: number[];
  cumulativeInterestPaid: number[];

  // For KEEP scenario
  monthlyKeepInvestmentValue: number[];
  monthlyKeepRealCosts: number[];
  monthlyKeepNetPosition: number[];
  monthlyKeepInvestmentReturns: number[];

  // Table data
  periods: Period[];
  amortizationTable?: AmortizationRow[];
  expenditureTable?: ExpenditureRow[];
  saleProceedsTable: SaleProceedsRow[];
  comparisonTable?: ComparisonRow[];
  sellVsKeepTable?: SellVsKeepRow[];
  keepExpensesTable?: KeepExpensesRow[];
}
