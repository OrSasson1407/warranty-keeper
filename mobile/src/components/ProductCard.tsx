import { useEffect, useRef } from 'react';
import { Animated, Image, StyleSheet, Text, View } from 'react-native';

import type { Product } from '../api/types';
import { categoryIcon } from '../data/categoryIcons';
import { colors, statusAccent } from '../theme/colors';
import { typography } from '../theme/typography';
import { daysUntil, formatHebrewDate, warrantyStatus } from '../utils/warrantyStatus';
import PressableScale from './PressableScale';
import StatusBadge from './StatusBadge';

const URGENT_THRESHOLD_DAYS = 7;

export default function ProductCard({
  product,
  onPress,
}: {
  product: Product;
  onPress: () => void;
}) {
  const status = warrantyStatus(product.warranty_expires_at);
  const accent = statusAccent(status);
  const urgent = status === 'warning' && daysUntil(product.warranty_expires_at) <= URGENT_THRESHOLD_DAYS;

  const pulse = useRef(new Animated.Value(1)).current;
  useEffect(() => {
    if (!urgent) return;
    const loop = Animated.loop(
      Animated.sequence([
        Animated.timing(pulse, { toValue: 0.3, duration: 600, useNativeDriver: true }),
        Animated.timing(pulse, { toValue: 1, duration: 600, useNativeDriver: true }),
      ]),
    );
    loop.start();
    return () => loop.stop();
  }, [urgent, pulse]);

  return (
    <PressableScale style={styles.card} onPress={onPress}>
      <Animated.View
        style={[styles.beacon, { backgroundColor: accent }, urgent && { opacity: pulse }]}
      />

      <View style={styles.header}>
        {product.photo_url ? (
          <Image source={{ uri: product.photo_url }} style={styles.thumb} />
        ) : (
          <View style={[styles.thumb, styles.thumbPlaceholder]}>
            <Text style={styles.thumbEmoji}>{categoryIcon(product.category)}</Text>
          </View>
        )}

        <View style={styles.info}>
          <Text style={styles.eyebrow} numberOfLines={1}>
            {product.category}
            {product.room ? `  •  ${product.room}` : ''}
          </Text>
          <Text style={styles.title} numberOfLines={1}>
            {product.name}
          </Text>
        </View>

        <View style={styles.trailing}>
          {product.price ? <Text style={styles.price}>{`₪${product.price.toLocaleString()}`}</Text> : null}
          <Text style={styles.date}>{formatHebrewDate(product.purchase_date)}</Text>
        </View>
      </View>

      <View style={styles.footer}>
        <StatusBadge expiresAt={product.warranty_expires_at} />
        <Text style={styles.link}>פרטים ←</Text>
      </View>
    </PressableScale>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.surfaceContainer,
    borderWidth: 1,
    borderColor: colors.border,
    padding: 16,
    gap: 12,
  },
  beacon: { position: 'absolute', top: 0, right: 0, left: 0, height: 2 },
  header: { flexDirection: 'row-reverse', alignItems: 'flex-start', gap: 12 },
  thumb: { width: 40, height: 40, backgroundColor: colors.surfaceContainerHigh },
  thumbPlaceholder: { alignItems: 'center', justifyContent: 'center' },
  thumbEmoji: { fontSize: 18 },
  info: { flex: 1, alignItems: 'flex-end', gap: 2 },
  eyebrow: { ...typography.labelSm, color: colors.outline },
  title: { ...typography.headlineSm, color: colors.text, textAlign: 'right' },
  trailing: { alignItems: 'flex-start', gap: 2 },
  price: { ...typography.labelNumeric, color: colors.text },
  date: { ...typography.labelSm, color: colors.outline, textTransform: 'none' },
  footer: { flexDirection: 'row-reverse', alignItems: 'center', justifyContent: 'space-between' },
  link: { ...typography.labelSm, color: colors.primary, textTransform: 'none' },
});
