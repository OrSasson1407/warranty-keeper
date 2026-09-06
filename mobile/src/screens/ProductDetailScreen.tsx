import { useCallback, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Image,
  Linking,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import { useFocusEffect } from '@react-navigation/native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api, ApiError } from '../api/client';
import type { Product, ProductCost, WarrantyClaim } from '../api/types';
import { addWarrantyExpiryToCalendar } from '../calendar/syncWarrantyEvent';
import StatusBadge from '../components/StatusBadge';
import { colors } from '../theme/colors';
import { daysUntil, formatHebrewDate } from '../utils/warrantyStatus';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'ProductDetail'>;

const CLAIM_STATUS_LABELS: Record<WarrantyClaim['status'], string> = {
  open: 'פתוח',
  in_progress: 'בטיפול',
  closed: 'נסגר',
};

export default function ProductDetailScreen({ navigation, route }: Props) {
  const { productId } = route.params;
  const [product, setProduct] = useState<Product | null>(null);
  const [claims, setClaims] = useState<WarrantyClaim[]>([]);
  const [costs, setCosts] = useState<ProductCost[]>([]);
  const [loading, setLoading] = useState(true);
  const [addingToCalendar, setAddingToCalendar] = useState(false);
  const [addingCost, setAddingCost] = useState(false);
  const [savingCost, setSavingCost] = useState(false);
  const [costAmount, setCostAmount] = useState('');
  const [costDescription, setCostDescription] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [p, c, tco] = await Promise.all([
        api.getProduct(productId),
        api.listClaims(productId),
        api.listProductCosts(productId),
      ]);
      setProduct(p);
      setClaims(c);
      setCosts(tco);
    } finally {
      setLoading(false);
    }
  }, [productId]);

  useFocusEffect(
    useCallback(() => {
      load();
    }, [load]),
  );

  if (loading || !product) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  const days = daysUntil(product.warranty_expires_at);

  const totalCost = (product.price ?? 0) + costs.reduce((sum, c) => sum + c.amount, 0);

  const onSaveCost = async () => {
    const amount = Number(costAmount);
    if (!costAmount || isNaN(amount) || amount <= 0) {
      Alert.alert('סכום לא תקין', 'נא להזין סכום גדול מ-0');
      return;
    }
    setSavingCost(true);
    try {
      const cost = await api.createProductCost(product.id, {
        amount,
        description: costDescription.trim() || undefined,
      });
      setCosts((prev) => [cost, ...prev]);
      setCostAmount('');
      setCostDescription('');
      setAddingCost(false);
    } catch (e) {
      Alert.alert('שגיאה', e instanceof ApiError ? e.message : 'לא הצלחנו לשמור את העלות');
    } finally {
      setSavingCost(false);
    }
  };

  const onAddToCalendar = async () => {
    setAddingToCalendar(true);
    try {
      const ok = await addWarrantyExpiryToCalendar(product.name, product.warranty_expires_at);
      Alert.alert(
        ok ? 'נוסף ליומן' : 'לא ניתן להוסיף ליומן',
        ok ? 'תזכורת לתפוגת האחריות נוספה ליומן המכשיר.' : 'נדרשת הרשאת יומן במכשיר.',
      );
    } finally {
      setAddingToCalendar(false);
    }
  };

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {product.photo_url ? (
        <Image source={{ uri: product.photo_url }} style={styles.image} resizeMode="cover" />
      ) : (
        <View style={[styles.image, styles.imagePlaceholder]}>
          <Text style={{ fontSize: 40 }}>📦</Text>
        </View>
      )}

      <Text style={styles.name}>{product.name}</Text>
      <Text style={styles.meta}>
        נקנה: {formatHebrewDate(product.purchase_date)}
        {product.price ? ` | ₪${product.price.toLocaleString()}` : ''}
      </Text>
      {product.room ? <Text style={styles.meta}>חדר: {product.room}</Text> : null}

      <View style={styles.statusSection}>
        <StatusBadge expiresAt={product.warranty_expires_at} />
        <Text style={styles.daysText}>
          {days >= 0 ? `נותרו ${days} ימים` : `פג לפני ${Math.abs(days)} ימים`}
        </Text>
        {product.warranty_uncertain ? (
          <Text style={styles.uncertainNote}>תאריך משוער — ייתכן שקיימת אחריות שונה בפועל</Text>
        ) : null}
      </View>

      <TouchableOpacity
        style={styles.calendarButton}
        onPress={onAddToCalendar}
        disabled={addingToCalendar}
      >
        {addingToCalendar ? (
          <ActivityIndicator color={colors.primary} />
        ) : (
          <Text style={styles.calendarButtonText}>📅 הוסף תזכורת ליומן</Text>
        )}
      </TouchableOpacity>

      {product.receipt_id && product.photo_url ? (
        <TouchableOpacity
          style={styles.receiptButton}
          onPress={() => Linking.openURL(product.photo_url)}
        >
          <Text style={styles.receiptButtonText}>🧾 צפה בקבלה</Text>
        </TouchableOpacity>
      ) : null}

      <TouchableOpacity
        style={styles.claimCta}
        onPress={() => navigation.navigate('Claim', { productId: product.id })}
      >
        <Text style={styles.claimCtaText}>המוצר התקלקל?</Text>
      </TouchableOpacity>

      <View style={styles.claimsLog}>
        <Text style={styles.claimsHeading}>יומן תקלות</Text>
        {claims.length === 0 ? (
          <Text style={styles.meta}>אין רשומות</Text>
        ) : (
          claims.map((claim) => (
            <View key={claim.id} style={styles.claimItem}>
              <Text style={styles.claimDate}>{formatHebrewDate(claim.created_at)}</Text>
              <Text style={styles.claimDescription}>{claim.issue_description}</Text>
              <Text style={styles.claimStatus}>{CLAIM_STATUS_LABELS[claim.status]}</Text>
            </View>
          ))
        )}
      </View>

      <View style={styles.claimsLog}>
        <View style={styles.tcoHeader}>
          <TouchableOpacity onPress={() => setAddingCost((v) => !v)}>
            <Text style={styles.addCostLink}>{addingCost ? 'ביטול' : '+ הוסף עלות'}</Text>
          </TouchableOpacity>
          <Text style={styles.claimsHeading}>עלות בעלות כוללת</Text>
        </View>
        <Text style={styles.tcoTotal}>{`₪${totalCost.toLocaleString()}`}</Text>

        {addingCost ? (
          <View style={styles.costForm}>
            <TextInput
              style={styles.costInput}
              placeholder="סכום"
              keyboardType="numeric"
              value={costAmount}
              onChangeText={setCostAmount}
            />
            <TextInput
              style={styles.costInput}
              placeholder="תיאור (אופציונלי)"
              value={costDescription}
              onChangeText={setCostDescription}
            />
            <TouchableOpacity
              style={styles.costSaveButton}
              onPress={onSaveCost}
              disabled={savingCost}
            >
              {savingCost ? (
                <ActivityIndicator color={colors.primaryText} />
              ) : (
                <Text style={styles.costSaveButtonText}>שמור עלות</Text>
              )}
            </TouchableOpacity>
          </View>
        ) : null}

        {costs.length === 0 ? (
          <Text style={styles.meta}>
            {product.price ? 'לא נוספו עלויות נוספות' : 'אין נתוני עלות'}
          </Text>
        ) : (
          costs.map((cost) => (
            <View key={cost.id} style={styles.claimItem}>
              <Text style={styles.claimDate}>{formatHebrewDate(cost.incurred_at)}</Text>
              <Text style={styles.claimDescription}>
                {cost.description || 'עלות נוספת'} — ₪{cost.amount.toLocaleString()}
              </Text>
            </View>
          ))
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  content: { padding: 20, gap: 12, paddingBottom: 48 },
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
  image: { width: '100%', height: 200, borderRadius: 14, backgroundColor: colors.border },
  imagePlaceholder: { alignItems: 'center', justifyContent: 'center' },
  name: { fontSize: 22, fontWeight: '700', color: colors.text, textAlign: 'right' },
  meta: { fontSize: 14, color: colors.textMuted, textAlign: 'right' },
  statusSection: { gap: 6, alignItems: 'flex-end', marginTop: 4 },
  daysText: { fontSize: 13, color: colors.textMuted },
  uncertainNote: { fontSize: 12, color: colors.statusWarning },
  receiptButton: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingVertical: 12,
    alignItems: 'center',
  },
  receiptButtonText: { color: colors.text, fontSize: 15 },
  calendarButton: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingVertical: 12,
    alignItems: 'center',
  },
  calendarButtonText: { color: colors.text, fontSize: 15 },
  claimCta: {
    backgroundColor: colors.danger,
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 8,
  },
  claimCtaText: { color: colors.primaryText, fontSize: 16, fontWeight: '600' },
  claimsLog: { marginTop: 16, gap: 10 },
  claimsHeading: { fontSize: 16, fontWeight: '700', color: colors.text, textAlign: 'right' },
  claimItem: {
    backgroundColor: colors.surface,
    borderRadius: 10,
    padding: 12,
    gap: 4,
  },
  claimDate: { fontSize: 12, color: colors.textMuted, textAlign: 'right' },
  claimDescription: { fontSize: 14, color: colors.text, textAlign: 'right' },
  claimStatus: { fontSize: 12, color: colors.primary, textAlign: 'right', fontWeight: '600' },
  tcoHeader: {
    flexDirection: 'row-reverse',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  addCostLink: { color: colors.primary, fontSize: 14, fontWeight: '600' },
  tcoTotal: { fontSize: 22, fontWeight: '700', color: colors.text, textAlign: 'right' },
  costForm: { gap: 8, marginBottom: 4 },
  costInput: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 15,
    textAlign: 'right',
  },
  costSaveButton: {
    backgroundColor: colors.primary,
    borderRadius: 10,
    paddingVertical: 10,
    alignItems: 'center',
  },
  costSaveButtonText: { color: colors.primaryText, fontWeight: '600', fontSize: 14 },
});
