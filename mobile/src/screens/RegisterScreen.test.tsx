import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import RegisterScreen from './RegisterScreen';
import { ApiError } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { createMockNavigation, createMockRoute } from '../testUtils/navigation';

jest.mock('../context/AuthContext', () => ({
  useAuth: jest.fn(),
}));

const mockUseAuth = useAuth as jest.Mock;

function renderScreen(params?: { inviteCode?: string }) {
  const navigation = createMockNavigation();
  render(<RegisterScreen navigation={navigation as any} route={createMockRoute(params) as any} />);
  return navigation;
}

beforeEach(() => {
  mockUseAuth.mockReturnValue({ register: jest.fn().mockResolvedValue(undefined) });
});

describe('RegisterScreen', () => {
  it('pre-fills the invite code from route params', () => {
    renderScreen({ inviteCode: 'ABCD1234' });
    expect(screen.getByDisplayValue('ABCD1234')).toBeTruthy();
  });

  it('shows a validation error when required fields are missing', async () => {
    const register = jest.fn();
    mockUseAuth.mockReturnValue({ register });
    renderScreen();

    fireEvent.press(screen.getByText('הרשמה'));

    expect(await screen.findByText('נא למלא את כל השדות')).toBeTruthy();
    expect(register).not.toHaveBeenCalled();
  });

  it('shows a validation error when the password is too short', async () => {
    const register = jest.fn();
    mockUseAuth.mockReturnValue({ register });
    renderScreen();

    fireEvent.changeText(screen.getByPlaceholderText('שם מלא'), 'Ron');
    fireEvent.changeText(screen.getByPlaceholderText('אימייל'), 'ron@example.com');
    fireEvent.changeText(screen.getByPlaceholderText('סיסמה (8 תווים לפחות)'), 'short');
    fireEvent.press(screen.getByText('הרשמה'));

    expect(await screen.findByText('הסיסמה חייבת להכיל לפחות 8 תווים')).toBeTruthy();
    expect(register).not.toHaveBeenCalled();
  });

  it('calls register with trimmed values and the invite code when everything is valid', async () => {
    const register = jest.fn().mockResolvedValue(undefined);
    mockUseAuth.mockReturnValue({ register });
    renderScreen();

    fireEvent.changeText(screen.getByPlaceholderText('שם מלא'), '  Ron Cohen  ');
    fireEvent.changeText(screen.getByPlaceholderText('אימייל'), '  ron@example.com  ');
    fireEvent.changeText(screen.getByPlaceholderText('סיסמה (8 תווים לפחות)'), 'supersecret1');
    fireEvent.changeText(
      screen.getByPlaceholderText('קוד הזמנה (אופציונלי, אם מצטרפים למשק בית קיים)'),
      '  ABCD1234  '
    );
    fireEvent.press(screen.getByText('הרשמה'));

    await waitFor(() =>
      expect(register).toHaveBeenCalledWith('ron@example.com', 'supersecret1', 'Ron Cohen', 'ABCD1234')
    );
  });

  it('omits the invite code entirely when left blank', async () => {
    const register = jest.fn().mockResolvedValue(undefined);
    mockUseAuth.mockReturnValue({ register });
    renderScreen();

    fireEvent.changeText(screen.getByPlaceholderText('שם מלא'), 'Ron');
    fireEvent.changeText(screen.getByPlaceholderText('אימייל'), 'ron@example.com');
    fireEvent.changeText(screen.getByPlaceholderText('סיסמה (8 תווים לפחות)'), 'supersecret1');
    fireEvent.press(screen.getByText('הרשמה'));

    await waitFor(() =>
      expect(register).toHaveBeenCalledWith('ron@example.com', 'supersecret1', 'Ron', undefined)
    );
  });

  it('shows the server-provided message when registration rejects with an ApiError', async () => {
    const register = jest.fn().mockRejectedValue(new ApiError(409, 'האימייל כבר רשום'));
    mockUseAuth.mockReturnValue({ register });
    renderScreen();

    fireEvent.changeText(screen.getByPlaceholderText('שם מלא'), 'Ron');
    fireEvent.changeText(screen.getByPlaceholderText('אימייל'), 'ron@example.com');
    fireEvent.changeText(screen.getByPlaceholderText('סיסמה (8 תווים לפחות)'), 'supersecret1');
    fireEvent.press(screen.getByText('הרשמה'));

    expect(await screen.findByText('האימייל כבר רשום')).toBeTruthy();
  });

  it('navigates to Login when the sign-in link is pressed', () => {
    const navigation = renderScreen();
    fireEvent.press(screen.getByText('כבר יש לי חשבון — התחברות'));
    expect(navigation.navigate).toHaveBeenCalledWith('Login');
  });
});
