import { useCallback, useState } from 'react';
import { FlatList, RefreshControl, StyleSheet, Text, View } from 'react-native';
import { useFocusEffect } from '@react-navigation/native';

import { api } from '../api/client';
import type { Product } from '../api/types';
import ProductCard from '../components/ProductCard';
import { ProductCardSkeleton } from '../components/Skeleton';
import { colors } from '../theme/colors';
import { typography } from '../theme/typography';
import { warrantyStatus } from '../utils/warrantyStatus';
import type { AppTabScreenProps } from '../navigation/types';

type Props = AppTabScreenProps<'ExpiringTab'>;

export default function ExpiringScreen({ navigation }: Props) {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    try {
      const data = await api.listProducts();
      setProducts(data.filter((p) => warrantyStatus(p.warranty_expires_at) !== 'ok'));
    } finally {
      setLoading(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      load();
    }, [load]),
  );

  return (
    <View style={styles.container}>
      <Text style={styles.heading}>קרוב לתפוגה / פג תוקף</Text>
      <FlatList
        data={products}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.list}
        refreshControl={<RefreshControl refreshing={loading} onRefresh={load} />}
        ListEmptyComponent={
          loading ? (
            <View style={{ gap: 12 }}>
              <ProductCardSkeleton />
              <ProductCardSkeleton />
            </View>
          ) : (
            <View style={styles.empty}>
              <Text style={styles.emptyIcon}>⏰</Text>
              <Text style={styles.emptyText}>אין מוצרים שעומדים לפוג או שפגה אחריותם כרגע.</Text>
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
  heading: {
    ...typography.headlineLg,
    color: colors.text,
    textAlign: 'right',
    paddingHorizontal: 20,
    paddingTop: 16,
    paddingBottom: 8,
  },
  list: { paddingHorizontal: 20, paddingBottom: 24, gap: 12 },
  empty: { paddingTop: 60, paddingHorizontal: 20, alignItems: 'center', gap: 12 },
  emptyIcon: { fontSize: 40 },
  emptyText: { ...typography.bodyMd, color: colors.textMuted, textAlign: 'center' },
});
