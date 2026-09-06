import AsyncStorage from '@react-native-async-storage/async-storage';

import { getManufacturerContact, loadManufacturerContacts } from './manufacturerContacts';
import { api } from '../api/client';
import { saveManufacturerContactsCache } from '../api/offlineCache';
import type { ManufacturerContact } from '../api/types';

jest.mock('../api/client', () => ({ api: { listManufacturerContacts: jest.fn() } }));

const mockListManufacturerContacts = api.listManufacturerContacts as jest.Mock;

const bosch: ManufacturerContact = {
  brand: 'בוש',
  phone: '03-1234567',
  website: 'https://www.bosch-home.co.il',
};
const samsung: ManufacturerContact = {
  brand: 'Samsung',
  phone: '*6444',
  website: 'https://samsung.com',
};

beforeEach(async () => {
  await AsyncStorage.clear();
});

describe('getManufacturerContact', () => {
  const contacts = { [bosch.brand]: bosch, [samsung.brand]: samsung };

  it('returns contact info for a known brand', () => {
    expect(getManufacturerContact(contacts, 'בוש')).toEqual(bosch);
  });

  it('returns null for an unlisted brand', () => {
    expect(getManufacturerContact(contacts, 'SomeObscureBrand')).toBeNull();
  });

  it('returns null for an empty brand', () => {
    expect(getManufacturerContact(contacts, '')).toBeNull();
  });

  it('is case-sensitive (brand keys are stored verbatim)', () => {
    expect(getManufacturerContact(contacts, 'bosch')).toBeNull();
  });
});

describe('loadManufacturerContacts', () => {
  it('fetches from the API and returns a map keyed by brand', async () => {
    mockListManufacturerContacts.mockResolvedValue([bosch, samsung]);

    const contacts = await loadManufacturerContacts();

    expect(contacts).toEqual({ [bosch.brand]: bosch, [samsung.brand]: samsung });
  });

  it('caches a successful fetch for offline use', async () => {
    mockListManufacturerContacts.mockResolvedValue([bosch]);
    await loadManufacturerContacts();

    mockListManufacturerContacts.mockRejectedValue(new Error('offline'));
    const contacts = await loadManufacturerContacts();

    expect(contacts).toEqual({ [bosch.brand]: bosch });
  });

  it('falls back to a pre-existing cache when the fetch fails', async () => {
    await saveManufacturerContactsCache([samsung]);
    mockListManufacturerContacts.mockRejectedValue(new Error('offline'));

    const contacts = await loadManufacturerContacts();

    expect(contacts).toEqual({ [samsung.brand]: samsung });
  });

  it('returns an empty map when the fetch fails and there is no cache', async () => {
    mockListManufacturerContacts.mockRejectedValue(new Error('offline'));

    expect(await loadManufacturerContacts()).toEqual({});
  });
});
