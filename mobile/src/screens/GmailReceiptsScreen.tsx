import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  RefreshControl,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api } from '../api/client';
import type { ReceiptDraft } from '../api/types';
import { colors } from '../theme/colors';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'GmailReceipts'>;

// Gmail-sourced receipts are created asynchronously by a background scan,
// not in response to a request this screen made -- so unlike a photo
// upload (which hands back a draft immediately), this screen has to poll
// the list endpoint to discover what showed up since last time.
export default function GmailReceiptsScreen({ navigation }: Props) {
  const [drafts, setDrafts] = useState<ReceiptDraft[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async (showSpinner: boolean) => {
    if (showSpinner) setLoading(true);
    try {
      const result = await api.listReceipts({ status: 'pending', source: 'gmail' });
      setDrafts(result);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- kicking off the initial async load, not derived state
    load(true);
  }, [load]);

  const onRefresh = () => {
    setRefreshing(true);
    load(false);
  };

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  return (
    <FlatList
      style={styles.container}
      contentContainerStyle={drafts.length === 0 ? styles.emptyContainer : styles.list}
      data={drafts}
      keyExtractor={(item) => item.receipt_id}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
      ListEmptyComponent={
        <Text style={styles.emptyText}>
          לא נמצאו קבלות חדשות מ-Gmail. סריקות מתבצעות מדי כמה שעות — נסו שוב מאוחר יותר.
        </Text>
      }
      renderItem={({ item }) => (
        <TouchableOpacity
          style={styles.card}
          onPress={() => navigation.navigate('ConfirmProduct', { draft: item })}
        >
          <Text style={styles.vendor}>{item.parsed_vendor || 'קבלה מ-Gmail'}</Text>
          {item.parsed_date ? <Text style={styles.meta}>{item.parsed_date}</Text> : null}
          {item.parsed_amount != null ? (
            <Text style={styles.meta}>₪{item.parsed_amount}</Text>
          ) : null}
          {item.confidence < 0.5 ? (
            <Text style={styles.lowConfidence}>נדרש אימות ידני של הפרטים</Text>
          ) : null}
        </TouchableOpacity>
      )}
    />
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
  list: { padding: 16, gap: 12 },
  emptyContainer: { flexGrow: 1, alignItems: 'center', justifyContent: 'center', padding: 24 },
  emptyText: { color: colors.textMuted, textAlign: 'center', fontSize: 15 },
  card: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: 14,
    gap: 4,
    borderWidth: 1,
    borderColor: colors.border,
  },
  vendor: { fontSize: 16, fontWeight: '700', color: colors.text, textAlign: 'right' },
  meta: { fontSize: 14, color: colors.textMuted, textAlign: 'right' },
  lowConfidence: { fontSize: 13, color: colors.statusWarning, textAlign: 'right', marginTop: 4 },
});
