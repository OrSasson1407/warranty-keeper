import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import LoginScreen from './LoginScreen';
import { ApiError } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('../context/AuthContext', () => ({
  useAuth: jest.fn(),
}));

const mockUseAuth = useAuth as jest.Mock;

function renderScreen(navigation = createMockNavigation()) {
  render(<LoginScreen navigation={navigation as any} route={{} as any} />);
  return navigation;
}

beforeEach(() => {
  mockUseAuth.mockReturnValue({ login: jest.fn().mockResolvedValue(undefined) });
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
});
