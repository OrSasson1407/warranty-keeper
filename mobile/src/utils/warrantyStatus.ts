import type { WarrantyStatus } from '../theme/colors';

const DAY_MS = 1000 * 60 * 60 * 24;
const WARNING_THRESHOLD_DAYS = 30;

export function daysUntil(dateStr: string): number {
  const target = new Date(dateStr);
  const now = new Date();
  target.setHours(0, 0, 0, 0);
  now.setHours(0, 0, 0, 0);
  return Math.round((target.getTime() - now.getTime()) / DAY_MS);
}

export function warrantyStatus(expiresAt: string): WarrantyStatus {
  const days = daysUntil(expiresAt);
  if (days < 0) return 'expired';
  if (days <= WARNING_THRESHOLD_DAYS) return 'warning';
  return 'ok';
}

export function formatHebrewDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString('he-IL', { year: 'numeric', month: 'long', day: 'numeric' });
}

export function statusLabel(expiresAt: string): string {
  const days = daysUntil(expiresAt);
  if (days < 0) return 'האחריות פגה';
  if (days === 0) return 'האחריות פגה היום';
  if (days <= WARNING_THRESHOLD_DAYS) return `פג בעוד ${days} ימים`;
  return `באחריות עד ${formatHebrewDate(expiresAt)}`;
}
