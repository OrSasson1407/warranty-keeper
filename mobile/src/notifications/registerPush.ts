import { Platform } from 'react-native';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';

import { api } from '../api/client';

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: false,
    shouldSetBadge: false,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
});

// Registers this device for the single MVP push notification (30-day
// warranty expiry warning, see server/internal/notify). Returns false if
// permission was denied or push isn't available (e.g. web, simulator).
export async function registerForExpiryPush(): Promise<boolean> {
  if (Platform.OS === 'web') return false;
  if (!Device.isDevice) return false;

  const existing = await Notifications.getPermissionsAsync();
  let status = existing.status;
  if (status !== 'granted') {
    const requested = await Notifications.requestPermissionsAsync();
    status = requested.status;
  }
  if (status !== 'granted') return false;

  if (Platform.OS === 'android') {
    await Notifications.setNotificationChannelAsync('default', {
      name: 'default',
      importance: Notifications.AndroidImportance.DEFAULT,
    });
  }

  try {
    const tokenResponse = await Notifications.getExpoPushTokenAsync();
    await api.registerDevice(tokenResponse.data);
    return true;
  } catch {
    return false;
  }
}
