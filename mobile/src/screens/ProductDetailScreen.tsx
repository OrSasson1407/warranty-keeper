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
import { colors, statusAccent } from '../theme/colors';
import { fonts, typography } from '../theme/typography';
import { daysUntil, formatHebrewDate, warrantyStatus } from '../utils/warrantyStatus';
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
  const [reportingRule, setReportingRule] = useState(false);

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
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  const days = daysUntil(product.warranty_expires_at);
  const accent = statusAccent(warrantyStatus(product.warranty_expires_at));
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

  const onReportWarrantyRule = async () => {
    setReportingRule(true);
    try {
      await api.reportWarrantyRule(product.id);
      Alert.alert('תודה על הדיווח', 'נבדוק את תקופת האחריות לקטגוריה הזו.');
    } catch {
      Alert.alert('שגיאה', 'לא הצלחנו לשלוח את הדיווח, נסו שוב.');
    } finally {
      setReportingRule(false);
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
      <View style={styles.imageWrap}>
        {product.photo_url ? (
          <Image source={{ uri: product.photo_url }} style={styles.image} resizeMode="cover" />
        ) : (
          <View style={[styles.image, styles.imagePlaceholder]}>
            <Text style={{ fontSize: 40 }}>📦</Text>
          </View>
        )}
        <View style={[styles.imageAccent, { backgroundColor: accent }]} />
      </View>

      <View style={styles.titleRow}>
        <Text style={styles.name}>{product.name}</Text>
        <StatusBadge expiresAt={product.warranty_expires_at} />
      </View>
      <Text style={styles.daysText}>
        {days >= 0 ? `נותרו ${days} ימים` : `פג לפני ${Math.abs(days)} ימים`}
      </Text>
      {product.warranty_uncertain ? (
        <Text style={styles.uncertainNote}>תאריך משוער — ייתכן שקיימת אחריות שונה בפועל</Text>
      ) : null}
      <TouchableOpacity onPress={onReportWarrantyRule} disabled={reportingRule}>
        <Text style={styles.reportRuleLink}>תקופת האחריות נראית לא נכונה?</Text>
      </TouchableOpacity>

      <View style={styles.statRow}>
        <View style={styles.statCell}>
          <Text style={styles.statLabel}>חדר</Text>
          <Text style={styles.statValue}>{product.room || '—'}</Text>
        </View>
        <View style={styles.statCell}>
          <Text style={styles.statLabel}>תאריך רכישה</Text>
          <Text style={styles.statValue}>{formatHebrewDate(product.purchase_date)}</Text>
        </View>
        <View style={styles.statCell}>
          <Text style={styles.statLabel}>מחיר רכישה</Text>
          <Text style={styles.statValue}>{product.price ? `₪${product.price.toLocaleString()}` : '—'}</Text>
        </View>
      </View>

      <View style={styles.actionsRow}>
        <TouchableOpacity
          style={styles.actionButton}
          onPress={onAddToCalendar}
          disabled={addingToCalendar}
        >
          {addingToCalendar ? (
            <ActivityIndicator color={colors.primary} />
          ) : (
            <Text style={styles.actionButtonText}>תזכורת ליומן</Text>
          )}
        </TouchableOpacity>
        {product.receipt_id && product.photo_url ? (
          <TouchableOpacity
            style={styles.actionButton}
            onPress={() => Linking.openURL(product.photo_url)}
          >
            <Text style={styles.actionButtonText}>קבלה מקורית</Text>
          </TouchableOpacity>
        ) : null}
      </View>

      <TouchableOpacity
        style={styles.claimCta}
        onPress={() => navigation.navigate('Claim', { productId: product.id })}
      >
        <Text style={styles.claimCtaText}>המוצר התקלקל?</Text>
      </TouchableOpacity>

      <View style={styles.section}>
        <Text style={styles.sectionHeading}>יומן תקלות</Text>
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

      <View style={styles.section}>
        <View style={styles.tcoHeader}>
          <TouchableOpacity onPress={() => setAddingCost((v) => !v)}>
            <Text style={styles.addCostLink}>{addingCost ? 'ביטול' : '+ הוסף עלות'}</Text>
          </TouchableOpacity>
          <Text style={styles.sectionHeading}>עלות בעלות כוללת</Text>
        </View>
        <Text style={styles.tcoTotal}>{`₪${totalCost.toLocaleString()}`}</Text>

        {addingCost ? (
          <View style={styles.costForm}>
            <TextInput
              style={styles.costInput}
              placeholder="סכום"
              placeholderTextColor={colors.outline}
              keyboardType="numeric"
              value={costAmount}
              onChangeText={setCostAmount}
            />
            <TextInput
              style={styles.costInput}
              placeholder="תיאור (אופציונלי)"
              placeholderTextColor={colors.outline}
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
  content: { padding: 20, gap: 4, paddingBottom: 48 },
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
  imageWrap: { position: 'relative', marginBottom: 12 },
  image: { width: '100%', height: 200, backgroundColor: colors.surfaceContainerHigh },
  imagePlaceholder: { alignItems: 'center', justifyContent: 'center' },
  imageAccent: { position: 'absolute', bottom: 0, right: 0, left: 0, height: 2 },
  titleRow: {
    flexDirection: 'row-reverse',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    gap: 8,
  },
  name: { ...typography.headlineLg, color: colors.text, textAlign: 'right', flex: 1 },
  meta: { ...typography.bodyMd, color: colors.textMuted, textAlign: 'right' },
  daysText: { ...typography.bodySm, color: colors.textMuted, textAlign: 'right', marginTop: 4 },
  uncertainNote: { ...typography.bodySm, color: colors.secondary, textAlign: 'right', marginTop: 4 },
  reportRuleLink: {
    ...typography.bodySm,
    color: colors.textMuted,
    textAlign: 'right',
    textDecorationLine: 'underline',
    marginTop: 4,
  },
  statRow: {
    flexDirection: 'row',
    marginTop: 16,
    backgroundColor: colors.surfaceContainer,
    borderWidth: 1,
    borderColor: colors.border,
  },
  statCell: { flex: 1, alignItems: 'center', paddingVertical: 12, gap: 4 },
  statLabel: { ...typography.labelSm, color: colors.outline },
  statValue: { ...typography.bodyMd, color: colors.text, textAlign: 'center' },
  actionsRow: { flexDirection: 'row', gap: 8, marginTop: 8 },
  actionButton: {
    flex: 1,
    backgroundColor: colors.surfaceContainer,
    borderWidth: 1,
    borderColor: colors.border,
    paddingVertical: 12,
    alignItems: 'center',
  },
  actionButtonText: { ...typography.bodyMd, color: colors.text, fontFamily: fonts.bodyMdSemiBold },
  claimCta: {
    backgroundColor: colors.errorContainer,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 12,
  },
  claimCtaText: { ...typography.headlineSm, color: colors.onErrorContainer },
  section: { marginTop: 20, gap: 10 },
  sectionHeading: { ...typography.headlineSm, color: colors.text, textAlign: 'right' },
  claimItem: {
    backgroundColor: colors.surfaceContainer,
    borderWidth: 1,
    borderColor: colors.border,
    padding: 12,
    gap: 4,
  },
  claimDate: { ...typography.labelSm, color: colors.outline, textAlign: 'right', textTransform: 'none' },
  claimDescription: { ...typography.bodyMd, color: colors.text, textAlign: 'right' },
  claimStatus: { ...typography.labelSm, color: colors.primary, textAlign: 'right' },
  tcoHeader: {
    flexDirection: 'row-reverse',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  addCostLink: { ...typography.bodyMd, color: colors.primary },
  tcoTotal: { ...typography.displayLg, color: colors.text, textAlign: 'right' },
  costForm: { gap: 8, marginBottom: 4 },
  costInput: {
    backgroundColor: colors.surfaceContainerLowest,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: 12,
    paddingVertical: 10,
    color: colors.text,
    ...typography.bodyMd,
    textAlign: 'right',
  },
  costSaveButton: {
    backgroundColor: colors.primary,
    paddingVertical: 10,
    alignItems: 'center',
  },
  costSaveButtonText: { ...typography.bodyMd, color: colors.primaryText, fontFamily: fonts.bodyMdSemiBold },
});
