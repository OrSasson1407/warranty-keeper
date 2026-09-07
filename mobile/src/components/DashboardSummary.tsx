import { StyleSheet, Text, View } from 'react-native';

import type { DashboardAnalytics } from '../utils/analytics';
import { colors } from '../theme/colors';
import { typography } from '../theme/typography';

function formatILS(amount: number): string {
  return `₪${Math.round(amount).toLocaleString('he-IL')}`;
}

export default function DashboardSummary({ analytics }: { analytics: DashboardAnalytics }) {
  const { coveredValue, expiringSoonCount, okCount, totalCount, byCategory } = analytics;
  const coveragePct = totalCount > 0 ? Math.round(((totalCount - expiringSoonCount) / totalCount) * 100) : 100;

  return (
    <View style={styles.container}>
      <View style={styles.accent} />

      <View style={styles.headerRow}>
        <Text style={styles.eyebrow}>סך שווי מוצרים מבוטחים</Text>
      </View>
      <View style={styles.valueRow}>
        <Text style={styles.value}>{formatILS(coveredValue)}</Text>
        <Text style={styles.coveragePct}>{coveragePct}% מכוסה</Text>
      </View>

      <View style={styles.split}>
        <View style={styles.splitCell}>
          <Text style={[styles.splitLabel, { color: colors.secondary }]}>קרוב לתפוגה</Text>
          <View style={styles.splitValueRow}>
            <Text style={[styles.splitValue, { color: colors.secondary }]}>{expiringSoonCount}</Text>
            <Text style={styles.splitHint}>{'< 30 ימים'}</Text>
          </View>
        </View>
        <View style={styles.splitCell}>
          <Text style={[styles.splitLabel, { color: colors.tertiary }]}>בתוקף מלא</Text>
          <View style={styles.splitValueRow}>
            <Text style={[styles.splitValue, { color: colors.tertiary }]}>{okCount}</Text>
            <Text style={styles.splitHint}>תקין ויציב</Text>
          </View>
        </View>
      </View>

      {byCategory.length > 0 ? (
        <View style={styles.chipsRow}>
          {byCategory.slice(0, 3).map(({ category, count }) => (
            <View key={category} style={styles.chip}>
              <Text style={styles.chipText} numberOfLines={1}>
                {category}: {count}
              </Text>
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
    backgroundColor: colors.surfaceContainer,
    padding: 16,
  },
  accent: { position: 'absolute', top: 0, right: 0, left: 0, height: 2, backgroundColor: colors.primary },
  headerRow: { flexDirection: 'row-reverse', justifyContent: 'space-between', paddingBottom: 10 },
  eyebrow: { ...typography.labelSm, color: colors.outline },
  valueRow: {
    flexDirection: 'row-reverse',
    alignItems: 'baseline',
    gap: 8,
    paddingBottom: 12,
  },
  value: { ...typography.displayLg, color: colors.text },
  coveragePct: { ...typography.labelSm, color: colors.tertiary, textTransform: 'none' },
  split: { flexDirection: 'row', gap: 8, paddingBottom: 12 },
  splitCell: { flex: 1, backgroundColor: colors.surfaceContainerHigh, padding: 8, gap: 8 },
  splitLabel: { ...typography.labelSm, textAlign: 'right' },
  splitValueRow: { flexDirection: 'row-reverse', justifyContent: 'space-between', alignItems: 'baseline' },
  splitValue: { ...typography.headlineSm },
  splitHint: { ...typography.labelSm, color: colors.outline, textTransform: 'none' },
  chipsRow: { flexDirection: 'row-reverse', justifyContent: 'space-between', paddingTop: 8, gap: 6 },
  chip: { flex: 1, backgroundColor: colors.surfaceContainerLowest, paddingVertical: 6, paddingHorizontal: 8 },
  chipText: { ...typography.labelSm, color: colors.textMuted, textTransform: 'none', textAlign: 'center' },
});
