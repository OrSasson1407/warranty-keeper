import { useEffect, useState } from 'react';
import { FlatList, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';

import { api } from '../api/client';
import type { Product } from '../api/types';
import ProductCard from '../components/ProductCard';
import SelectField from '../components/SelectField';
import { ProductCardSkeleton } from '../components/Skeleton';
import { CATEGORIES, ROOMS } from '../data/categories';
import { colors } from '../theme/colors';
import { typography } from '../theme/typography';
import type { AppTabScreenProps } from '../navigation/types';

type Props = AppTabScreenProps<'SearchTab'>;

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

  const activeFilterCount = [category, room, status, priceMin, priceMax].filter(Boolean).length;
  const hasActiveFilter = activeFilterCount > 0;

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
          placeholderTextColor={colors.outline}
          value={query}
          onChangeText={setQuery}
          autoFocus
        />
        <TouchableOpacity
          style={[styles.filterToggle, hasActiveFilter && styles.filterToggleActive]}
          onPress={() => setShowFilters((v) => !v)}
        >
          <Text style={styles.filterToggleText}>🔧 סינון</Text>
          {activeFilterCount > 0 ? (
            <View style={styles.filterCountBadge}>
              <Text style={styles.filterCountBadgeText}>{activeFilterCount}</Text>
            </View>
          ) : null}
        </TouchableOpacity>
      </View>

      {showFilters ? (
        <View style={styles.filtersPanel}>
          <View style={styles.filtersAccent} />
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
              placeholderTextColor={colors.outline}
              keyboardType="numeric"
              value={priceMin}
              onChangeText={setPriceMin}
            />
            <TextInput
              style={styles.priceInput}
              placeholder="מחיר מקסימלי"
              placeholderTextColor={colors.outline}
              keyboardType="numeric"
              value={priceMax}
              onChangeText={setPriceMax}
            />
          </View>
        </View>
      ) : null}

      {searched && !loading ? (
        <Text style={styles.resultsLabel}>תוצאות ({results.length}):</Text>
      ) : null}

      <FlatList
        data={results}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.list}
        ListEmptyComponent={
          loading ? (
            <View style={{ gap: 12 }}>
              <ProductCardSkeleton />
              <ProductCardSkeleton />
            </View>
          ) : searched ? (
            <View style={styles.empty}>
              <Text style={styles.emptyIcon}>🔍</Text>
              <Text style={styles.emptyText}>לא נמצאו מוצרים תואמים.</Text>
            </View>
          ) : null
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
  container: { flex: 1, backgroundColor: colors.background, padding: 20, gap: 12 },
  searchRow: { flexDirection: 'row', gap: 8 },
  input: {
    flex: 1,
    backgroundColor: colors.surfaceContainerLowest,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: 14,
    paddingVertical: 12,
    color: colors.text,
    ...typography.bodyLg,
    textAlign: 'right',
  },
  filterToggle: {
    backgroundColor: colors.surfaceContainerLowest,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: 14,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  filterToggleActive: { borderColor: colors.primary },
  filterToggleText: { ...typography.bodyMd, color: colors.text },
  filterCountBadge: {
    minWidth: 18,
    height: 18,
    paddingHorizontal: 4,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  filterCountBadgeText: { ...typography.labelSm, color: colors.primaryText, textTransform: 'none' },
  filtersPanel: {
    backgroundColor: colors.surfaceContainer,
    padding: 14,
    gap: 12,
    position: 'relative',
  },
  filtersAccent: { position: 'absolute', top: 0, right: 0, left: 0, height: 2, backgroundColor: colors.primary },
  statusRow: { flexDirection: 'row-reverse', gap: 8, flexWrap: 'wrap' },
  statusChip: {
    paddingHorizontal: 12,
    paddingVertical: 8,
    backgroundColor: colors.surfaceContainerLowest,
    borderWidth: 1,
    borderColor: colors.border,
  },
  statusChipActive: { backgroundColor: colors.primary, borderColor: colors.primary },
  statusChipText: { ...typography.bodySm, color: colors.text },
  statusChipTextActive: { color: colors.primaryText, fontFamily: typography.bodySm.fontFamily },
  priceRow: { flexDirection: 'row-reverse', gap: 8 },
  priceInput: {
    flex: 1,
    backgroundColor: colors.surfaceContainerLowest,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: 12,
    paddingVertical: 10,
    color: colors.text,
    ...typography.bodyMd,
    textAlign: 'right',
  },
  resultsLabel: { ...typography.bodySm, color: colors.textMuted, textAlign: 'right' },
  list: { gap: 12 },
  empty: { paddingTop: 40, alignItems: 'center', gap: 12 },
  emptyIcon: { fontSize: 32 },
  emptyText: { ...typography.bodyMd, color: colors.textMuted, textAlign: 'center' },
});
