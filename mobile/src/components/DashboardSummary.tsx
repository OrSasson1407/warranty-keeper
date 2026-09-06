import { StyleSheet, Text, View } from 'react-native';

import type { DashboardAnalytics } from '../utils/analytics';
import { colors } from '../theme/colors';

function formatILS(amount: number): string {
  return `₪${Math.round(amount).toLocaleString('he-IL')}`;
}

export default function DashboardSummary({ analytics }: { analytics: DashboardAnalytics }) {
  const { coveredValue, expiringSoonCount, byCategory } = analytics;
  const maxCount = byCategory.length > 0 ? byCategory[0].count : 1;

  return (
    <View style={styles.container}>
      <View style={styles.statsRow}>
        <View style={styles.stat}>
          <Text style={styles.statValue}>{formatILS(coveredValue)}</Text>
          <Text style={styles.statLabel}>שווי מוצרים באחריות</Text>
        </View>
        <View style={styles.stat}>
          <Text style={[styles.statValue, expiringSoonCount > 0 && styles.statValueWarning]}>
            {expiringSoonCount}
          </Text>
          <Text style={styles.statLabel}>פגות תוקף ב-30 הימים הקרובים</Text>
        </View>
      </View>

      {byCategory.length > 0 ? (
        <View style={styles.categories}>
          {byCategory.map(({ category, count }) => (
            <View key={category} style={styles.categoryRow}>
              <Text style={styles.categoryLabel} numberOfLines={1}>
                {category}
              </Text>
              <View style={styles.barTrack}>
                <View style={[styles.barFill, { width: `${(count / maxCount) * 100}%` }]} />
              </View>
              <Text style={styles.categoryCount}>{count}</Text>
            </View>
          ))}
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    marginBottom: 12,
    backgroundColor: colors.surface,
    borderRadius: 14,
    padding: 14,
    gap: 12,
  },
  statsRow: { flexDirection: 'row', gap: 12 },
  stat: { flex: 1, alignItems: 'center', gap: 4 },
  statValue: { fontSize: 18, fontWeight: '700', color: colors.text },
  statValueWarning: { color: colors.statusWarning },
  statLabel: { fontSize: 11, color: colors.textMuted, textAlign: 'center' },
  categories: { gap: 6, borderTopWidth: 1, borderTopColor: colors.border, paddingTop: 10 },
  categoryRow: { flexDirection: 'row-reverse', alignItems: 'center', gap: 8 },
  categoryLabel: { width: 72, fontSize: 12, color: colors.text, textAlign: 'right' },
  barTrack: { flex: 1, height: 8, borderRadius: 4, backgroundColor: colors.border },
  barFill: { height: 8, borderRadius: 4, backgroundColor: colors.primary },
  categoryCount: { width: 20, fontSize: 12, color: colors.textMuted, textAlign: 'left' },
});
