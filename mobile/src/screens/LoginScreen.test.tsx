import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import LoginScreen from './LoginScreen';
import { ApiError } from '../api/client';
import { extractIdToken, isGoogleSignInConfigured, useGoogleSignIn } from '../auth/googleSignIn';
import { useAuth } from '../context/AuthContext';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('../context/AuthContext', () => ({
  useAuth: jest.fn(),
}));
jest.mock('../auth/googleSignIn', () => ({
  useGoogleSignIn: jest.fn(),
  extractIdToken: jest.fn(),
  isGoogleSignInConfigured: jest.fn(),
}));

const mockUseAuth = useAuth as jest.Mock;
const mockUseGoogleSignIn = useGoogleSignIn as jest.Mock;
const mockExtractIdToken = extractIdToken as jest.Mock;
const mockIsGoogleSignInConfigured = isGoogleSignInConfigured as jest.Mock;

function renderScreen(navigation = createMockNavigation()) {
  render(<LoginScreen navigation={navigation as any} route={{} as any} />);
  return navigation;
}

const mockPromptGoogleSignIn = jest.fn();

beforeEach(() => {
  mockUseAuth.mockReturnValue({
    login: jest.fn().mockResolvedValue(undefined),
    loginWithGoogle: jest.fn().mockResolvedValue(undefined),
  });
  mockUseGoogleSignIn.mockReturnValue([{ some: 'request' }, null, mockPromptGoogleSignIn]);
  mockExtractIdToken.mockReturnValue(null);
  mockIsGoogleSignInConfigured.mockReturnValue(true);
});

afterEach(() => {
  jest.clearAllMocks();
});

describe('LoginScreen', () => {
  it('renders email and password fields', () => {
    renderScreen();
    expect(screen.getByPlaceholderText('אימייל')).toBeTruthy();
    expect(screen.getByPlaceholderText('סיסמה')).toBeTruthy();
  });

  it('calls login with the entered (and trimmed) email and the password as-is', async () => {
    const login = jest.fn().mockResolvedValue(undefined);
    mockUseAuth.mockReturnValue({ login });
    renderScreen();

    fireEvent.changeText(screen.getByPlaceholderText('אימייל'), '  a@example.com  ');
    fireEvent.changeText(screen.getByPlaceholderText('סיסמה'), 'supersecret1');
    fireEvent.press(screen.getAllByText('התחברות')[1]);

    await waitFor(() => expect(login).toHaveBeenCalledWith('a@example.com', 'supersecret1'));
  });

  it('shows the server-provided message when login rejects with an ApiError', async () => {
    const login = jest.fn().mockRejectedValue(new ApiError(401, 'סיסמה שגויה'));
    mockUseAuth.mockReturnValue({ login });
    renderScreen();

    fireEvent.press(screen.getAllByText('התחברות')[1]);

    await waitFor(() => expect(screen.getByText('סיסמה שגויה')).toBeTruthy());
  });

  it('shows a generic message when login rejects with a non-ApiError', async () => {
    const login = jest.fn().mockRejectedValue(new Error('network down'));
    mockUseAuth.mockReturnValue({ login });
    renderScreen();

    fireEvent.press(screen.getAllByText('התחברות')[1]);

    await waitFor(() => expect(screen.getByText('ההתחברות נכשלה, נסו שוב')).toBeTruthy());
  });

  it('navigates to Register when the sign-up link is pressed', () => {
    const navigation = renderScreen();
    fireEvent.press(screen.getByText('אין לי חשבון — הרשמה'));
    expect(navigation.navigate).toHaveBeenCalledWith('Register');
  });

  it('hides the Google button entirely when Google sign-in is not configured', () => {
    mockIsGoogleSignInConfigured.mockReturnValue(false);
    renderScreen();
    expect(screen.queryByText('התחברות עם Google')).toBeNull();
  });

  it('prompts the Google sign-in flow when its button is pressed', () => {
    renderScreen();
    fireEvent.press(screen.getByText('התחברות עם Google'));
    expect(mockPromptGoogleSignIn).toHaveBeenCalled();
  });

  it('disables the Google button until the auth request has loaded', () => {
    mockUseGoogleSignIn.mockReturnValue([null, null, mockPromptGoogleSignIn]);
    renderScreen();
    fireEvent.press(screen.getByText('התחברות עם Google'));
    expect(mockPromptGoogleSignIn).not.toHaveBeenCalled();
  });

  it('logs in with the extracted id_token once the Google response succeeds', async () => {
    const loginWithGoogle = jest.fn().mockResolvedValue(undefined);
    mockUseAuth.mockReturnValue({ login: jest.fn(), loginWithGoogle });
    mockExtractIdToken.mockReturnValue('a-real-id-token');
    mockUseGoogleSignIn.mockReturnValue([
      { some: 'request' },
      { type: 'success', params: { id_token: 'a-real-id-token' } },
      mockPromptGoogleSignIn,
    ]);

    renderScreen();

    await waitFor(() => expect(loginWithGoogle).toHaveBeenCalledWith('a-real-id-token'));
  });

  it('shows an error when the Google login rejects', async () => {
    const loginWithGoogle = jest.fn().mockRejectedValue(new ApiError(401, 'טוקן לא תקין'));
    mockUseAuth.mockReturnValue({ login: jest.fn(), loginWithGoogle });
    mockExtractIdToken.mockReturnValue('bad-token');
    mockUseGoogleSignIn.mockReturnValue([
      { some: 'request' },
      { type: 'success', params: { id_token: 'bad-token' } },
      mockPromptGoogleSignIn,
    ]);

    renderScreen();

    await waitFor(() => expect(screen.getByText('טוקן לא תקין')).toBeTruthy());
  });

  it('does nothing when the Google response has no id_token (cancelled, etc.)', () => {
    const loginWithGoogle = jest.fn();
    mockUseAuth.mockReturnValue({ login: jest.fn(), loginWithGoogle });
    mockExtractIdToken.mockReturnValue(null);
    mockUseGoogleSignIn.mockReturnValue([
      { some: 'request' },
      { type: 'cancel' },
      mockPromptGoogleSignIn,
    ]);

    renderScreen();

    expect(loginWithGoogle).not.toHaveBeenCalled();
  });
});
