import AsyncStorage from '@react-native-async-storage/async-storage';

import type { ManufacturerContact, Product } from './types';

const PRODUCTS_CACHE_KEY = 'wk_products_cache';
const MANUFACTURER_CONTACTS_CACHE_KEY = 'wk_manufacturer_contacts_cache';

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

/** Same read-through-cache pattern for the manufacturer contact list, which
 * used to be a static bundled file (see internal/models/manufacturer_contact.go
 * on the server side) and is now fetched from the API. */
export async function saveManufacturerContactsCache(
  contacts: ManufacturerContact[],
): Promise<void> {
  await AsyncStorage.setItem(MANUFACTURER_CONTACTS_CACHE_KEY, JSON.stringify(contacts));
}

export async function loadManufacturerContactsCache(): Promise<ManufacturerContact[] | null> {
  const raw = await AsyncStorage.getItem(MANUFACTURER_CONTACTS_CACHE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as ManufacturerContact[];
  } catch {
    return null;
  }
}
