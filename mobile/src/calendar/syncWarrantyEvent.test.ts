import { Platform } from 'react-native';
import * as Calendar from 'expo-calendar';

import { addWarrantyExpiryToCalendar } from './syncWarrantyEvent';

jest.mock('expo-calendar', () => ({
  requestCalendarPermissions: jest.fn(),
  getCalendars: jest.fn(),
  getDefaultCalendarSync: jest.fn(),
  EntityTypes: { EVENT: 'event' },
}));

const mockRequestPermissions = Calendar.requestCalendarPermissions as jest.Mock;
const mockGetCalendars = Calendar.getCalendars as jest.Mock;
const mockGetDefaultCalendarSync = Calendar.getDefaultCalendarSync as jest.Mock;

const originalOS = Platform.OS;

function setPlatform(os: string) {
  Object.defineProperty(Platform, 'OS', { value: os, configurable: true });
}

function fakeCalendar(
  overrides: Partial<{ isPrimary: boolean; allowsModifications: boolean }> = {},
) {
  return {
    id: 'cal1',
    isPrimary: overrides.isPrimary ?? false,
    allowsModifications: overrides.allowsModifications ?? true,
    createEvent: jest.fn().mockResolvedValue({ id: 'event1' }),
  };
}

beforeEach(() => {
  jest.clearAllMocks();
  mockRequestPermissions.mockResolvedValue({ status: 'granted' });
});

afterEach(() => {
  Object.defineProperty(Platform, 'OS', { value: originalOS, configurable: true });
});

describe('addWarrantyExpiryToCalendar', () => {
  it('returns false immediately on web, without requesting permission', async () => {
    setPlatform('web');
    const result = await addWarrantyExpiryToCalendar('מוצר', '2028-01-01');

    expect(result).toBe(false);
    expect(mockRequestPermissions).not.toHaveBeenCalled();
  });

  it('returns false when calendar permission is denied', async () => {
    setPlatform('android');
    mockRequestPermissions.mockResolvedValue({ status: 'denied' });

    const result = await addWarrantyExpiryToCalendar('מוצר', '2028-01-01');

    expect(result).toBe(false);
    expect(mockGetCalendars).not.toHaveBeenCalled();
  });

  it('on iOS, uses the default calendar directly', async () => {
    setPlatform('ios');
    const cal = fakeCalendar();
    mockGetDefaultCalendarSync.mockReturnValue(cal);

    const result = await addWarrantyExpiryToCalendar('מזגן', '2028-01-01');

    expect(result).toBe(true);
    expect(cal.createEvent).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'אחריות פגה: מזגן', allDay: true }),
    );
  });

  it('returns false on iOS if there is no default calendar', async () => {
    setPlatform('ios');
    mockGetDefaultCalendarSync.mockImplementation(() => {
      throw new Error('no default calendar');
    });

    const result = await addWarrantyExpiryToCalendar('מזגן', '2028-01-01');

    expect(result).toBe(false);
  });

  it('on Android, prefers the primary writable calendar', async () => {
    setPlatform('android');
    const primary = fakeCalendar({ isPrimary: true, allowsModifications: true });
    const other = fakeCalendar({ isPrimary: false, allowsModifications: true });
    mockGetCalendars.mockResolvedValue([other, primary]);

    const result = await addWarrantyExpiryToCalendar('מזגן', '2028-01-01');

    expect(result).toBe(true);
    expect(primary.createEvent).toHaveBeenCalled();
    expect(other.createEvent).not.toHaveBeenCalled();
  });

  it('on Android, falls back to any writable calendar when there is no primary', async () => {
    setPlatform('android');
    const writable = fakeCalendar({ isPrimary: false, allowsModifications: true });
    mockGetCalendars.mockResolvedValue([writable]);

    const result = await addWarrantyExpiryToCalendar('מזגן', '2028-01-01');

    expect(result).toBe(true);
    expect(writable.createEvent).toHaveBeenCalled();
  });

  it('returns false on Android when no calendar is writable', async () => {
    setPlatform('android');
    mockGetCalendars.mockResolvedValue([fakeCalendar({ allowsModifications: false })]);

    const result = await addWarrantyExpiryToCalendar('מזגן', '2028-01-01');

    expect(result).toBe(false);
  });

  it('returns false if creating the event throws', async () => {
    setPlatform('ios');
    const cal = fakeCalendar();
    cal.createEvent.mockRejectedValue(new Error('failed to create event'));
    mockGetDefaultCalendarSync.mockReturnValue(cal);

    const result = await addWarrantyExpiryToCalendar('מזגן', '2028-01-01');

    expect(result).toBe(false);
  });
});
