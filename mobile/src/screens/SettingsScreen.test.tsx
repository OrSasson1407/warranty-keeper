import { Alert, Share } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import SettingsScreen from './SettingsScreen';
import { api } from '../api/client';
import type { GmailStatus, Household } from '../api/types';
import { useAuth } from '../context/AuthContext';
import { registerForExpiryPush } from '../notifications/registerPush';
import { extractAuthCode, extractCodeVerifier, useGmailConnectRequest } from '../auth/gmailConnect';
import { isGoogleSignInConfigured } from '../auth/googleSignIn';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('../context/AuthContext', () => ({ useAuth: jest.fn() }));
jest.mock('../api/client', () => ({
  api: {
    getMyHousehold: jest.fn(),
    upgradeHousehold: jest.fn(),
    gmailStatus: jest.fn(),
    connectGmail: jest.fn(),
    disconnectGmail: jest.fn(),
  },
  ApiError: jest.requireActual('../api/client').ApiError,
}));
jest.mock('../notifications/registerPush', () => ({ registerForExpiryPush: jest.fn() }));
jest.mock('../auth/gmailConnect', () => ({
  useGmailConnectRequest: jest.fn(),
  extractAuthCode: jest.fn(),
  extractCodeVerifier: jest.fn(),
}));
jest.mock('../auth/googleSignIn', () => ({ isGoogleSignInConfigured: jest.fn() }));

const mockUseAuth = useAuth as jest.Mock;
const mockGetMyHousehold = api.getMyHousehold as jest.Mock;
const mockUpgradeHousehold = api.upgradeHousehold as jest.Mock;
const mockRegisterForExpiryPush = registerForExpiryPush as jest.Mock;
const mockGmailStatus = api.gmailStatus as jest.Mock;
const mockConnectGmail = api.connectGmail as jest.Mock;
const mockDisconnectGmail = api.disconnectGmail as jest.Mock;
const mockUseGmailConnectRequest = useGmailConnectRequest as jest.Mock;
const mockExtractAuthCode = extractAuthCode as jest.Mock;
const mockExtractCodeVerifier = extractCodeVerifier as jest.Mock;
const mockIsGoogleSignInConfigured = isGoogleSignInConfigured as jest.Mock;

const mockPromptGmailConnect = jest.fn();

function gmailStatus(overrides: Partial<GmailStatus> = {}): GmailStatus {
  return { connected: false, last_scan_at: null, ...overrides };
}

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
  mockGmailStatus.mockResolvedValue(gmailStatus());
  mockUseGmailConnectRequest.mockReturnValue([{ some: 'request' }, null, mockPromptGmailConnect]);
  mockExtractAuthCode.mockReturnValue(null);
  mockExtractCodeVerifier.mockReturnValue(null);
  mockIsGoogleSignInConfigured.mockReturnValue(true);
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

  it('hides the Gmail section entirely when Google sign-in is not configured', async () => {
    mockIsGoogleSignInConfigured.mockReturnValue(false);
    renderScreen();
    await screen.findByText('משק בית: "הבית של מיכל כהן"');
    expect(screen.queryByText('ייבוא קבלות מ-Gmail')).toBeNull();
  });

  it('shows a connect button when Gmail is not connected', async () => {
    renderScreen();
    expect(await screen.findByText('התחברות ל-Gmail')).toBeTruthy();
  });

  it('shows the connected address and a disconnect button once connected', async () => {
    mockGmailStatus.mockResolvedValue(
      gmailStatus({ connected: true, gmail_address: 'me@gmail.com' }),
    );
    renderScreen();
    expect(await screen.findByText('מחובר: me@gmail.com')).toBeTruthy();
    expect(screen.getByText('ניתוק Gmail')).toBeTruthy();
    expect(screen.queryByText('התחברות ל-Gmail')).toBeNull();
  });

  it('prompts the Gmail connect flow when the connect button is pressed', async () => {
    renderScreen();
    fireEvent.press(await screen.findByText('התחברות ל-Gmail'));
    expect(mockPromptGmailConnect).toHaveBeenCalled();
  });

  it('connects Gmail once the OAuth response yields a code', async () => {
    mockExtractAuthCode.mockReturnValue('auth-code-1');
    mockExtractCodeVerifier.mockReturnValue('verifier-1');
    mockUseGmailConnectRequest.mockReturnValue([
      { redirectUri: 'http://localhost:8081' },
      { type: 'success', params: { code: 'auth-code-1' } },
      mockPromptGmailConnect,
    ]);
    mockConnectGmail.mockResolvedValue(
      gmailStatus({ connected: true, gmail_address: 'me@gmail.com' }),
    );

    renderScreen();

    await waitFor(() =>
      expect(mockConnectGmail).toHaveBeenCalledWith(
        'auth-code-1',
        'http://localhost:8081',
        'verifier-1',
      ),
    );
    expect(await screen.findByText('מחובר: me@gmail.com')).toBeTruthy();
  });

  it('shows an error when connecting Gmail fails', async () => {
    mockExtractAuthCode.mockReturnValue('auth-code-1');
    mockUseGmailConnectRequest.mockReturnValue([
      { redirectUri: 'http://localhost:8081' },
      { type: 'success', params: { code: 'auth-code-1' } },
      mockPromptGmailConnect,
    ]);
    mockConnectGmail.mockRejectedValue(new Error('network error'));

    renderScreen();

    await waitFor(() => expect(Alert.alert).toHaveBeenCalledWith('שגיאה', expect.any(String)));
  });

  it('disconnects Gmail when the disconnect button is pressed', async () => {
    mockGmailStatus.mockResolvedValue(
      gmailStatus({ connected: true, gmail_address: 'me@gmail.com' }),
    );
    mockDisconnectGmail.mockResolvedValue({ disconnected: true });
    renderScreen();

    fireEvent.press(await screen.findByText('ניתוק Gmail'));

    await waitFor(() => expect(mockDisconnectGmail).toHaveBeenCalled());
    expect(await screen.findByText('התחברות ל-Gmail')).toBeTruthy();
  });

  it('navigates to GmailReceipts when "צפייה בקבלות שנמצאו" is pressed', async () => {
    mockGmailStatus.mockResolvedValue(
      gmailStatus({ connected: true, gmail_address: 'me@gmail.com' }),
    );
    const navigation = renderScreen();

    fireEvent.press(await screen.findByText('צפייה בקבלות שנמצאו'));
    expect(navigation.navigate).toHaveBeenCalledWith('GmailReceipts');
  });

  it('calls logout when the logout button is pressed', async () => {
    const logout = jest.fn();
    mockUseAuth.mockReturnValue({ user: { id: 'u1' }, logout });
    renderScreen();

    fireEvent.press(await screen.findByText('התנתקות'));
    expect(logout).toHaveBeenCalled();
  });
});
