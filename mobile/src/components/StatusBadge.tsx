import { StyleSheet, Text, View } from 'react-native';

import { statusColor } from '../theme/colors';
import { typography } from '../theme/typography';
import { statusLabel, warrantyStatus } from '../utils/warrantyStatus';

export default function StatusBadge({ expiresAt }: { expiresAt: string }) {
  const status = warrantyStatus(expiresAt);
  const { fg, bg } = statusColor(status);

  return (
    <View style={[styles.badge, { backgroundColor: bg }]}>
      <View style={[styles.dot, { backgroundColor: fg }]} />
      <Text style={[styles.text, { color: fg }]}>{statusLabel(expiresAt)}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    flexDirection: 'row',
    alignItems: 'center',
    alignSelf: 'flex-start',
    paddingHorizontal: 10,
    paddingVertical: 6,
    gap: 6,
  },
  dot: { width: 6, height: 6 },
  text: { ...typography.labelSm },
});
