import { fireEvent, render, screen } from '@testing-library/react-native';

import DashboardScreen from './DashboardScreen';
import { api } from '../api/client';
import type { Product } from '../api/types';
import { useAuth } from '../context/AuthContext';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('@react-navigation/native', () => ({
  useFocusEffect: (callback: () => void) => {
    // eslint-disable-next-line @typescript-eslint/no-require-imports -- jest.mock factories can't reference out-of-scope imports
    const React = require('react');
    // eslint-disable-next-line react-hooks/exhaustive-deps -- this is a test mock, not a real component effect
    React.useEffect(callback, []);
  },
}));
jest.mock('../context/AuthContext', () => ({ useAuth: jest.fn() }));
jest.mock('../api/client', () => ({ api: { listProducts: jest.fn() } }));

const mockUseAuth = useAuth as jest.Mock;
const mockListProducts = api.listProducts as jest.Mock;

function product(overrides: Partial<Product>): Product {
  return {
    id: 'p1',
    household_id: 'h1',
    name: 'מוצר',
    category: 'x',
    brand: '',
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
  mockUseAuth.mockReturnValue({ user: { full_name: 'מיכל כהן' } });
  mockListProducts.mockResolvedValue([]);
});

describe('DashboardScreen', () => {
  it('greets the user by their first name', () => {
    render(<DashboardScreen navigation={createMockNavigation() as any} route={{} as any} />);
    expect(screen.getByText('שלום, מיכל 👋')).toBeTruthy();
  });

  it('shows an empty-state message once loading finishes with no products', async () => {
    render(<DashboardScreen navigation={createMockNavigation() as any} route={{} as any} />);
    expect(
      await screen.findByText('עדיין אין מוצרים. הוסיפו את הראשון עם הכפתור למטה!'),
    ).toBeTruthy();
  });

  it('lists returned products by name', async () => {
    mockListProducts.mockResolvedValue([
      product({ id: 'p1', name: 'מזגן טורנדו' }),
      product({ id: 'p2', name: 'אוזניות JBL' }),
    ]);
    render(<DashboardScreen navigation={createMockNavigation() as any} route={{} as any} />);

    expect(await screen.findByText('מזגן טורנדו')).toBeTruthy();
    expect(screen.getByText('אוזניות JBL')).toBeTruthy();
  });

  it('the "קרוב לתפוגה" filter hides products that are still safely in warranty', async () => {
    mockListProducts.mockResolvedValue([
      product({ id: 'p1', name: 'בטוח', warranty_expires_at: '2030-01-01' }),
      product({ id: 'p2', name: 'פג', warranty_expires_at: '2020-01-01' }),
    ]);
    render(<DashboardScreen navigation={createMockNavigation() as any} route={{} as any} />);
    await screen.findByText('בטוח');

    fireEvent.press(screen.getByText('קרוב לתפוגה'));

    expect(screen.getByText('פג')).toBeTruthy();
    expect(screen.queryByText('בטוח')).toBeNull();
  });

  it('navigates to AddProductChoose when the + button is pressed', () => {
    const navigation = createMockNavigation();
    render(<DashboardScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('+'));
    expect(navigation.navigate).toHaveBeenCalledWith('AddProductChoose');
  });

  it('navigates to Search and Settings from the header icons', () => {
    const navigation = createMockNavigation();
    render(<DashboardScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('🔍'));
    expect(navigation.navigate).toHaveBeenCalledWith('Search');

    fireEvent.press(screen.getByText('⚙️'));
    expect(navigation.navigate).toHaveBeenCalledWith('Settings');
  });

  it('navigates to ProductDetail with the right id when a product is tapped', async () => {
    mockListProducts.mockResolvedValue([product({ id: 'p42', name: 'מוצר לבדיקה' })]);
    const navigation = createMockNavigation();
    render(<DashboardScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(await screen.findByText('מוצר לבדיקה'));
    expect(navigation.navigate).toHaveBeenCalledWith('ProductDetail', { productId: 'p42' });
  });
});
