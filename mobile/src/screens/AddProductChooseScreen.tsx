import { useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Platform,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import * as ImagePicker from 'expo-image-picker';
import * as DocumentPicker from 'expo-document-picker';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api, ApiError } from '../api/client';
import { colors } from '../theme/colors';
import { fonts, typography } from '../theme/typography';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'AddProductChoose'>;

type Source = 'camera' | 'gallery' | 'file';

export default function AddProductChooseScreen({ navigation }: Props) {
  const [uploading, setUploading] = useState(false);

  const uploadAsset = async (asset: { uri: string; name?: string | null; mimeType?: string | null }) => {
    setUploading(true);
    try {
      const draft = await api.uploadReceipt({
        uri: asset.uri,
        name: asset.name ?? 'receipt.jpg',
        type: asset.mimeType ?? 'image/jpeg',
      });
      navigation.navigate('ConfirmProduct', { draft });
    } catch (e) {
      Alert.alert('שגיאה', e instanceof ApiError ? e.message : 'לא הצלחנו לעבד את הקבלה');
    } finally {
      setUploading(false);
    }
  };

  const pickAndUpload = async (source: Source) => {
    try {
      if (source === 'file') {
        const result = await DocumentPicker.getDocumentAsync({
          type: ['application/pdf', 'image/*'],
        });
        if (result.canceled || !result.assets?.length) return;
        const asset = result.assets[0];
        await uploadAsset({ uri: asset.uri, name: asset.name, mimeType: asset.mimeType });
        return;
      }

      let result: ImagePicker.ImagePickerResult;
      if (source === 'camera' && Platform.OS !== 'web') {
        const perm = await ImagePicker.requestCameraPermissionsAsync();
        if (!perm.granted) {
          Alert.alert('נדרשת הרשאת מצלמה', 'כדי לצלם קבלה יש לאשר גישה למצלמה בהגדרות המכשיר.');
          return;
        }
        result = await ImagePicker.launchCameraAsync({ quality: 0.7 });
      } else {
        result = await ImagePicker.launchImageLibraryAsync({ quality: 0.7 });
      }

      if (result.canceled || !result.assets?.length) return;
      const asset = result.assets[0];
      await uploadAsset({ uri: asset.uri, name: asset.fileName, mimeType: asset.mimeType });
    } catch (e) {
      Alert.alert('שגיאה', e instanceof ApiError ? e.message : 'לא הצלחנו לעבד את הקבלה');
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.heading}>בוא נתחיל!</Text>

      <TouchableOpacity
        style={styles.captureButton}
        onPress={() => pickAndUpload('camera')}
        disabled={uploading}
      >
        {uploading ? (
          <ActivityIndicator color={colors.primaryText} size="large" />
        ) : (
          <>
            <Text style={styles.captureEmoji}>📷</Text>
            <Text style={styles.captureText}>צלם קבלה</Text>
          </>
        )}
      </TouchableOpacity>

      <TouchableOpacity
        style={styles.galleryButton}
        onPress={() => pickAndUpload('gallery')}
        disabled={uploading}
      >
        <Text style={styles.galleryEmoji}>🖼️</Text>
        <Text style={styles.galleryText}>בחר תמונה מהגלריה</Text>
      </TouchableOpacity>

      <TouchableOpacity
        style={styles.galleryButton}
        onPress={() => pickAndUpload('file')}
        disabled={uploading}
      >
        <Text style={styles.galleryEmoji}>📄</Text>
        <Text style={styles.galleryText}>בחר קובץ (PDF / תמונה)</Text>
      </TouchableOpacity>

      <TouchableOpacity
        onPress={() => navigation.navigate('ConfirmProduct', {})}
        disabled={uploading}
      >
        <Text style={styles.manualLink}>הזן ידנית במקום זאת</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
    padding: 24,
    justifyContent: 'center',
    gap: 16,
  },
  heading: {
    ...typography.headlineLg,
    color: colors.text,
    textAlign: 'center',
    marginBottom: 8,
  },
  captureButton: {
    backgroundColor: colors.primary,
    paddingVertical: 48,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
  },
  captureEmoji: { fontSize: 40 },
  captureText: { color: colors.primaryText, ...typography.headlineSm, fontFamily: fonts.headlineSm },
  galleryButton: {
    backgroundColor: colors.surfaceContainer,
    borderWidth: 1,
    borderColor: colors.border,
    paddingVertical: 20,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
  },
  galleryEmoji: { fontSize: 24 },
  galleryText: { color: colors.text, ...typography.bodyLg, fontFamily: fonts.bodyMdSemiBold },
  manualLink: { color: colors.textMuted, textAlign: 'center', ...typography.bodyMd, marginTop: 8 },
});
