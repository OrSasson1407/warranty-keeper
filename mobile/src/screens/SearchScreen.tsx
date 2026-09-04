import { useEffect, useState } from 'react';
import { ActivityIndicator, FlatList, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api } from '../api/client';
import type { Product } from '../api/types';
import StatusBadge from '../components/StatusBadge';
import { colors } from '../theme/colors';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'Search'>;

export default function SearchScreen({ navigation }: Props) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  useEffect(() => {
    const q = query.trim();
    if (!q) {
      setResults([]);
      setSearched(false);
      return;
    }
    setLoading(true);
    const timeout = setTimeout(() => {
      api
        .listProducts(q)
        .then((data) => {
          setResults(data);
          setSearched(true);
        })
        .finally(() => setLoading(false));
    }, 300);
    return () => clearTimeout(timeout);
  }, [query]);

  return (
    <View style={styles.container}>
      <TextInput
        style={styles.input}
        placeholder="🔍 חפש מוצר..."
        value={query}
        onChangeText={setQuery}
        autoFocus
      />

      {loading ? <ActivityIndicator style={{ marginTop: 20 }} /> : null}

      {searched && !loading ? (
        <Text style={styles.resultsLabel}>תוצאות ({results.length}):</Text>
      ) : null}

      <FlatList
        data={results}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.list}
        renderItem={({ item }) => (
          <TouchableOpacity
            style={styles.card}
            onPress={() => navigation.navigate('ProductDetail', { productId: item.id })}
          >
            <Text style={styles.cardTitle}>{item.name}</Text>
            <StatusBadge expiresAt={item.warranty_expires_at} />
          </TouchableOpacity>
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background, padding: 20, gap: 12 },
  input: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 16,
    textAlign: 'right',
  },
  resultsLabel: { fontSize: 13, color: colors.textMuted, textAlign: 'right' },
  list: { gap: 10 },
  card: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: 14,
    gap: 8,
  },
  cardTitle: { fontSize: 16, fontWeight: '600', color: colors.text, textAlign: 'right' },
});
