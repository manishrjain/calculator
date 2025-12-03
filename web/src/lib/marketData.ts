import marketDataJson from '../data/market_data.json';

export interface MarketData {
  last_updated: string;
  voo: Record<string, number>;
  qqq: Record<string, number>;
  vti: Record<string, number>;
  bnd: Record<string, number>;
}

export interface MarketYearData {
  year: string;
  voo: number;
  qqq: number;
  vti: number;
  bnd: number;
  mix6040: number;
}

export interface MarketAverages {
  voo: number;
  qqq: number;
  vti: number;
  bnd: number;
  mix6040: number;
}

export const marketData: MarketData = marketDataJson;

export function getMarketYears(): MarketYearData[] {
  const years = Object.keys(marketData.voo).sort();

  return years.map(year => {
    const voo = marketData.voo[year] || 0;
    const qqq = marketData.qqq[year] || 0;
    const vti = marketData.vti[year] || 0;
    const bnd = marketData.bnd[year] || 0;
    const mix6040 = vti * 0.6 + bnd * 0.4;

    return { year, voo, qqq, vti, bnd, mix6040 };
  });
}

export function getMarketAverages(): MarketAverages {
  const currentYear = new Date().getFullYear();
  const years = Object.keys(marketData.voo);

  let vooSum = 0, qqqSum = 0, vtiSum = 0, bndSum = 0;
  let count = 0;

  for (const year of years) {
    const yearInt = parseInt(year, 10);
    // Only include complete years (not current year)
    if (yearInt < currentYear) {
      const hasAll =
        marketData.voo[year] !== undefined &&
        marketData.qqq[year] !== undefined &&
        marketData.vti[year] !== undefined &&
        marketData.bnd[year] !== undefined;

      if (hasAll) {
        vooSum += marketData.voo[year];
        qqqSum += marketData.qqq[year];
        vtiSum += marketData.vti[year];
        bndSum += marketData.bnd[year];
        count++;
      }
    }
  }

  if (count === 0) {
    return { voo: 0, qqq: 0, vti: 0, bnd: 0, mix6040: 0 };
  }

  const voo = vooSum / count;
  const qqq = qqqSum / count;
  const vti = vtiSum / count;
  const bnd = bndSum / count;
  const mix6040 = vti * 0.6 + bnd * 0.4;

  return { voo, qqq, vti, bnd, mix6040 };
}

export function getLastUpdated(): string {
  return marketData.last_updated;
}
