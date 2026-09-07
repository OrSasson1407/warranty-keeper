import { useCallback, useState } from 'react';
import { FlatList, RefreshControl, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { useFocusEffect } from '@react-navigation/native';

import { api } from '../api/client';
import { loadProductsCache, saveProductsCache } from '../api/offlineCache';
import type { Product } from '../api/types';
import DashboardSummary from '../components/DashboardSummary';
import ProductCard from '../components/ProductCard';
import { ProductCardSkeleton } from '../components/Skeleton';
import { useAuth } from '../context/AuthContext';
import { colors } from '../theme/colors';
import { typography } from '../theme/typography';
import { computeAnalytics } from '../utils/analytics';
import { formatHebrewDate, warrantyStatus } from '../utils/warrantyStatus';
import type { AppTabScreenProps } from '../navigation/types';

type Props = AppTabScreenProps<'DashboardTab'>;

type Filter = 'all' | 'warning' | 'expired';

export default function DashboardScreen({ navigation }: Props) {
  const { user } = useAuth();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<Filter>('all');
  const [offlineSince, setOfflineSince] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const data = await api.listProducts();
      setProducts(data);
      setOfflineSince(null);
      saveProductsCache(data);
    } catch {
      const cache = await loadProductsCache();
      if (cache) {
        setProducts(cache.products);
        setOfflineSince(cache.cachedAt);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      load();
    }, [load]),
  );

  const visible =
    filter === 'all' ? products : products.filter((p) => warrantyStatus(p.warranty_expires_at) === filter);

  const expiredCount = products.filter((p) => warrantyStatus(p.warranty_expires_at) === 'expired').length;
  const analytics = computeAnalytics(products);

  const firstName = user?.full_name?.split(' ')[0] ?? '';

  return (
    <View style={styles.container}>
      {offlineSince ? (
        <View style={styles.offlineBanner}>
          <Text style={styles.offlineBannerText}>
            אין חיבור לאינטרנט — מוצג מידע שמור מ-{formatHebrewDate(offlineSince)}
          </Text>
        </View>
      ) : null}

      <View style={styles.header}>
        <Text style={styles.greeting}>שלום, {firstName} 👋</Text>
        <Text style={styles.subGreeting}>
          {analytics.totalCount} מוצרים פעילים
          {analytics.expiringSoonCount > 0 ? `, ${analytics.expiringSoonCount} עומד לפוג בקרוב` : ''}
        </Text>
      </View>

      <View style={styles.tabs}>
        <TouchableOpacity
          onPress={() => setFilter('all')}
          style={[styles.tab, filter === 'all' && styles.tabActive]}
        >
          <Text style={[styles.tabText, filter === 'all' && styles.tabTextActive]}>
            הכל ({products.length})
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          onPress={() => setFilter('warning')}
          style={[styles.tab, filter === 'warning' && styles.tabActive]}
        >
          <Text style={[styles.tabText, filter === 'warning' && styles.tabTextActive]}>
            קרוב לתפוגה ({analytics.expiringSoonCount})
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          onPress={() => setFilter('expired')}
          style={[styles.tab, filter === 'expired' && styles.tabActive]}
        >
          <Text style={[styles.tabText, filter === 'expired' && styles.tabTextActive]}>
            פג תוקף ({expiredCount})
          </Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={visible}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.list}
        refreshControl={<RefreshControl refreshing={loading} onRefresh={load} />}
        ListHeaderComponent={
          products.length > 0 ? <DashboardSummary analytics={analytics} /> : null
        }
        ListEmptyComponent={
          loading ? (
            <View style={{ gap: 12 }}>
              <ProductCardSkeleton />
              <ProductCardSkeleton />
              <ProductCardSkeleton />
            </View>
          ) : (
            <View style={styles.empty}>
              <Text style={styles.emptyIcon}>🛡️</Text>
              <Text style={styles.emptyText}>עדיין אין מוצרים. הוסיפו את הראשון עם הכפתור למטה!</Text>
            </View>
          )
        }
        renderItem={({ item }) => (
          <ProductCard
            product={item}
            onPress={() => navigation.navigate('ProductDetail', { productId: item.id })}
          />
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  offlineBanner: {
    marginHorizontal: 20,
    marginTop: 12,
    backgroundColor: colors.secondaryContainer,
    padding: 10,
  },
  offlineBannerText: { ...typography.bodySm, color: colors.onSecondaryContainer, textAlign: 'center' },
  header: { paddingHorizontal: 20, paddingTop: 16, paddingBottom: 8, alignItems: 'flex-end' },
  greeting: { ...typography.headlineLg, color: colors.text },
  subGreeting: { ...typography.bodySm, color: colors.textMuted, marginTop: 4 },
  tabs: { flexDirection: 'row', gap: 4, paddingHorizontal: 20, marginBottom: 12 },
  tab: {
    flex: 1,
    paddingVertical: 8,
    alignItems: 'center',
    backgroundColor: colors.surfaceContainerLow,
  },
  tabActive: { backgroundColor: colors.primary },
  tabText: { ...typography.labelSm, color: colors.textMuted, textTransform: 'none' },
  tabTextActive: { color: colors.primaryText, fontFamily: typography.labelSm.fontFamily },
  list: { paddingHorizontal: 20, paddingBottom: 24, gap: 12 },
  empty: { paddingTop: 60, paddingHorizontal: 20, alignItems: 'center', gap: 12 },
  emptyIcon: { fontSize: 40 },
  emptyText: { ...typography.bodyMd, color: colors.textMuted, textAlign: 'center' },
});
