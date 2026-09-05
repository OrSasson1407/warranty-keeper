import { daysUntil, formatHebrewDate, statusLabel, warrantyStatus } from './warrantyStatus';

const FAKE_NOW = new Date('2026-09-04T12:00:00Z');

beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(FAKE_NOW);
});

afterEach(() => {
  jest.useRealTimers();
});

describe('daysUntil', () => {
  it('returns 0 for today', () => {
    expect(daysUntil('2026-09-04')).toBe(0);
  });

  it('returns a positive count for a future date', () => {
    expect(daysUntil('2026-09-14')).toBe(10);
  });

  it('returns a negative count for a past date', () => {
    expect(daysUntil('2026-08-25')).toBe(-10);
  });

  it('ignores time-of-day, comparing whole days only', () => {
    expect(daysUntil('2026-09-05T23:59:59')).toBe(1);
  });
});

describe('warrantyStatus', () => {
  it('is "expired" once the date has passed', () => {
    expect(warrantyStatus('2026-09-03')).toBe('expired');
  });

  it('is "warning" today (0 days left)', () => {
    expect(warrantyStatus('2026-09-04')).toBe('warning');
  });

  it('is "warning" at the 30-day boundary', () => {
    expect(warrantyStatus('2026-10-04')).toBe('warning');
  });

  it('is "ok" just past the 30-day boundary', () => {
    expect(warrantyStatus('2026-10-05')).toBe('ok');
  });

  it('is "ok" far in the future', () => {
    expect(warrantyStatus('2028-01-01')).toBe('ok');
  });
});

describe('statusLabel', () => {
  it('says the warranty expired for a past date', () => {
    expect(statusLabel('2026-09-03')).toBe('האחריות פגה');
  });

  it('has a distinct message for expiring exactly today', () => {
    expect(statusLabel('2026-09-04')).toBe('האחריות פגה היום');
  });

  it('counts down the days within the warning window', () => {
    expect(statusLabel('2026-09-16')).toBe('פג בעוד 12 ימים');
  });

  it('shows the formatted expiry date once safely in warranty', () => {
    expect(statusLabel('2028-01-01')).toBe(`באחריות עד ${formatHebrewDate('2028-01-01')}`);
  });
});

describe('formatHebrewDate', () => {
  it('formats a date using Hebrew month names', () => {
    expect(formatHebrewDate('2026-01-15')).toContain('ינואר');
    expect(formatHebrewDate('2026-01-15')).toContain('2026');
  });
});
