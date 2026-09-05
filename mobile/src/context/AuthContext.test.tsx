import { Text, TouchableOpacity } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import { AuthProvider, useAuth } from './AuthContext';
import { api } from '../api/client';
import * as tokenStore from '../api/tokenStore';
import type { AuthResponse, User } from '../api/types';

jest.mock('../api/client', () => ({
  api: { register: jest.fn(), login: jest.fn() },
}));
jest.mock('../api/tokenStore', () => ({
  getAccessToken: jest.fn(),
  getRefreshToken: jest.fn(),
  setTokens: jest.fn(),
  setAccessToken: jest.fn(),
  clearTokens: jest.fn(),
}));

const mockRegister = api.register as jest.Mock;
const mockLogin = api.login as jest.Mock;
const mockTokenStore = tokenStore as jest.Mocked<typeof tokenStore>;

const USER_KEY = 'wk_user';

function testUser(overrides: Partial<User> = {}): User {
  return {
    id: 'u1',
    email: 'a@example.com',
    full_name: 'Ron Cohen',
    household_id: 'h1',
    ...overrides,
  };
}

function authResponse(overrides: Partial<AuthResponse> = {}): AuthResponse {
  return {
    access_token: 'access-1',
    refresh_token: 'refresh-1',
    user: testUser(),
    ...overrides,
  };
}

function TestConsumer() {
  const { user, isLoading, register, login, logout } = useAuth();
  return (
    <>
      <Text testID="loading">{String(isLoading)}</Text>
      <Text testID="user">{user ? user.full_name : 'none'}</Text>
      <TouchableOpacity
        testID="register-btn"
        onPress={() => register('a@example.com', 'pw', 'Ron Cohen', 'CODE1234').catch(() => {})}
      >
        <Text>register</Text>
      </TouchableOpacity>
      <TouchableOpacity
        testID="login-btn"
        onPress={() => login('a@example.com', 'pw').catch(() => {})}
      >
        <Text>login</Text>
      </TouchableOpacity>
      <TouchableOpacity testID="logout-btn" onPress={() => logout()}>
        <Text>logout</Text>
      </TouchableOpacity>
    </>
  );
}

function renderProvider() {
  return render(
    <AuthProvider>
      <TestConsumer />
    </AuthProvider>,
  );
}

beforeEach(async () => {
  await AsyncStorage.clear();
  mockTokenStore.getAccessToken.mockResolvedValue(null);
  mockTokenStore.setTokens.mockResolvedValue(undefined);
  mockTokenStore.clearTokens.mockResolvedValue(undefined);
});

afterEach(() => {
  jest.clearAllMocks();
});

describe('AuthProvider bootstrap', () => {
  it('starts loading and settles to no user when nothing is stored', async () => {
    renderProvider();
    expect(screen.getByTestId('loading').props.children).toBe('true');

    await waitFor(() => expect(screen.getByTestId('loading').props.children).toBe('false'));
    expect(screen.getByTestId('user').props.children).toBe('none');
  });

  it('restores the session when both a token and a stored user exist', async () => {
    mockTokenStore.getAccessToken.mockResolvedValue('stored-token');
    await AsyncStorage.setItem(USER_KEY, JSON.stringify(testUser({ full_name: 'מיכל כהן' })));

    renderProvider();

    await waitFor(() => expect(screen.getByTestId('user').props.children).toBe('מיכל כהן'));
  });

  it('does not restore a session when the token is missing, even with a stored user', async () => {
    mockTokenStore.getAccessToken.mockResolvedValue(null);
    await AsyncStorage.setItem(USER_KEY, JSON.stringify(testUser()));

    renderProvider();

    await waitFor(() => expect(screen.getByTestId('loading').props.children).toBe('false'));
    expect(screen.getByTestId('user').props.children).toBe('none');
  });

  it('does not restore a session when the stored user is missing, even with a token', async () => {
    mockTokenStore.getAccessToken.mockResolvedValue('stored-token');
    // no AsyncStorage user set

    renderProvider();

    await waitFor(() => expect(screen.getByTestId('loading').props.children).toBe('false'));
    expect(screen.getByTestId('user').props.children).toBe('none');
  });
});

describe('register', () => {
  it('persists tokens and the user, and updates context state', async () => {
    mockRegister.mockResolvedValue(authResponse());
    renderProvider();
    await waitFor(() => expect(screen.getByTestId('loading').props.children).toBe('false'));

    fireEvent.press(screen.getByTestId('register-btn'));

    await waitFor(() => expect(screen.getByTestId('user').props.children).toBe('Ron Cohen'));
    expect(mockRegister).toHaveBeenCalledWith('a@example.com', 'pw', 'Ron Cohen', 'CODE1234');
    expect(mockTokenStore.setTokens).toHaveBeenCalledWith('access-1', 'refresh-1');
    expect(JSON.parse((await AsyncStorage.getItem(USER_KEY))!)).toEqual(testUser());
  });

  it('propagates a rejection without changing state', async () => {
    mockRegister.mockRejectedValue(new Error('email taken'));
    renderProvider();
    await waitFor(() => expect(screen.getByTestId('loading').props.children).toBe('false'));

    fireEvent.press(screen.getByTestId('register-btn'));

    await waitFor(() => expect(mockRegister).toHaveBeenCalled());
    expect(screen.getByTestId('user').props.children).toBe('none');
    expect(mockTokenStore.setTokens).not.toHaveBeenCalled();
  });
});

describe('login', () => {
  it('persists tokens and the user, and updates context state', async () => {
    mockLogin.mockResolvedValue(authResponse({ user: testUser({ full_name: 'דני כהן' }) }));
    renderProvider();
    await waitFor(() => expect(screen.getByTestId('loading').props.children).toBe('false'));

    fireEvent.press(screen.getByTestId('login-btn'));

    await waitFor(() => expect(screen.getByTestId('user').props.children).toBe('דני כהן'));
    expect(mockLogin).toHaveBeenCalledWith('a@example.com', 'pw');
    expect(mockTokenStore.setTokens).toHaveBeenCalledWith('access-1', 'refresh-1');
  });
});

describe('logout', () => {
  it('clears tokens, removes the stored user, and resets context state', async () => {
    mockLogin.mockResolvedValue(authResponse());
    renderProvider();
    await waitFor(() => expect(screen.getByTestId('loading').props.children).toBe('false'));

    fireEvent.press(screen.getByTestId('login-btn'));
    await waitFor(() => expect(screen.getByTestId('user').props.children).toBe('Ron Cohen'));

    fireEvent.press(screen.getByTestId('logout-btn'));

    await waitFor(() => expect(screen.getByTestId('user').props.children).toBe('none'));
    expect(mockTokenStore.clearTokens).toHaveBeenCalled();
    expect(await AsyncStorage.getItem(USER_KEY)).toBeNull();
  });
});

describe('useAuth', () => {
  it('throws when used outside an AuthProvider', () => {
    const consoleError = jest.spyOn(console, 'error').mockImplementation(() => {});
    function Bare() {
      useAuth();
      return null;
    }
    expect(() => render(<Bare />)).toThrow('useAuth must be used within AuthProvider');
    consoleError.mockRestore();
  });
});
