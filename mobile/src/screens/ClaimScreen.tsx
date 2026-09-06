import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Linking,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api, ApiError } from '../api/client';
import type { ManufacturerContact, Product } from '../api/types';
import {
  getManufacturerContact,
  loadManufacturerContacts,
  type ManufacturerContactMap,
} from '../data/manufacturerContacts';
import { colors } from '../theme/colors';
import { warrantyStatus } from '../utils/warrantyStatus';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'Claim'>;

export default function ClaimScreen({ navigation, route }: Props) {
  const { productId } = route.params;
  const [product, setProduct] = useState<Product | null>(null);
  const [contacts, setContacts] = useState<ManufacturerContactMap>({});
  const [vendorContact, setVendorContact] = useState<ManufacturerContact | null>(null);
  const [description, setDescription] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    api.getProduct(productId).then(setProduct);
    loadManufacturerContacts().then(setContacts);
  }, [productId]);

  // Fall back to the receipt's parsed vendor when the product's own brand
  // field doesn't match a known contact (e.g. the user never filled it in).
  useEffect(() => {
    if (!product?.receipt_id || getManufacturerContact(contacts, product.brand)) return;
    api
      .getReceipt(product.receipt_id)
      .then((receipt) => {
        if (receipt.parsed_vendor) {
          setVendorContact(getManufacturerContact(contacts, receipt.parsed_vendor));
        }
      })
      .catch(() => {
        /* no receipt data available -- fall through to the generic contact message */
      });
  }, [product, contacts]);

  const onSave = async () => {
    if (!description.trim()) {
      Alert.alert('נא לתאר את התקלה');
      return;
    }
    setSaving(true);
    try {
      await api.createClaim(productId, description.trim());
      navigation.goBack();
    } catch (e) {
      Alert.alert('שגיאה', e instanceof ApiError ? e.message : 'לא הצלחנו לשמור את התיעוד');
    } finally {
      setSaving(false);
    }
  };

  if (!product) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  const inWarranty = warrantyStatus(product.warranty_expires_at) !== 'expired';
  const contact = getManufacturerContact(contacts, product.brand) ?? vendorContact;

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.title}>
        {product.name} — {inWarranty ? 'באחריות ✓' : 'האחריות פגה'}
      </Text>

      <Text style={styles.label}>מה קרה?</Text>
      <TextInput
        style={styles.textarea}
        multiline
        numberOfLines={5}
        value={description}
        onChangeText={setDescription}
        placeholder="תארו את התקלה..."
        textAlignVertical="top"
      />

      <View style={styles.contactBox}>
        <Text style={styles.contactHeading}>איך לממש</Text>
        {contact ? (
          <>
            {contact.phone ? (
              <TouchableOpacity onPress={() => Linking.openURL(`tel:${contact.phone}`)}>
                <Text style={styles.contactLine}>
                  📞 שירות לקוחות {product.brand}: {contact.phone}
                </Text>
              </TouchableOpacity>
            ) : null}
            {contact.website ? (
              <TouchableOpacity onPress={() => Linking.openURL(contact.website!)}>
                <Text style={styles.contactLine}>🌐 טופס תביעה מקוון: {contact.website}</Text>
              </TouchableOpacity>
            ) : null}
          </>
        ) : (
          <Text style={styles.contactLine}>
            אין לנו פרטי קשר ליצרן זה עדיין — מומלץ לפנות לחנות בה בוצעה הרכישה, או לשמור את הקבלה
            ולחפש את שירות הלקוחות של {product.brand || 'היצרן'} באינטרנט.
          </Text>
        )}
      </View>

      <TouchableOpacity style={styles.saveButton} onPress={onSave} disabled={saving}>
        {saving ? (
          <ActivityIndicator color={colors.primaryText} />
        ) : (
          <Text style={styles.saveButtonText}>שמור תיעוד תקלה</Text>
        )}
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  content: { padding: 20, gap: 12 },
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
  title: { fontSize: 18, fontWeight: '700', color: colors.text, textAlign: 'right' },
  label: { fontSize: 15, fontWeight: '600', color: colors.text, textAlign: 'right', marginTop: 8 },
  textarea: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    padding: 12,
    fontSize: 15,
    minHeight: 110,
    textAlign: 'right',
  },
  contactBox: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: 14,
    gap: 8,
    marginTop: 8,
  },
  contactHeading: { fontSize: 14, fontWeight: '700', color: colors.textMuted, textAlign: 'right' },
  contactLine: { fontSize: 14, color: colors.primary, textAlign: 'right', lineHeight: 20 },
  saveButton: {
    backgroundColor: colors.primary,
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    marginTop: 8,
  },
  saveButtonText: { color: colors.primaryText, fontSize: 16, fontWeight: '600' },
});
