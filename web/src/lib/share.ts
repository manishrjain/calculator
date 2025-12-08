// Short keys to minimize URL length
const KEY_MAP: Record<string, string> = {
  scenario: 's',
  inflationRate: 'ir',
  investmentReturnRate: 'ivr',
  include30Year: 'i30',
  purchasePrice: 'pp',
  currentMarketValue: 'cmv',
  annualInsurance: 'ai',
  annualTaxes: 'at',
  annualIncome: 'ain',
  appreciationRate: 'ar',
  includeRefinance: 'irf',
  payoffBalance: 'pb',
  loanAmount: 'la',
  loanRate: 'lr',
  loanTerm: 'lt',
  remainingLoanTerm: 'rlt',
  closingCosts: 'cc',
  mortgageInterestDeduction: 'mid',
  rentDeposit: 'rd',
  monthlyRent: 'mr',
  annualRentCosts: 'arc',
  otherAnnualCosts: 'oac',
  includeSelling: 'is',
  includeRentingSell: 'irs',
  agentCommission: 'ac',
  stagingCosts: 'sc',
  taxFreeLimits: 'tfl',
  capitalGainsTax: 'cgt',
};

// Reverse map for decoding
const REVERSE_KEY_MAP: Record<string, string> = Object.fromEntries(
  Object.entries(KEY_MAP).map(([k, v]) => [v, k])
);

export function encodeInputsToURL(inputs: Record<string, any>): string {
  const params = new URLSearchParams();

  for (const [key, value] of Object.entries(inputs)) {
    if (value !== undefined && value !== null && value !== '') {
      const shortKey = KEY_MAP[key] || key;
      params.set(shortKey, String(value));
    }
  }

  const baseURL = window.location.origin + window.location.pathname;
  return `${baseURL}?${params.toString()}`;
}

export function decodeInputsFromURL(): Record<string, any> | null {
  const params = new URLSearchParams(window.location.search);

  if (params.toString() === '') {
    return null;
  }

  const inputs: Record<string, any> = {};

  for (const [shortKey, value] of params.entries()) {
    const fullKey = REVERSE_KEY_MAP[shortKey] || shortKey;
    inputs[fullKey] = value;
  }

  return Object.keys(inputs).length > 0 ? inputs : null;
}

export function clearURLParams(): void {
  const baseURL = window.location.origin + window.location.pathname;
  window.history.replaceState({}, '', baseURL);
}

export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}
