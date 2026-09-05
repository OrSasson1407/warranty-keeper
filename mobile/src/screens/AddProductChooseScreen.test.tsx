import { Alert } from 'react-native';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';
import * as ImagePicker from 'expo-image-picker';

import AddProductChooseScreen from './AddProductChooseScreen';
import { api, ApiError } from '../api/client';
import { createMockNavigation } from '../testUtils/navigation';

jest.mock('expo-image-picker');
jest.mock('../api/client', () => ({
  api: { uploadReceipt: jest.fn() },
  ApiError: jest.requireActual('../api/client').ApiError,
}));

const mockRequestPermissions = ImagePicker.requestCameraPermissionsAsync as jest.Mock;
const mockLaunchCamera = ImagePicker.launchCameraAsync as jest.Mock;
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

    await waitFor(() => expect(Alert.alert).toHaveBeenCalledWith('נדרשת הרשאת מצלמה', expect.any(String)));
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

    await waitFor(() => expect(navigation.navigate).toHaveBeenCalledWith('ConfirmProduct', { draft }));
    expect(mockUploadReceipt).toHaveBeenCalledWith({ uri: 'file:///x.jpg', name: 'x.jpg', type: 'image/jpeg' });
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
