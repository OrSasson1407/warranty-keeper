import { Alert, Share } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import SettingsScreen from './SettingsScreen';
import { api } from '../api/client';
import type { Household } from '../api/types';
import { useAuth } from '../context/AuthContext';
import { registerForExpiryPush } from '../notifications/registerPush';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('../context/AuthContext', () => ({ useAuth: jest.fn() }));
jest.mock('../api/client', () => ({
  api: { getMyHousehold: jest.fn(), upgradeHousehold: jest.fn() },
}));
jest.mock('../notifications/registerPush', () => ({ registerForExpiryPush: jest.fn() }));

const mockUseAuth = useAuth as jest.Mock;
const mockGetMyHousehold = api.getMyHousehold as jest.Mock;
const mockUpgradeHousehold = api.upgradeHousehold as jest.Mock;
const mockRegisterForExpiryPush = registerForExpiryPush as jest.Mock;

function household(overrides: Partial<Household> = {}): Household {
  return {
    id: 'h1',
    name: 'הבית של מיכל כהן',
    invite_code: 'ABCD1234',
    tier: 'free',
    members: [{ id: 'u1', full_name: 'מיכל כהן', email: 'michal@example.com' }],
    ...overrides,
  };
}

beforeEach(async () => {
  await AsyncStorage.clear();
  mockUseAuth.mockReturnValue({ user: { id: 'u1' }, logout: jest.fn() });
  mockGetMyHousehold.mockResolvedValue(household());
  mockUpgradeHousehold.mockResolvedValue({ tier: 'premium' });
  mockRegisterForExpiryPush.mockResolvedValue(true);
  jest.spyOn(Alert, 'alert').mockImplementation(() => {});
  jest.spyOn(Share, 'share').mockResolvedValue({ action: Share.sharedAction } as any);
});

afterEach(() => {
  jest.restoreAllMocks();
});

function renderScreen() {
  const navigation = createMockNavigation();
  render(<SettingsScreen navigation={navigation as any} route={{} as any} />);
  return navigation;
}

describe('SettingsScreen', () => {
  it('shows the household name and marks the current user among members', async () => {
    renderScreen();
    expect(await screen.findByText('משק בית: "הבית של מיכל כהן"')).toBeTruthy();
    expect(screen.getByText('מיכל כהן (את/ה)')).toBeTruthy();
  });

  it('offers an invite when the household has fewer than 2 members', async () => {
    renderScreen();
    expect(await screen.findByText('+ הזמן בן/בת משפחה (קוד: ABCD1234)')).toBeTruthy();
  });

  it('shows a "household full" note instead of an invite once there are 2 members', async () => {
    mockGetMyHousehold.mockResolvedValue(
      household({
        members: [
          { id: 'u1', full_name: 'מיכל כהן', email: 'michal@example.com' },
          { id: 'u2', full_name: 'דני כהן', email: 'dani@example.com' },
        ],
      }),
    );
    renderScreen();
    expect(await screen.findByText('משק הבית מלא (מקסימום 2 חברים)')).toBeTruthy();
  });

  it('shares an invite message with the household name and code', async () => {
    renderScreen();
    fireEvent.press(await screen.findByText('+ הזמן בן/בת משפחה (קוד: ABCD1234)'));

    await waitFor(() =>
      expect(Share.share).toHaveBeenCalledWith({
        message: expect.stringContaining('ABCD1234'),
      }),
    );
  });

  it('turns push on and persists the preference when permission is granted', async () => {
    renderScreen();
    const toggle = await screen.findByRole('switch');

    fireEvent(toggle, 'valueChange', true);

    await waitFor(() => expect(mockRegisterForExpiryPush).toHaveBeenCalled());
    await waitFor(async () => expect(await AsyncStorage.getItem('wk_push_enabled')).toBe('true'));
  });

  it('reverts and alerts when push permission is denied', async () => {
    mockRegisterForExpiryPush.mockResolvedValue(false);
    renderScreen();
    const toggle = await screen.findByRole('switch');

    fireEvent(toggle, 'valueChange', true);

    await waitFor(() =>
      expect(Alert.alert).toHaveBeenCalledWith('לא ניתן להפעיל התראות', expect.any(String)),
    );
    await waitFor(async () => expect(await AsyncStorage.getItem('wk_push_enabled')).toBe('false'));
  });

  it('starts with the switch already on if the preference was previously saved', async () => {
    await AsyncStorage.setItem('wk_push_enabled', 'true');
    renderScreen();

    const toggle = await screen.findByRole('switch');
    await waitFor(() => expect(toggle.props.value).toBe(true));
  });

  it('shows the free plan note and an upgrade button on the free tier', async () => {
    renderScreen();
    expect(await screen.findByText('תוכנית חינמית — עד 20 מוצרים')).toBeTruthy();
    expect(screen.getByText('שדרגו ל-Premium')).toBeTruthy();
  });

  it('shows the premium badge instead of an upgrade button on the premium tier', async () => {
    mockGetMyHousehold.mockResolvedValue(household({ tier: 'premium' }));
    renderScreen();
    expect(await screen.findByText('⭐ Premium — ללא הגבלת מוצרים')).toBeTruthy();
    expect(screen.queryByText('שדרגו ל-Premium')).toBeNull();
  });

  it('upgrades and refreshes the household when "שדרגו ל-Premium" is pressed', async () => {
    mockGetMyHousehold
      .mockResolvedValueOnce(household({ tier: 'free' }))
      .mockResolvedValueOnce(household({ tier: 'premium' }));
    renderScreen();

    fireEvent.press(await screen.findByText('שדרגו ל-Premium'));

    await waitFor(() => expect(mockUpgradeHousehold).toHaveBeenCalled());
    expect(await screen.findByText('⭐ Premium — ללא הגבלת מוצרים')).toBeTruthy();
    await waitFor(() =>
      expect(Alert.alert).toHaveBeenCalledWith('שודרג בהצלחה', expect.any(String)),
    );
  });

  it('alerts on failure without changing the displayed tier', async () => {
    mockUpgradeHousehold.mockRejectedValue(new Error('network error'));
    renderScreen();

    fireEvent.press(await screen.findByText('שדרגו ל-Premium'));

    await waitFor(() => expect(Alert.alert).toHaveBeenCalledWith('שגיאה', expect.any(String)));
    expect(screen.getByText('תוכנית חינמית — עד 20 מוצרים')).toBeTruthy();
  });

  it('calls logout when the logout button is pressed', async () => {
    const logout = jest.fn();
    mockUseAuth.mockReturnValue({ user: { id: 'u1' }, logout });
    renderScreen();

    fireEvent.press(await screen.findByText('התנתקות'));
    expect(logout).toHaveBeenCalled();
  });
});
