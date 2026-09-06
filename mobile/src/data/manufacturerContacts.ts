import { api } from '../api/client';
import { loadManufacturerContactsCache, saveManufacturerContactsCache } from '../api/offlineCache';
import type { ManufacturerContact } from '../api/types';

export type ManufacturerContactMap = Record<string, ManufacturerContact>;

/** Fetches the server-managed manufacturer contact list (see the "server-side
 * manufacturer contact database" backlog item) and caches it for offline use,
 * falling back to the cache if the fetch fails. Replaces what used to be a
 * static bundled file — brands not in the list still fall back to
 * ClaimScreen's generic "contact your point of purchase" message. */
export async function loadManufacturerContacts(): Promise<ManufacturerContactMap> {
  try {
    const contacts = await api.listManufacturerContacts();
    saveManufacturerContactsCache(contacts);
    return toMap(contacts);
  } catch {
    const cached = await loadManufacturerContactsCache();
    return cached ? toMap(cached) : {};
  }
}

export function getManufacturerContact(
  contacts: ManufacturerContactMap,
  brand: string,
): ManufacturerContact | null {
  return contacts[brand] ?? null;
}

function toMap(contacts: ManufacturerContact[]): ManufacturerContactMap {
  const map: ManufacturerContactMap = {};
  for (const c of contacts) {
    map[c.brand] = c;
  }
  return map;
}
