import { Platform } from 'react-native';
import * as Calendar from 'expo-calendar';

// On-device calendar export (tier 1 of the calendar-sync backlog item):
// adds a single all-day event on the warranty expiry date to a writable
// device calendar. One-way and one-shot -- it does not keep the event in
// sync if the warranty date changes later, and there's no OAuth/account
// sync (that would be tier 2, a separate and much larger piece of work).
export async function addWarrantyExpiryToCalendar(
  productName: string,
  warrantyExpiresAt: string,
): Promise<boolean> {
  if (Platform.OS === 'web') return false;

  const { status } = await Calendar.requestCalendarPermissions();
  if (status !== 'granted') return false;

  const targetCalendar = await findWritableCalendar();
  if (!targetCalendar) return false;

  const date = new Date(warrantyExpiresAt);
  try {
    await targetCalendar.createEvent({
      title: `אחריות פגה: ${productName}`,
      notes: 'נוסף אוטומטית על ידי WarrantyKeeper',
      startDate: date,
      endDate: date,
      allDay: true,
    });
    return true;
  } catch {
    return false;
  }
}

async function findWritableCalendar(): Promise<Calendar.ExpoCalendar | null> {
  if (Platform.OS === 'ios') {
    try {
      return Calendar.getDefaultCalendarSync();
    } catch {
      return null;
    }
  }

  const calendars = await Calendar.getCalendars(Calendar.EntityTypes.EVENT);
  return (
    calendars.find((c) => c.isPrimary && c.allowsModifications) ??
    calendars.find((c) => c.allowsModifications) ??
    null
  );
}
