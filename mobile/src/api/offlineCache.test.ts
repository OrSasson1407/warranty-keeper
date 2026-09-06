import AsyncStorage from '@react-native-async-storage/async-storage';

import { loadProductsCache, saveProductsCache } from './offlineCache';
import type { Product } from './types';

beforeEach(async () => {
  await AsyncStorage.clear();
});

function product(overrides: Partial<Product> = {}): Product {
  return {
    id: 'p1',
    household_id: 'h1',
    name: 'מוצר',
    category: 'x',
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

describe('offlineCache', () => {
  it('returns null when nothing has been cached yet', async () => {
    expect(await loadProductsCache()).toBeNull();
  });

  it('round-trips a saved product list', async () => {
    const products = [product({ id: 'p1' }), product({ id: 'p2' })];
    await saveProductsCache(products);

    const cache = await loadProductsCache();
    expect(cache?.products).toEqual(products);
  });

  it('stamps the cache with a cachedAt timestamp', async () => {
    const before = new Date().toISOString();
    await saveProductsCache([product()]);
    const cache = await loadProductsCache();

    expect(cache?.cachedAt).toBeDefined();
    expect(new Date(cache!.cachedAt).getTime()).toBeGreaterThanOrEqual(new Date(before).getTime());
  });

  it('returns null instead of throwing on corrupted cache data', async () => {
    await AsyncStorage.setItem('wk_products_cache', 'not valid json{');
    expect(await loadProductsCache()).toBeNull();
  });

  it('overwrites the previous cache on a new save', async () => {
    await saveProductsCache([product({ id: 'old' })]);
    await saveProductsCache([product({ id: 'new' })]);

    const cache = await loadProductsCache();
    expect(cache?.products).toEqual([product({ id: 'new' })]);
  });
});
