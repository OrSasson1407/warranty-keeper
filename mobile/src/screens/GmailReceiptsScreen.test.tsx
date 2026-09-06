import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';

import GmailReceiptsScreen from './GmailReceiptsScreen';
import { api } from '../api/client';
import type { ReceiptDraft } from '../api/types';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('../api/client', () => ({
  api: { listReceipts: jest.fn() },
}));

const mockListReceipts = api.listReceipts as jest.Mock;

function draft(overrides: Partial<ReceiptDraft> = {}): ReceiptDraft {
  return {
    receipt_id: 'r1',
    image_url: '',
    status: 'pending',
    parsed_vendor: 'Amazon',
    parsed_date: '2026-03-01',
    parsed_amount: 199.9,
    raw_ocr_text: 'Order confirmed',
    confidence: 0.45,
    suggested_category: '',
    warranty_expires_at: '2027-03-01',
    warranty_uncertain: true,
    ...overrides,
  };
}

function renderScreen() {
  const navigation = createMockNavigation();
  render(<GmailReceiptsScreen navigation={navigation as any} route={{} as any} />);
  return navigation;
}

beforeEach(() => {
  mockListReceipts.mockResolvedValue([]);
});

afterEach(() => {
  jest.clearAllMocks();
});

describe('GmailReceiptsScreen', () => {
  it('requests only pending gmail-sourced receipts', async () => {
    renderScreen();
    await waitFor(() =>
      expect(mockListReceipts).toHaveBeenCalledWith({ status: 'pending', source: 'gmail' }),
    );
  });

  it('shows an empty state when there are no pending receipts', async () => {
    renderScreen();
    expect(await screen.findByText(/לא נמצאו קבלות חדשות מ-Gmail/)).toBeTruthy();
  });

  it('lists each pending receipt with vendor, date, and amount', async () => {
    mockListReceipts.mockResolvedValue([draft()]);
    renderScreen();
    expect(await screen.findByText('Amazon')).toBeTruthy();
    expect(screen.getByText('2026-03-01')).toBeTruthy();
    expect(screen.getByText('₪199.9')).toBeTruthy();
  });

  it('shows a low-confidence hint for uncertain drafts', async () => {
    mockListReceipts.mockResolvedValue([draft({ confidence: 0.3 })]);
    renderScreen();
    expect(await screen.findByText('נדרש אימות ידני של הפרטים')).toBeTruthy();
  });

  it('hides the low-confidence hint for confident drafts', async () => {
    mockListReceipts.mockResolvedValue([draft({ confidence: 0.9 })]);
    renderScreen();
    await screen.findByText('Amazon');
    expect(screen.queryByText('נדרש אימות ידני של הפרטים')).toBeNull();
  });

  it('navigates to ConfirmProduct with the draft when a receipt is pressed', async () => {
    const d = draft();
    mockListReceipts.mockResolvedValue([d]);
    const navigation = renderScreen();

    fireEvent.press(await screen.findByText('Amazon'));

    expect(navigation.navigate).toHaveBeenCalledWith('ConfirmProduct', { draft: d });
  });
});
