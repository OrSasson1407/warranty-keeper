import { computeAnalytics } from './analytics';
import type { Product } from '../api/types';

const FAKE_NOW = new Date('2026-09-04T12:00:00Z');

beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(FAKE_NOW);
});

afterEach(() => {
  jest.useRealTimers();
});

function product(overrides: Partial<Product>): Product {
  return {
    id: 'p1',
    household_id: 'h1',
    name: 'מוצר',
    category: 'קטגוריה',
    brand: '',
    purchase_date: '2026-01-01',
    price: null,
    room: '',
    warranty_expires_at: '2028-01-01',
    warranty_uncertain: false,
    photo_url: '',
    receipt_id: null,
    created_at: '2026-01-01',
    updated_at: '2026-01-01',
    ...overrides,
  };
}

describe('computeAnalytics', () => {
  it('returns zeros for an empty product list', () => {
    expect(computeAnalytics([])).toEqual({ coveredValue: 0, expiringSoonCount: 0, byCategory: [] });
  });

  it('sums the price of products that are not expired', () => {
    const result = computeAnalytics([
      product({ id: 'p1', price: 1000, warranty_expires_at: '2028-01-01' }),
      product({ id: 'p2', price: 500, warranty_expires_at: '2026-10-01' }),
      product({ id: 'p3', price: 999, warranty_expires_at: '2020-01-01' }), // expired, excluded
    ]);
    expect(result.coveredValue).toBe(1500);
  });

  it('excludes products with a null price from the covered-value sum', () => {
    const result = computeAnalytics([product({ price: null, warranty_expires_at: '2028-01-01' })]);
    expect(result.coveredValue).toBe(0);
  });

  it('counts products expiring within the 30-day warning window', () => {
    const result = computeAnalytics([
      product({ id: 'p1', warranty_expires_at: '2026-09-10' }), // warning
      product({ id: 'p2', warranty_expires_at: '2028-01-01' }), // ok
      product({ id: 'p3', warranty_expires_at: '2020-01-01' }), // expired, not "expiring soon"
    ]);
    expect(result.expiringSoonCount).toBe(1);
  });

  it('groups products by category, sorted by count descending', () => {
    const result = computeAnalytics([
      product({ id: 'p1', category: 'מזגן' }),
      product({ id: 'p2', category: 'מזגן' }),
      product({ id: 'p3', category: 'מקרר' }),
    ]);
    expect(result.byCategory).toEqual([
      { category: 'מזגן', count: 2 },
      { category: 'מקרר', count: 1 },
    ]);
  });

  it('groups products with no category under a fallback label', () => {
    const result = computeAnalytics([product({ category: '' })]);
    expect(result.byCategory).toEqual([{ category: 'ללא קטגוריה', count: 1 }]);
  });
});
