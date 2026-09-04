import { useState } from 'react';
import { ActivityIndicator, Alert, Platform, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import * as ImagePicker from 'expo-image-picker';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api, ApiError } from '../api/client';
import { colors } from '../theme/colors';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'AddProductChoose'>;

export default function AddProductChooseScreen({ navigation }: Props) {
  const [uploading, setUploading] = useState(false);

  const pickAndUpload = async (fromCamera: boolean) => {
    try {
      let result: ImagePicker.ImagePickerResult;
      if (fromCamera && Platform.OS !== 'web') {
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
      setUploading(true);

      const draft = await api.uploadReceipt({
        uri: asset.uri,
        name: asset.fileName ?? 'receipt.jpg',
        type: asset.mimeType ?? 'image/jpeg',
      });

      navigation.navigate('ConfirmProduct', { draft });
    } catch (e) {
      Alert.alert('שגיאה', e instanceof ApiError ? e.message : 'לא הצלחנו לעבד את הקבלה');
    } finally {
      setUploading(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.heading}>בוא נתחיל!</Text>

      <TouchableOpacity
        style={styles.captureButton}
        onPress={() => pickAndUpload(true)}
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
        onPress={() => navigation.navigate('ConfirmProduct', {})}
        disabled={uploading}
      >
        <Text style={styles.manualLink}>הזן ידנית במקום זאת</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background, padding: 24, justifyContent: 'center', gap: 24 },
  heading: { fontSize: 22, fontWeight: '700', color: colors.text, textAlign: 'center', marginBottom: 8 },
  captureButton: {
    backgroundColor: colors.primary,
    borderRadius: 20,
    paddingVertical: 48,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
  },
  captureEmoji: { fontSize: 40 },
  captureText: { color: colors.primaryText, fontSize: 18, fontWeight: '600' },
  manualLink: { color: colors.textMuted, textAlign: 'center', fontSize: 15 },
});
