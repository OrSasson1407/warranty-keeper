import { useEffect, useRef } from 'react';
import { Animated, StyleSheet, View, type ViewStyle } from 'react-native';

import { colors } from '../theme/colors';

function Skeleton({ style }: { style?: ViewStyle | ViewStyle[] }) {
  const opacity = useRef(new Animated.Value(0.4)).current;

  useEffect(() => {
    const loop = Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, { toValue: 1, duration: 700, useNativeDriver: true }),
        Animated.timing(opacity, { toValue: 0.4, duration: 700, useNativeDriver: true }),
      ]),
    );
    loop.start();
    return () => loop.stop();
  }, [opacity]);

  return <Animated.View style={[styles.base, style, { opacity }]} />;
}

// A skeleton shaped like ProductCard, for the Dashboard/Search/Expiring
// product lists while the first fetch is in flight.
export function ProductCardSkeleton() {
  return (
    <View style={styles.card}>
      <View style={styles.header}>
        <Skeleton style={styles.thumb} />
        <View style={styles.info}>
          <Skeleton style={styles.lineShort} />
          <Skeleton style={styles.lineLong} />
        </View>
      </View>
      <Skeleton style={styles.badge} />
    </View>
  );
}

const styles = StyleSheet.create({
  base: { backgroundColor: colors.surfaceContainerHigh },
  card: {
    backgroundColor: colors.surfaceContainer,
    borderWidth: 1,
    borderColor: colors.border,
    padding: 16,
    gap: 12,
  },
  header: { flexDirection: 'row-reverse', alignItems: 'flex-start', gap: 12 },
  thumb: { width: 40, height: 40 },
  info: { flex: 1, alignItems: 'flex-end', gap: 8 },
  lineShort: { width: '40%', height: 10 },
  lineLong: { width: '75%', height: 16 },
  badge: { width: 120, height: 24, alignSelf: 'flex-start' },
});

export default Skeleton;
