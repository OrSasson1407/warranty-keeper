import { render, screen } from '@testing-library/react-native';

import RootNavigator from './index';
import { useAuth } from '../context/AuthContext';

// Stub out every screen with a trivially identifiable component. The point
// of this test is RootNavigator's own logic (loading state -> which
// navigator -> which initial route) — not re-exercising each screen's real
// behavior, which already has its own test file and its own mocks for
// api/client, AuthContext, etc. that would otherwise all have to be
// satisfied simultaneously just to mount this tree.
//
// Jest's factory hoisting forbids referencing an outer-scope `Text` import
// here, so each factory requires react-native itself and uses
// React.createElement (no JSX) to build the stub without needing a
// lexically-scoped `Text` at all.
function mockStubScreen(label: string) {
  const RN = require('react-native');
  const React = require('react');
  function StubScreen() {
    return React.createElement(RN.Text, null, label);
  }
  return { __esModule: true, default: StubScreen };
}

jest.mock('../screens/OnboardingScreen', () => mockStubScreen('stub:Onboarding'));
jest.mock('../screens/LoginScreen', () => mockStubScreen('stub:Login'));
jest.mock('../screens/RegisterScreen', () => mockStubScreen('stub:Register'));
jest.mock('../screens/DashboardScreen', () => mockStubScreen('stub:Dashboard'));
jest.mock('../screens/AddProductChooseScreen', () => mockStubScreen('stub:AddProductChoose'));
jest.mock('../screens/ConfirmProductScreen', () => mockStubScreen('stub:ConfirmProduct'));
jest.mock('../screens/ProductDetailScreen', () => mockStubScreen('stub:ProductDetail'));
jest.mock('../screens/ClaimScreen', () => mockStubScreen('stub:Claim'));
jest.mock('../screens/SearchScreen', () => mockStubScreen('stub:Search'));
jest.mock('../screens/SettingsScreen', () => mockStubScreen('stub:Settings'));
jest.mock('../context/AuthContext', () => ({ useAuth: jest.fn() }));

const mockUseAuth = useAuth as jest.Mock;

describe('RootNavigator', () => {
  it('shows a loading indicator while auth state is still loading', () => {
    mockUseAuth.mockReturnValue({ user: null, isLoading: true });
    render(<RootNavigator />);

    expect(screen.queryByText('stub:Onboarding')).toBeNull();
    expect(screen.queryByText('stub:Dashboard')).toBeNull();
  });

  it('renders the auth stack, starting on Onboarding, when there is no user', () => {
    mockUseAuth.mockReturnValue({ user: null, isLoading: false });
    render(<RootNavigator />);

    expect(screen.getByText('stub:Onboarding')).toBeTruthy();
    expect(screen.queryByText('stub:Login')).toBeNull();
    expect(screen.queryByText('stub:Register')).toBeNull();
    expect(screen.queryByText('stub:Dashboard')).toBeNull();
  });

  it('renders the app stack, starting on Dashboard, once a user is present', () => {
    mockUseAuth.mockReturnValue({ user: { id: 'u1', full_name: 'Ron' }, isLoading: false });
    render(<RootNavigator />);

    expect(screen.getByText('stub:Dashboard')).toBeTruthy();
    expect(screen.queryByText('stub:Onboarding')).toBeNull();
  });

  it('switches from the auth stack to the app stack when the user logs in', () => {
    mockUseAuth.mockReturnValue({ user: null, isLoading: false });
    const { rerender } = render(<RootNavigator />);
    expect(screen.getByText('stub:Onboarding')).toBeTruthy();

    mockUseAuth.mockReturnValue({ user: { id: 'u1', full_name: 'Ron' }, isLoading: false });
    rerender(<RootNavigator />);

    expect(screen.getByText('stub:Dashboard')).toBeTruthy();
    expect(screen.queryByText('stub:Onboarding')).toBeNull();
  });
});
