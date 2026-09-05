import { Alert } from 'react-native';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import ConfirmProductScreen from './ConfirmProductScreen';
import { api, ApiError } from '../api/client';
import { createMockNavigation, createMockRoute } from '../testUtils/navigation';
import type { Product, ReceiptDraft } from '../api/types';

jest.mock('../api/client', () => ({
  api: { resolveWarranty: jest.fn(), createProduct: jest.fn() },
  ApiError: jest.requireActual('../api/client').ApiError,
}));

const mockResolveWarranty = api.resolveWarranty as jest.Mock;
const mockCreateProduct = api.createProduct as jest.Mock;

function draft(overrides: Partial<ReceiptDraft> = {}): ReceiptDraft {
  return {
    receipt_id: 'r1',
    image_url: 'https://example.test/r1.jpg',
    status: 'processed',
    parsed_vendor: 'טורנדו',
    parsed_date: '2026-01-15',
    parsed_amount: 3200,
    raw_ocr_text: '',
    confidence: 0.9,
    suggested_category: 'מזגן',
    warranty_expires_at: '2028-01-15',
    warranty_uncertain: false,
    ...overrides,
  };
}

function product(overrides: Partial<Product> = {}): Product {
  return {
    id: 'p1',
    household_id: 'h1',
    name: 'X',
    category: 'מזגן',
    brand: '',
    purchase_date: '2026-01-15',
    price: null,
    room: '',
    warranty_expires_at: '2028-01-15',
    warranty_uncertain: false,
    photo_url: '',
    receipt_id: null,
    created_at: '2026-01-15',
    updated_at: '2026-01-15',
    ...overrides,
  };
}

beforeEach(() => {
  mockResolveWarranty.mockResolvedValue({
    warranty_expires_at: '2028-01-15',
    duration_months: 24,
    uncertain: false,
    source: 'default',
  });
  jest.spyOn(Alert, 'alert').mockImplementation(() => {});
});

afterEach(() => {
  jest.restoreAllMocks();
});

function renderManual() {
  const navigation = createMockNavigation();
  render(<ConfirmProductScreen navigation={navigation as any} route={createMockRoute({}) as any} />);
  return navigation;
}

function renderWithDraft(d: ReceiptDraft) {
  const navigation = createMockNavigation();
  render(<ConfirmProductScreen navigation={navigation as any} route={createMockRoute({ draft: d }) as any} />);
  return navigation;
}

describe('ConfirmProductScreen', () => {
  it('shows the manual-entry heading and an empty name field with no draft', () => {
    renderManual();
    expect(screen.getByText('הזנה ידנית')).toBeTruthy();
    expect(screen.getByPlaceholderText('לדוגמה: מזגן טורנדו').props.value).toBe('');
  });

  it('pre-fills fields from a receipt draft', () => {
    renderWithDraft(draft());
    expect(screen.getByText('מצאנו את זה:')).toBeTruthy();
    expect(screen.getByPlaceholderText('לדוגמה: מזגן טורנדו').props.value).toBe('רכישה מטורנדו');
    expect(screen.getByText('מזגן')).toBeTruthy(); // category field showing the suggested category
  });

  it('shows a low-confidence warning when the draft OCR confidence is under 0.5', () => {
    renderWithDraft(draft({ confidence: 0.2 }));
    expect(screen.getByText('לא הצלחנו לזהות את כל הפרטים אוטומטית — נא להשלים ולוודא ידנית.')).toBeTruthy();
  });

  it('does not show the low-confidence warning for a confident draft', () => {
    renderWithDraft(draft({ confidence: 0.9 }));
    expect(
      screen.queryByText('לא הצלחנו לזהות את כל הפרטים אוטומטית — נא להשלים ולוודא ידנית.')
    ).toBeNull();
  });

  it('re-resolves and displays the warranty estimate when a category is picked', async () => {
    mockResolveWarranty.mockResolvedValue({
      warranty_expires_at: '2030-06-01',
      duration_months: 24,
      uncertain: false,
      source: 'default',
    });
    renderManual();

    // Category SelectField is the first "בחר..." trigger (before the room field's).
    fireEvent.press(screen.getAllByText('בחר...')[0]);
    fireEvent.press(screen.getByText('מזגן'));

    await waitFor(() => expect(mockResolveWarranty).toHaveBeenCalled());
    expect(await screen.findByText(/אחריות עד:/)).toBeTruthy();
  });

  it('flags the estimate as approximate when the resolver reports uncertain', async () => {
    mockResolveWarranty.mockResolvedValue({
      warranty_expires_at: '2027-01-01',
      duration_months: 12,
      uncertain: true,
      source: 'fallback',
    });
    renderManual();

    // Pick a category within FlatList's default render window (first ~10 items).
    fireEvent.press(screen.getAllByText('בחר...')[0]);
    fireEvent.press(screen.getByText('מזגן'));

    await waitFor(() => expect(screen.getByText(/\(משוער\)/)).toBeTruthy());
    expect(
      screen.getByText('לא מצאנו כלל מדויק לקטגוריה הזו — ברירת מחדל של 12 חודשים. ניתן לערוך ידנית בהמשך.')
    ).toBeTruthy();
  });

  it('blocks saving and alerts when name or category is missing', () => {
    renderManual();
    fireEvent.press(screen.getByText('שמור מוצר'));

    expect(Alert.alert).toHaveBeenCalledWith('חסרים פרטים', expect.any(String));
    expect(mockCreateProduct).not.toHaveBeenCalled();
  });

  it('saves the product and navigates to ProductDetail on success', async () => {
    mockCreateProduct.mockResolvedValue(product({ id: 'new-id' }));
    const navigation = renderWithDraft(draft());

    fireEvent.press(screen.getByText('שמור מוצר'));

    await waitFor(() =>
      expect(navigation.replace).toHaveBeenCalledWith('ProductDetail', { productId: 'new-id' })
    );
    const payload = mockCreateProduct.mock.calls[0][0];
    expect(payload.name).toBe('רכישה מטורנדו');
    expect(payload.category).toBe('מזגן');
    expect(payload.receipt_id).toBe('r1');
    expect(payload.photo_url).toBe('https://example.test/r1.jpg');
  });

  it('shows the server error message when saving fails', async () => {
    mockCreateProduct.mockRejectedValue(new ApiError(500, 'שמירה נכשלה בשרת'));
    renderWithDraft(draft());

    fireEvent.press(screen.getByText('שמור מוצר'));

    await waitFor(() => expect(Alert.alert).toHaveBeenCalledWith('שגיאה בשמירה', 'שמירה נכשלה בשרת'));
  });
});
