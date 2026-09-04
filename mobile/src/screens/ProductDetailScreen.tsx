import { useCallback, useState } from 'react';
import { ActivityIndicator, Image, Linking, ScrollView, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { useFocusEffect } from '@react-navigation/native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api } from '../api/client';
import type { Product, WarrantyClaim } from '../api/types';
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
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [p, c] = await Promise.all([api.getProduct(productId), api.listClaims(productId)]);
      setProduct(p);
      setClaims(c);
    } finally {
      setLoading(false);
    }
  }, [productId]);

  useFocusEffect(
    useCallback(() => {
      load();
    }, [load])
  );

  if (loading || !product) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  const days = daysUntil(product.warranty_expires_at);

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

      {product.receipt_id && product.photo_url ? (
        <TouchableOpacity style={styles.receiptButton} onPress={() => Linking.openURL(product.photo_url)}>
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
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  content: { padding: 20, gap: 12, paddingBottom: 48 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.background },
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
});
