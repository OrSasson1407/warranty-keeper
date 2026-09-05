import { Platform } from 'react-native';
import * as Notifications from 'expo-notifications';

import { registerForExpiryPush } from './registerPush';
import { api } from '../api/client';

// `isDevice` is a getter (not a plain property) so mutating mockIsDeviceValue
// stays visible after Babel's namespace-import interop copies this module's
// exports — a plain property would get snapshotted into an independent copy
// per importing file, silently decoupling the test's mutation from the
// value registerPush.ts reads.
let mockIsDeviceValue = true;
jest.mock('expo-device', () => ({
  get isDevice() {
    return mockIsDeviceValue;
  },
}));
jest.mock('expo-notifications', () => ({
  setNotificationHandler: jest.fn(),
  getPermissionsAsync: jest.fn(),
  requestPermissionsAsync: jest.fn(),
  setNotificationChannelAsync: jest.fn(),
  getExpoPushTokenAsync: jest.fn(),
  AndroidImportance: { DEFAULT: 3 },
}));
jest.mock('../api/client', () => ({ api: { registerDevice: jest.fn() } }));

const mockGetPermissions = Notifications.getPermissionsAsync as jest.Mock;
const mockRequestPermissions = Notifications.requestPermissionsAsync as jest.Mock;
const mockSetChannel = Notifications.setNotificationChannelAsync as jest.Mock;
const mockGetToken = Notifications.getExpoPushTokenAsync as jest.Mock;
const mockRegisterDevice = api.registerDevice as jest.Mock;

const originalOS = Platform.OS;

beforeEach(() => {
  mockIsDeviceValue = true;
  mockGetPermissions.mockResolvedValue({ status: 'granted' });
  mockRequestPermissions.mockResolvedValue({ status: 'granted' });
  mockGetToken.mockResolvedValue({ data: 'ExponentPushToken[abc123]' });
  mockRegisterDevice.mockResolvedValue({ status: 'registered' });
});

afterEach(() => {
  jest.clearAllMocks();
  Object.defineProperty(Platform, 'OS', { value: originalOS, configurable: true });
});

function setPlatform(os: string) {
  Object.defineProperty(Platform, 'OS', { value: os, configurable: true });
}

describe('registerForExpiryPush', () => {
  it('returns false immediately on web, without checking permissions', async () => {
    setPlatform('web');
    const result = await registerForExpiryPush();

    expect(result).toBe(false);
    expect(mockGetPermissions).not.toHaveBeenCalled();
    expect(mockRegisterDevice).not.toHaveBeenCalled();
  });

  it('returns false on a simulator/emulator (Device.isDevice is false)', async () => {
    setPlatform('ios');
    mockIsDeviceValue = false;
    const result = await registerForExpiryPush();

    expect(result).toBe(false);
    expect(mockGetPermissions).not.toHaveBeenCalled();
  });

  it('skips requesting permission when it is already granted', async () => {
    setPlatform('ios');
    mockGetPermissions.mockResolvedValue({ status: 'granted' });

    const result = await registerForExpiryPush();

    expect(result).toBe(true);
    expect(mockRequestPermissions).not.toHaveBeenCalled();
  });

  it('requests permission when not already granted, and proceeds if the user allows it', async () => {
    setPlatform('ios');
    mockGetPermissions.mockResolvedValue({ status: 'undetermined' });
    mockRequestPermissions.mockResolvedValue({ status: 'granted' });

    const result = await registerForExpiryPush();

    expect(result).toBe(true);
    expect(mockRequestPermissions).toHaveBeenCalled();
    expect(mockRegisterDevice).toHaveBeenCalledWith('ExponentPushToken[abc123]');
  });

  it('returns false without registering a device when permission is denied', async () => {
    setPlatform('ios');
    mockGetPermissions.mockResolvedValue({ status: 'undetermined' });
    mockRequestPermissions.mockResolvedValue({ status: 'denied' });

    const result = await registerForExpiryPush();

    expect(result).toBe(false);
    expect(mockGetToken).not.toHaveBeenCalled();
    expect(mockRegisterDevice).not.toHaveBeenCalled();
  });

  it('creates the Android notification channel only on Android', async () => {
    setPlatform('android');
    await registerForExpiryPush();
    expect(mockSetChannel).toHaveBeenCalledWith('default', expect.objectContaining({ importance: 3 }));
  });

  it('does not touch the notification channel on iOS', async () => {
    setPlatform('ios');
    await registerForExpiryPush();
    expect(mockSetChannel).not.toHaveBeenCalled();
  });

  it('registers the Expo push token with the API and returns true on success', async () => {
    setPlatform('ios');
    mockGetToken.mockResolvedValue({ data: 'ExponentPushToken[xyz789]' });

    const result = await registerForExpiryPush();

    expect(result).toBe(true);
    expect(mockRegisterDevice).toHaveBeenCalledWith('ExponentPushToken[xyz789]');
  });

  it('returns false if fetching the push token fails', async () => {
    setPlatform('ios');
    mockGetToken.mockRejectedValue(new Error('no push service available'));

    const result = await registerForExpiryPush();

    expect(result).toBe(false);
    expect(mockRegisterDevice).not.toHaveBeenCalled();
  });

  it('returns false if registering the device with the API fails', async () => {
    setPlatform('ios');
    mockRegisterDevice.mockRejectedValue(new Error('network error'));

    const result = await registerForExpiryPush();

    expect(result).toBe(false);
  });
});
