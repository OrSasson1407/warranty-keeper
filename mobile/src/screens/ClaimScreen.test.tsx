import { Alert, Linking } from 'react-native';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import ClaimScreen from './ClaimScreen';
import { api, ApiError } from '../api/client';
import type { Product } from '../api/types';
import { createMockNavigation, createMockRoute } from '../testUtils/navigation';

jest.mock('../api/client', () => ({
  api: { getProduct: jest.fn(), createClaim: jest.fn() },
  ApiError: jest.requireActual('../api/client').ApiError,
}));

const mockGetProduct = api.getProduct as jest.Mock;
const mockCreateClaim = api.createClaim as jest.Mock;

function product(overrides: Partial<Product> = {}): Product {
  return {
    id: 'p1',
    household_id: 'h1',
    name: 'מזגן טורנדו',
    category: 'מזגן',
    brand: 'טורנדו',
    purchase_date: '2026-01-01',
    price: null,
    room: '',
    warranty_expires_at: '2028-01-01',
    warranty_uncertain: false,
    photo_url: '',
    receipt_id: null,
    created_at: '2026-01-01',
    updated_at: '2026-01-01',
    ...overrides,
  };
}

beforeEach(() => {
  mockGetProduct.mockResolvedValue(product());
  jest.spyOn(Alert, 'alert').mockImplementation(() => {});
  jest.spyOn(Linking, 'openURL').mockResolvedValue(true as any);
});

afterEach(() => {
  jest.restoreAllMocks();
});

function renderScreen(productId = 'p1') {
  const navigation = createMockNavigation();
  render(<ClaimScreen navigation={navigation as any} route={createMockRoute({ productId }) as any} />);
  return navigation;
}

describe('ClaimScreen', () => {
  it('shows the product name and an in-warranty status once loaded', async () => {
    renderScreen();
    expect(await screen.findByText(/מזגן טורנדו — באחריות ✓/)).toBeTruthy();
  });

  it('shows an expired status for a product past its warranty', async () => {
    mockGetProduct.mockResolvedValue(product({ warranty_expires_at: '2020-01-01' }));
    renderScreen();
    expect(await screen.findByText(/מזגן טורנדו — האחריות פגה/)).toBeTruthy();
  });

  it('shows manufacturer contact info for a known brand', async () => {
    renderScreen();
    expect(await screen.findByText(/שירות לקוחות טורנדו: 1-700-505-105/)).toBeTruthy();
    expect(screen.getByText(/טופס תביעה מקוון/)).toBeTruthy();
  });

  it('falls back to a generic message for an unlisted brand', async () => {
    mockGetProduct.mockResolvedValue(product({ brand: 'מותג לא מוכר' }));
    renderScreen();
    expect(await screen.findByText(/אין לנו פרטי קשר ליצרן זה עדיין/)).toBeTruthy();
  });

  it('calls Linking.openURL with the phone number when the contact line is pressed', async () => {
    renderScreen();
    fireEvent.press(await screen.findByText(/שירות לקוחות טורנדו/));
    expect(Linking.openURL).toHaveBeenCalledWith('tel:1-700-505-105');
  });

  it('blocks saving and alerts when the description is empty', async () => {
    renderScreen();
    fireEvent.press(await screen.findByText('שמור תיעוד תקלה'));

    expect(Alert.alert).toHaveBeenCalledWith('נא לתאר את התקלה');
    expect(mockCreateClaim).not.toHaveBeenCalled();
  });

  it('saves the trimmed description and goes back on success', async () => {
    mockCreateClaim.mockResolvedValue({ id: 'c1' });
    const navigation = renderScreen();

    fireEvent.changeText(await screen.findByPlaceholderText('תארו את התקלה...'), '  לא נדלק  ');
    fireEvent.press(screen.getByText('שמור תיעוד תקלה'));

    await waitFor(() => expect(mockCreateClaim).toHaveBeenCalledWith('p1', 'לא נדלק'));
    await waitFor(() => expect(navigation.goBack).toHaveBeenCalled());
  });

  it('shows the server error message when saving fails', async () => {
    mockCreateClaim.mockRejectedValue(new ApiError(500, 'שמירה נכשלה'));
    renderScreen();

    fireEvent.changeText(await screen.findByPlaceholderText('תארו את התקלה...'), 'לא נדלק');
    fireEvent.press(screen.getByText('שמור תיעוד תקלה'));

    await waitFor(() => expect(Alert.alert).toHaveBeenCalledWith('שגיאה', 'שמירה נכשלה'));
  });
});
