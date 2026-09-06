import AsyncStorage from '@react-native-async-storage/async-storage';

import type { Product } from './types';

const PRODUCTS_CACHE_KEY = 'wk_products_cache';

export interface ProductsCache {
  products: Product[];
  cachedAt: string;
}

/** Phase 1 of offline support: a read-only cache of the last successful
 * product list, so the dashboard can still show something useful without a
 * connection. Does not attempt to queue writes made while offline. */
export async function saveProductsCache(products: Product[]): Promise<void> {
  const payload: ProductsCache = { products, cachedAt: new Date().toISOString() };
  await AsyncStorage.setItem(PRODUCTS_CACHE_KEY, JSON.stringify(payload));
}

export async function loadProductsCache(): Promise<ProductsCache | null> {
  const raw = await AsyncStorage.getItem(PRODUCTS_CACHE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as ProductsCache;
  } catch {
    return null;
  }
}
