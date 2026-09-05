import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Image,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api, ApiError } from '../api/client';
import SelectField from '../components/SelectField';
import { CATEGORIES, ROOMS } from '../data/categories';
import { colors } from '../theme/colors';
import { formatHebrewDate } from '../utils/warrantyStatus';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'ConfirmProduct'>;

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

export default function ConfirmProductScreen({ navigation, route }: Props) {
  const draft = route.params?.draft;

  const [name, setName] = useState(draft?.parsed_vendor ? `רכישה מ${draft.parsed_vendor}` : '');
  const [category, setCategory] = useState(draft?.suggested_category ?? '');
  const [brand, setBrand] = useState('');
  const [purchaseDate, setPurchaseDate] = useState(draft?.parsed_date ?? todayISO());
  const [price, setPrice] = useState(draft?.parsed_amount ? String(draft.parsed_amount) : '');
  const [room, setRoom] = useState('');
  const [warrantyExpiresAt, setWarrantyExpiresAt] = useState(draft?.warranty_expires_at ?? '');
  const [uncertain, setUncertain] = useState(draft?.warranty_uncertain ?? true);
  const [saving, setSaving] = useState(false);
  const [resolving, setResolving] = useState(false);

  useEffect(() => {
    if (!category) return;
    let cancelled = false;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- loading flag for the async call below, not derived state
    setResolving(true);
    api
      .resolveWarranty(category, brand, purchaseDate || todayISO())
      .then((res) => {
        if (cancelled) return;
        setWarrantyExpiresAt(res.warranty_expires_at);
        setUncertain(res.uncertain);
      })
      .catch(() => {
        /* keep previous estimate if the recompute fails */
      })
      .finally(() => !cancelled && setResolving(false));
    return () => {
      cancelled = true;
    };
  }, [category, brand, purchaseDate]);

  const onSave = async () => {
    if (!name.trim() || !category) {
      Alert.alert('חסרים פרטים', 'נא למלא לפחות שם מוצר וקטגוריה');
      return;
    }
    setSaving(true);
    try {
      const product = await api.createProduct({
        name: name.trim(),
        category,
        brand: brand.trim() || undefined,
        purchase_date: purchaseDate || todayISO(),
        price: price ? Number(price) : null,
        room: room || undefined,
        photo_url: draft?.image_url,
        receipt_id: draft?.receipt_id ?? null,
        warranty_expires_at: warrantyExpiresAt || undefined,
      });
      navigation.replace('ProductDetail', { productId: product.id });
    } catch (e) {
      Alert.alert('שגיאה בשמירה', e instanceof ApiError ? e.message : 'נסו שוב');
    } finally {
      setSaving(false);
    }
  };

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.heading}>{draft ? 'מצאנו את זה:' : 'הזנה ידנית'}</Text>

      {draft?.image_url ? (
        <Image source={{ uri: draft.image_url }} style={styles.receiptImage} resizeMode="cover" />
      ) : null}

      {draft && draft.confidence < 0.5 ? (
        <Text style={styles.lowConfidenceNote}>
          לא הצלחנו לזהות את כל הפרטים אוטומטית — נא להשלים ולוודא ידנית.
        </Text>
      ) : null}

      <Field label="שם מוצר">
        <TextInput
          style={styles.input}
          value={name}
          onChangeText={setName}
          placeholder="לדוגמה: מזגן טורנדו"
        />
      </Field>

      <SelectField label="קטגוריה" value={category} options={CATEGORIES} onChange={setCategory} />

      <Field label="מותג (אופציונלי)">
        <TextInput
          style={styles.input}
          value={brand}
          onChangeText={setBrand}
          placeholder="לדוגמה: טורנדו"
        />
      </Field>

      <Field label="תאריך קנייה (YYYY-MM-DD)">
        <TextInput
          style={styles.input}
          value={purchaseDate}
          onChangeText={setPurchaseDate}
          placeholder={todayISO()}
        />
      </Field>

      <Field label="מחיר (₪, אופציונלי)">
        <TextInput
          style={styles.input}
          value={price}
          onChangeText={setPrice}
          keyboardType="numeric"
        />
      </Field>

      <SelectField label="חדר" value={room} options={ROOMS} onChange={setRoom} />

      <View style={styles.warrantyBox}>
        {resolving ? (
          <ActivityIndicator />
        ) : (
          <>
            <Text style={styles.warrantyText}>
              אחריות עד: {warrantyExpiresAt ? formatHebrewDate(warrantyExpiresAt) : '—'}
              {uncertain ? ' (משוער)' : ''}
            </Text>
            {uncertain ? (
              <Text style={styles.warrantyHint}>
                לא מצאנו כלל מדויק לקטגוריה הזו — ברירת מחדל של 12 חודשים. ניתן לערוך ידנית בהמשך.
              </Text>
            ) : null}
          </>
        )}
      </View>

      <TouchableOpacity style={styles.saveButton} onPress={onSave} disabled={saving}>
        {saving ? (
          <ActivityIndicator color={colors.primaryText} />
        ) : (
          <Text style={styles.saveButtonText}>שמור מוצר</Text>
        )}
      </TouchableOpacity>
    </ScrollView>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <View style={styles.field}>
      <Text style={styles.fieldLabel}>{label}</Text>
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  content: { padding: 20, gap: 14, paddingBottom: 40 },
  heading: { fontSize: 20, fontWeight: '700', color: colors.text, textAlign: 'right' },
  receiptImage: { width: '100%', height: 160, borderRadius: 12, backgroundColor: colors.border },
  lowConfidenceNote: {
    backgroundColor: colors.statusWarningBg,
    color: colors.statusWarning,
    padding: 10,
    borderRadius: 8,
    textAlign: 'right',
    fontSize: 13,
  },
  field: { gap: 6 },
  fieldLabel: { fontSize: 13, color: colors.textMuted, textAlign: 'right' },
  input: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 16,
    textAlign: 'right',
  },
  warrantyBox: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: 14,
    borderWidth: 1,
    borderColor: colors.border,
    gap: 6,
  },
  warrantyText: { fontSize: 16, fontWeight: '600', color: colors.text, textAlign: 'right' },
  warrantyHint: { fontSize: 12, color: colors.textMuted, textAlign: 'right' },
  saveButton: {
    backgroundColor: colors.primary,
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    marginTop: 8,
  },
  saveButtonText: { color: colors.primaryText, fontSize: 16, fontWeight: '600' },
});
