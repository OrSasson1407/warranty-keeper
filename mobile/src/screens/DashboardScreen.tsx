import { useCallback, useState } from 'react';
import { FlatList, Image, RefreshControl, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { useFocusEffect } from '@react-navigation/native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api } from '../api/client';
import type { Product } from '../api/types';
import StatusBadge from '../components/StatusBadge';
import { useAuth } from '../context/AuthContext';
import { colors } from '../theme/colors';
import { warrantyStatus } from '../utils/warrantyStatus';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'Dashboard'>;

type Filter = 'all' | 'expiring';

export default function DashboardScreen({ navigation }: Props) {
  const { user } = useAuth();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<Filter>('all');

  const load = useCallback(async () => {
    try {
      const data = await api.listProducts();
      setProducts(data);
    } finally {
      setLoading(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      load();
    }, [load])
  );

  const visible =
    filter === 'expiring'
      ? products.filter((p) => warrantyStatus(p.warranty_expires_at) !== 'ok')
      : products;

  const firstName = user?.full_name?.split(' ')[0] ?? '';

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.greeting}>שלום, {firstName} 👋</Text>
        <View style={styles.headerIcons}>
          <TouchableOpacity onPress={() => navigation.navigate('Search')} style={styles.iconButton}>
            <Text style={styles.icon}>🔍</Text>
          </TouchableOpacity>
          <TouchableOpacity onPress={() => navigation.navigate('Settings')} style={styles.iconButton}>
            <Text style={styles.icon}>⚙️</Text>
          </TouchableOpacity>
        </View>
      </View>

      <View style={styles.tabs}>
        <TouchableOpacity onPress={() => setFilter('all')} style={[styles.tab, filter === 'all' && styles.tabActive]}>
          <Text style={[styles.tabText, filter === 'all' && styles.tabTextActive]}>הכל</Text>
        </TouchableOpacity>
        <TouchableOpacity
          onPress={() => setFilter('expiring')}
          style={[styles.tab, filter === 'expiring' && styles.tabActive]}
        >
          <Text style={[styles.tabText, filter === 'expiring' && styles.tabTextActive]}>קרוב לתפוגה</Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={visible}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.list}
        refreshControl={<RefreshControl refreshing={loading} onRefresh={load} />}
        ListEmptyComponent={
          !loading ? (
            <View style={styles.empty}>
              <Text style={styles.emptyText}>עדיין אין מוצרים. הוסיפו את הראשון עם הכפתור למטה!</Text>
            </View>
          ) : null
        }
        renderItem={({ item }) => (
          <TouchableOpacity
            style={styles.card}
            onPress={() => navigation.navigate('ProductDetail', { productId: item.id })}
          >
            {item.photo_url ? (
              <Image source={{ uri: item.photo_url }} style={styles.thumb} />
            ) : (
              <View style={[styles.thumb, styles.thumbPlaceholder]}>
                <Text style={styles.thumbEmoji}>📦</Text>
              </View>
            )}
            <View style={styles.cardBody}>
              <Text style={styles.cardTitle}>{item.name}</Text>
              <StatusBadge expiresAt={item.warranty_expires_at} />
            </View>
          </TouchableOpacity>
        )}
      />

      <TouchableOpacity style={styles.fab} onPress={() => navigation.navigate('AddProductChoose')}>
        <Text style={styles.fabText}>+</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingTop: 16,
    paddingBottom: 8,
  },
  greeting: { fontSize: 20, fontWeight: '700', color: colors.text },
  headerIcons: { flexDirection: 'row', gap: 12 },
  iconButton: { padding: 6 },
  icon: { fontSize: 20 },
  tabs: { flexDirection: 'row', gap: 8, paddingHorizontal: 20, marginBottom: 8 },
  tab: { paddingHorizontal: 14, paddingVertical: 8, borderRadius: 999, backgroundColor: colors.surface },
  tabActive: { backgroundColor: colors.primary },
  tabText: { color: colors.textMuted, fontWeight: '600', fontSize: 13 },
  tabTextActive: { color: colors.primaryText },
  list: { paddingHorizontal: 20, paddingBottom: 100, gap: 12 },
  empty: { paddingTop: 60, paddingHorizontal: 20 },
  emptyText: { textAlign: 'center', color: colors.textMuted, fontSize: 15 },
  card: {
    flexDirection: 'row',
    backgroundColor: colors.surface,
    borderRadius: 14,
    padding: 12,
    gap: 12,
    alignItems: 'center',
  },
  thumb: { width: 56, height: 56, borderRadius: 10, backgroundColor: colors.border },
  thumbPlaceholder: { alignItems: 'center', justifyContent: 'center' },
  thumbEmoji: { fontSize: 24 },
  cardBody: { flex: 1, gap: 6 },
  cardTitle: { fontSize: 16, fontWeight: '600', color: colors.text, textAlign: 'right' },
  fab: {
    position: 'absolute',
    bottom: 28,
    alignSelf: 'center',
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    elevation: 4,
    shadowColor: '#000',
    shadowOpacity: 0.2,
    shadowRadius: 8,
    shadowOffset: { width: 0, height: 4 },
  },
  fabText: { color: colors.primaryText, fontSize: 32, lineHeight: 34, fontWeight: '400' },
});
