import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api } from '../api/client';
import type { Product } from '../api/types';
import StatusBadge from '../components/StatusBadge';
import SelectField from '../components/SelectField';
import { CATEGORIES, ROOMS } from '../data/categories';
import { colors } from '../theme/colors';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'Search'>;

type StatusFilter = '' | 'ok' | 'warning' | 'expired';

const STATUS_OPTIONS: { value: StatusFilter; label: string }[] = [
  { value: '', label: 'כל הסטטוסים' },
  { value: 'ok', label: 'באחריות' },
  { value: 'warning', label: 'עומד לפוג' },
  { value: 'expired', label: 'פג תוקף' },
];

const ALL_OPTION = 'הכל';

export default function SearchScreen({ navigation }: Props) {
  const [query, setQuery] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [category, setCategory] = useState('');
  const [room, setRoom] = useState('');
  const [status, setStatus] = useState<StatusFilter>('');
  const [priceMin, setPriceMin] = useState('');
  const [priceMax, setPriceMax] = useState('');
  const [results, setResults] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const hasActiveFilter = Boolean(category || room || status || priceMin || priceMax);

  useEffect(() => {
    const q = query.trim();
    if (!q && !hasActiveFilter) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- resetting results when the query/filters are cleared, not derived state
      setResults([]);
      setSearched(false);
      return;
    }
    setLoading(true);
    const timeout = setTimeout(() => {
      api
        .listProducts({
          q: q || undefined,
          category: category || undefined,
          room: room || undefined,
          status: status || undefined,
          price_min: priceMin ? Number(priceMin) : undefined,
          price_max: priceMax ? Number(priceMax) : undefined,
        })
        .then((data) => {
          setResults(data);
          setSearched(true);
        })
        .finally(() => setLoading(false));
    }, 300);
    return () => clearTimeout(timeout);
  }, [query, category, room, status, priceMin, priceMax, hasActiveFilter]);

  return (
    <View style={styles.container}>
      <View style={styles.searchRow}>
        <TextInput
          style={styles.input}
          placeholder="🔍 חפש מוצר..."
          value={query}
          onChangeText={setQuery}
          autoFocus
        />
        <TouchableOpacity
          style={[styles.filterToggle, hasActiveFilter && styles.filterToggleActive]}
          onPress={() => setShowFilters((v) => !v)}
        >
          <Text style={styles.filterToggleText}>🔧 סינון</Text>
        </TouchableOpacity>
      </View>

      {showFilters ? (
        <View style={styles.filtersPanel}>
          <SelectField
            label="קטגוריה"
            value={category || ALL_OPTION}
            options={[ALL_OPTION, ...CATEGORIES]}
            onChange={(v) => setCategory(v === ALL_OPTION ? '' : v)}
          />
          <SelectField
            label="חדר"
            value={room || ALL_OPTION}
            options={[ALL_OPTION, ...ROOMS]}
            onChange={(v) => setRoom(v === ALL_OPTION ? '' : v)}
          />
          <View style={styles.statusRow}>
            {STATUS_OPTIONS.map((opt) => (
              <TouchableOpacity
                key={opt.value}
                style={[styles.statusChip, status === opt.value && styles.statusChipActive]}
                onPress={() => setStatus(opt.value)}
              >
                <Text
                  style={[
                    styles.statusChipText,
                    status === opt.value && styles.statusChipTextActive,
                  ]}
                >
                  {opt.label}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
          <View style={styles.priceRow}>
            <TextInput
              style={styles.priceInput}
              placeholder="מחיר מינימלי"
              keyboardType="numeric"
              value={priceMin}
              onChangeText={setPriceMin}
            />
            <TextInput
              style={styles.priceInput}
              placeholder="מחיר מקסימלי"
              keyboardType="numeric"
              value={priceMax}
              onChangeText={setPriceMax}
            />
          </View>
        </View>
      ) : null}

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
  searchRow: { flexDirection: 'row', gap: 8 },
  input: {
    flex: 1,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 16,
    textAlign: 'right',
  },
  filterToggle: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 14,
    justifyContent: 'center',
  },
  filterToggleActive: { borderColor: colors.primary },
  filterToggleText: { fontSize: 14, color: colors.text },
  filtersPanel: {
    backgroundColor: colors.surface,
    borderRadius: 12,
    padding: 14,
    gap: 12,
  },
  statusRow: { flexDirection: 'row-reverse', gap: 8, flexWrap: 'wrap' },
  statusChip: {
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 999,
    backgroundColor: colors.background,
    borderWidth: 1,
    borderColor: colors.border,
  },
  statusChipActive: { backgroundColor: colors.primary, borderColor: colors.primary },
  statusChipText: { fontSize: 13, color: colors.text },
  statusChipTextActive: { color: colors.primaryText, fontWeight: '600' },
  priceRow: { flexDirection: 'row-reverse', gap: 8 },
  priceInput: {
    flex: 1,
    backgroundColor: colors.background,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
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
