import { act, fireEvent, render, screen } from '@testing-library/react-native';

import SearchScreen from './SearchScreen';
import { api } from '../api/client';
import type { Product } from '../api/types';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('../api/client', () => ({ api: { listProducts: jest.fn() } }));

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
  jest.useFakeTimers();
  mockListProducts.mockResolvedValue([]);
});

afterEach(() => {
  jest.useRealTimers();
});

async function typeAndDebounce(text: string) {
  fireEvent.changeText(screen.getByPlaceholderText('🔍 חפש מוצר...'), text);
  await act(async () => {
    await jest.advanceTimersByTimeAsync(300);
  });
}

describe('SearchScreen', () => {
  it('does not search or show a results label for an empty query', () => {
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);
    expect(mockListProducts).not.toHaveBeenCalled();
    expect(screen.queryByText(/תוצאות/)).toBeNull();
  });

  it('debounces: only searches once, 300ms after typing stops', async () => {
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);
    const input = screen.getByPlaceholderText('🔍 חפש מוצר...');

    fireEvent.changeText(input, 'מ');
    fireEvent.changeText(input, 'מז');
    fireEvent.changeText(input, 'מזג');

    await act(async () => {
      await jest.advanceTimersByTimeAsync(300);
    });

    expect(mockListProducts).toHaveBeenCalledTimes(1);
    expect(mockListProducts).toHaveBeenCalledWith(expect.objectContaining({ q: 'מזג' }));
  });

  it('shows the result count and matching product names', async () => {
    mockListProducts.mockResolvedValue([
      product({ id: 'p1', name: 'מזגן טורנדו' }),
      product({ id: 'p2', name: 'מזגן אלקטרה' }),
    ]);
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);

    await typeAndDebounce('מזגן');

    expect(screen.getByText('תוצאות (2):')).toBeTruthy();
    expect(screen.getByText('מזגן טורנדו')).toBeTruthy();
    expect(screen.getByText('מזגן אלקטרה')).toBeTruthy();
  });

  it('shows a zero count when nothing matches', async () => {
    mockListProducts.mockResolvedValue([]);
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);

    await typeAndDebounce('nonexistent');

    expect(screen.getByText('תוצאות (0):')).toBeTruthy();
  });

  it('navigates to ProductDetail when a result is tapped', async () => {
    mockListProducts.mockResolvedValue([product({ id: 'p42', name: 'מוצר לבדיקה' })]);
    const navigation = createMockNavigation();
    render(<SearchScreen navigation={navigation as any} route={{} as any} />);

    await typeAndDebounce('בדיקה');
    fireEvent.press(screen.getByText('מוצר לבדיקה'));

    expect(navigation.navigate).toHaveBeenCalledWith('ProductDetail', { productId: 'p42' });
  });

  it('clears results and the label when the query is erased', async () => {
    mockListProducts.mockResolvedValue([product({ id: 'p1', name: 'מזגן' })]);
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);

    await typeAndDebounce('מזגן');
    expect(screen.getByText('מזגן')).toBeTruthy();

    fireEvent.changeText(screen.getByPlaceholderText('🔍 חפש מוצר...'), '');
    await act(async () => {
      await jest.advanceTimersByTimeAsync(300);
    });

    expect(screen.queryByText('מזגן')).toBeNull();
    expect(screen.queryByText(/תוצאות/)).toBeNull();
  });

  it('hides the filter panel by default and shows it when the filter toggle is pressed', () => {
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);
    expect(screen.queryByText('קטגוריה')).toBeNull();

    fireEvent.press(screen.getByText('🔧 סינון'));
    expect(screen.getByText('קטגוריה')).toBeTruthy();
    expect(screen.getByText('חדר')).toBeTruthy();
  });

  it('triggers a search from a status filter alone, with no text query', async () => {
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);
    fireEvent.press(screen.getByText('🔧 סינון'));

    fireEvent.press(screen.getByText('פג תוקף'));
    await act(async () => {
      await jest.advanceTimersByTimeAsync(300);
    });

    expect(mockListProducts).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'expired', q: undefined }),
    );
  });

  it('sends price_min/price_max as numbers', async () => {
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);
    fireEvent.press(screen.getByText('🔧 סינון'));

    fireEvent.changeText(screen.getByPlaceholderText('מחיר מינימלי'), '100');
    fireEvent.changeText(screen.getByPlaceholderText('מחיר מקסימלי'), '500');
    await act(async () => {
      await jest.advanceTimersByTimeAsync(300);
    });

    expect(mockListProducts).toHaveBeenCalledWith(
      expect.objectContaining({ price_min: 100, price_max: 500 }),
    );
  });

  it('combines the text query with active filters in one call', async () => {
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);
    fireEvent.press(screen.getByText('🔧 סינון'));
    fireEvent.press(screen.getByText('עומד לפוג'));

    await typeAndDebounce('מזגן');

    expect(mockListProducts).toHaveBeenCalledWith(
      expect.objectContaining({ q: 'מזגן', status: 'warning' }),
    );
  });

  it('clears results when both the query and all filters are cleared', async () => {
    mockListProducts.mockResolvedValue([product({ id: 'p1', name: 'מזגן' })]);
    render(<SearchScreen navigation={createMockNavigation() as any} route={{} as any} />);
    fireEvent.press(screen.getByText('🔧 סינון'));
    fireEvent.press(screen.getByText('פג תוקף'));
    await act(async () => {
      await jest.advanceTimersByTimeAsync(300);
    });
    expect(screen.getByText('מזגן')).toBeTruthy();

    fireEvent.press(screen.getByText('כל הסטטוסים'));
    await act(async () => {
      await jest.advanceTimersByTimeAsync(300);
    });

    expect(screen.queryByText('מזגן')).toBeNull();
    expect(screen.queryByText(/תוצאות/)).toBeNull();
  });
});
