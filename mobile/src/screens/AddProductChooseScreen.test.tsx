import { Alert } from 'react-native';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';
import * as ImagePicker from 'expo-image-picker';
import * as DocumentPicker from 'expo-document-picker';

import AddProductChooseScreen from './AddProductChooseScreen';
import { api, ApiError } from '../api/client';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('expo-image-picker');
jest.mock('expo-document-picker');
jest.mock('../api/client', () => ({
  api: { uploadReceipt: jest.fn() },
  ApiError: jest.requireActual('../api/client').ApiError,
}));

const mockRequestPermissions = ImagePicker.requestCameraPermissionsAsync as jest.Mock;
const mockLaunchCamera = ImagePicker.launchCameraAsync as jest.Mock;
const mockLaunchLibrary = ImagePicker.launchImageLibraryAsync as jest.Mock;
const mockGetDocument = DocumentPicker.getDocumentAsync as jest.Mock;
const mockUploadReceipt = api.uploadReceipt as jest.Mock;

beforeEach(() => {
  mockRequestPermissions.mockResolvedValue({ granted: true });
  jest.spyOn(Alert, 'alert').mockImplementation(() => {});
});

afterEach(() => {
  jest.restoreAllMocks();
});

describe('AddProductChooseScreen', () => {
  it('navigates straight to ConfirmProduct with no draft when "הזן ידנית" is pressed', () => {
    const navigation = createMockNavigation();
    render(<AddProductChooseScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('הזן ידנית במקום זאת'));
    expect(navigation.navigate).toHaveBeenCalledWith('ConfirmProduct', {});
  });

  it('alerts and does not proceed when camera permission is denied', async () => {
    mockRequestPermissions.mockResolvedValue({ granted: false });
    const navigation = createMockNavigation();
    render(<AddProductChooseScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('צלם קבלה'));

    await waitFor(() =>
      expect(Alert.alert).toHaveBeenCalledWith('נדרשת הרשאת מצלמה', expect.any(String)),
    );
    expect(navigation.navigate).not.toHaveBeenCalled();
    expect(mockUploadReceipt).not.toHaveBeenCalled();
  });

  it('does nothing when the user cancels the camera', async () => {
    mockLaunchCamera.mockResolvedValue({ canceled: true, assets: null });
    const navigation = createMockNavigation();
    render(<AddProductChooseScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('צלם קבלה'));

    await waitFor(() => expect(mockLaunchCamera).toHaveBeenCalled());
    expect(mockUploadReceipt).not.toHaveBeenCalled();
    expect(navigation.navigate).not.toHaveBeenCalled();
  });

  it('uploads the captured photo and navigates to ConfirmProduct with the draft', async () => {
    mockLaunchCamera.mockResolvedValue({
      canceled: false,
      assets: [{ uri: 'file:///x.jpg', fileName: 'x.jpg', mimeType: 'image/jpeg' }],
    });
    const draft = { receipt_id: 'r1', suggested_category: 'מזגן' };
    mockUploadReceipt.mockResolvedValue(draft);

    const navigation = createMockNavigation();
    render(<AddProductChooseScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('צלם קבלה'));

    await waitFor(() =>
      expect(navigation.navigate).toHaveBeenCalledWith('ConfirmProduct', { draft }),
    );
    expect(mockUploadReceipt).toHaveBeenCalledWith({
      uri: 'file:///x.jpg',
      name: 'x.jpg',
      type: 'image/jpeg',
    });
  });

  it('uploads a photo chosen from the gallery without requesting camera permission', async () => {
    mockLaunchLibrary.mockResolvedValue({
      canceled: false,
      assets: [{ uri: 'file:///y.jpg', fileName: 'y.jpg', mimeType: 'image/jpeg' }],
    });
    const draft = { receipt_id: 'r2', suggested_category: 'טלוויזיה' };
    mockUploadReceipt.mockResolvedValue(draft);

    const navigation = createMockNavigation();
    render(<AddProductChooseScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('בחר תמונה מהגלריה'));

    await waitFor(() =>
      expect(navigation.navigate).toHaveBeenCalledWith('ConfirmProduct', { draft }),
    );
    expect(mockUploadReceipt).toHaveBeenCalledWith({
      uri: 'file:///y.jpg',
      name: 'y.jpg',
      type: 'image/jpeg',
    });
  });

  it('uploads a PDF chosen from the file picker', async () => {
    mockGetDocument.mockResolvedValue({
      canceled: false,
      assets: [{ uri: 'file:///receipt.pdf', name: 'receipt.pdf', mimeType: 'application/pdf' }],
    });
    const draft = { receipt_id: 'r3', suggested_category: 'מזגן' };
    mockUploadReceipt.mockResolvedValue(draft);

    const navigation = createMockNavigation();
    render(<AddProductChooseScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('בחר קובץ (PDF / תמונה)'));

    await waitFor(() =>
      expect(navigation.navigate).toHaveBeenCalledWith('ConfirmProduct', { draft }),
    );
    expect(mockUploadReceipt).toHaveBeenCalledWith({
      uri: 'file:///receipt.pdf',
      name: 'receipt.pdf',
      type: 'application/pdf',
    });
  });

  it('does nothing when the user cancels the file picker', async () => {
    mockGetDocument.mockResolvedValue({ canceled: true, assets: null });
    render(<AddProductChooseScreen navigation={createMockNavigation() as any} route={{} as any} />);

    fireEvent.press(screen.getByText('בחר קובץ (PDF / תמונה)'));

    await waitFor(() => expect(mockGetDocument).toHaveBeenCalled());
  });

  it('shows the server error message when the upload fails', async () => {
    mockLaunchCamera.mockResolvedValue({
      canceled: false,
      assets: [{ uri: 'file:///x.jpg', fileName: 'x.jpg', mimeType: 'image/jpeg' }],
    });
    mockUploadReceipt.mockRejectedValue(new ApiError(500, 'השרת נכשל'));

    render(<AddProductChooseScreen navigation={createMockNavigation() as any} route={{} as any} />);
    fireEvent.press(screen.getByText('צלם קבלה'));

    await waitFor(() => expect(Alert.alert).toHaveBeenCalledWith('שגיאה', 'השרת נכשל'));
  });
});
