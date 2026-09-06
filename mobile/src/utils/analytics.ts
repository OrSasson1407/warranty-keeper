import type { Product } from '../api/types';
import { warrantyStatus } from './warrantyStatus';

export interface CategoryCount {
  category: string;
  count: number;
}

export interface DashboardAnalytics {
  coveredValue: number;
  expiringSoonCount: number;
  byCategory: CategoryCount[];
}

const UNCATEGORIZED_LABEL = 'ללא קטגוריה';

/** Pure aggregation over a household's products, computed client-side from
 * the data the dashboard already fetches -- no extra API calls needed at
 * MVP-scale product counts. */
export function computeAnalytics(products: Product[]): DashboardAnalytics {
  let coveredValue = 0;
  let expiringSoonCount = 0;
  const countByCategory = new Map<string, number>();

  for (const p of products) {
    const status = warrantyStatus(p.warranty_expires_at);
    if (status !== 'expired' && p.price) {
      coveredValue += p.price;
    }
    if (status === 'warning') {
      expiringSoonCount += 1;
    }
    const label = p.category || UNCATEGORIZED_LABEL;
    countByCategory.set(label, (countByCategory.get(label) ?? 0) + 1);
  }

  const byCategory = Array.from(countByCategory.entries())
    .map(([category, count]) => ({ category, count }))
    .sort((a, b) => b.count - a.count);

  return { coveredValue, expiringSoonCount, byCategory };
}
