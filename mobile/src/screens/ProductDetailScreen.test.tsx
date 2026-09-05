import { Linking } from 'react-native';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import ProductDetailScreen from './ProductDetailScreen';
import { api } from '../api/client';
import type { Product, WarrantyClaim } from '../api/types';
import { createMockNavigation, createMockRoute } from '../testUtils/navigation';

jest.mock('@react-navigation/native', () => ({
  useFocusEffect: (callback: () => void) => {
    // eslint-disable-next-line @typescript-eslint/no-require-imports -- jest.mock factories can't reference out-of-scope imports
    const React = require('react');
    // eslint-disable-next-line react-hooks/exhaustive-deps -- this is a test mock, not a real component effect
    React.useEffect(callback, []);
  },
}));
jest.mock('../api/client', () => ({
  api: { getProduct: jest.fn(), listClaims: jest.fn() },
}));

const mockGetProduct = api.getProduct as jest.Mock;
const mockListClaims = api.listClaims as jest.Mock;

function product(overrides: Partial<Product> = {}): Product {
  return {
    id: 'p1',
    household_id: 'h1',
    name: 'מזגן טורנדו',
    category: 'מזגן',
    brand: 'טורנדו',
    purchase_date: '2026-01-01',
    price: 3200,
    room: 'סלון',
    warranty_expires_at: '2028-01-01',
    warranty_uncertain: false,
    photo_url: 'https://example.test/p1.jpg',
    receipt_id: 'r1',
    created_at: '2026-01-01',
    updated_at: '2026-01-01',
    ...overrides,
  };
}

function claim(overrides: Partial<WarrantyClaim> = {}): WarrantyClaim {
  return {
    id: 'c1',
    product_id: 'p1',
    issue_description: 'לא נדלק',
    status: 'open',
    created_at: '2026-02-01',
    resolved_at: null,
    ...overrides,
  };
}

beforeEach(() => {
  mockGetProduct.mockResolvedValue(product());
  mockListClaims.mockResolvedValue([]);
  jest.spyOn(Linking, 'openURL').mockResolvedValue(true as any);
});

afterEach(() => {
  jest.restoreAllMocks();
});

function renderScreen(productId = 'p1') {
  const navigation = createMockNavigation();
  render(
    <ProductDetailScreen
      navigation={navigation as any}
      route={createMockRoute({ productId }) as any}
    />,
  );
  return navigation;
}

describe('ProductDetailScreen', () => {
  it('loads and displays the product name and price', async () => {
    renderScreen();
    expect(await screen.findByText('מזגן טורנדו')).toBeTruthy();
    expect(screen.getByText(/₪3,200|₪3200/)).toBeTruthy();
  });

  it('requests the product and its claims using the route param id', () => {
    renderScreen('specific-id');
    expect(mockGetProduct).toHaveBeenCalledWith('specific-id');
    expect(mockListClaims).toHaveBeenCalledWith('specific-id');
  });

  it('shows "אין רשומות" when there are no claims', async () => {
    renderScreen();
    expect(await screen.findByText('אין רשומות')).toBeTruthy();
  });

  it('lists claims with their description and Hebrew status label', async () => {
    mockListClaims.mockResolvedValue([
      claim({ issue_description: 'לא מקרר', status: 'in_progress' }),
    ]);
    renderScreen();

    expect(await screen.findByText('לא מקרר')).toBeTruthy();
    expect(screen.getByText('בטיפול')).toBeTruthy();
  });

  it('shows the uncertain-warranty note only when the product warranty is uncertain', async () => {
    mockGetProduct.mockResolvedValue(product({ warranty_uncertain: true }));
    renderScreen();
    expect(await screen.findByText('תאריך משוער — ייתכן שקיימת אחריות שונה בפועל')).toBeTruthy();
  });

  it('navigates to Claim with the product id when "המוצר התקלקל?" is pressed', async () => {
    const navigation = renderScreen();
    fireEvent.press(await screen.findByText('המוצר התקלקל?'));
    expect(navigation.navigate).toHaveBeenCalledWith('Claim', { productId: 'p1' });
  });

  it('opens the receipt image when "צפה בקבלה" is pressed', async () => {
    renderScreen();
    fireEvent.press(await screen.findByText('🧾 צפה בקבלה'));
    await waitFor(() =>
      expect(Linking.openURL).toHaveBeenCalledWith('https://example.test/p1.jpg'),
    );
  });

  it('hides the receipt button when there is no linked receipt', async () => {
    mockGetProduct.mockResolvedValue(product({ receipt_id: null }));
    renderScreen();
    await screen.findByText('מזגן טורנדו');
    expect(screen.queryByText('🧾 צפה בקבלה')).toBeNull();
  });
});
