import { useState } from 'react';
import { FlatList, Modal, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

import { colors } from '../theme/colors';

interface Props {
  label: string;
  value: string;
  options: string[];
  onChange: (value: string) => void;
  placeholder?: string;
}

export default function SelectField({ label, value, options, onChange, placeholder }: Props) {
  const [open, setOpen] = useState(false);

  return (
    <View style={styles.wrapper}>
      <Text style={styles.label}>{label}</Text>
      <TouchableOpacity style={styles.field} onPress={() => setOpen(true)}>
        <Text style={value ? styles.valueText : styles.placeholderText}>
          {value || placeholder || 'בחר...'}
        </Text>
        <Text style={styles.chevron}>▾</Text>
      </TouchableOpacity>

      <Modal visible={open} transparent animationType="fade" onRequestClose={() => setOpen(false)}>
        <TouchableOpacity style={styles.backdrop} activeOpacity={1} onPress={() => setOpen(false)}>
          <View style={styles.sheet}>
            <FlatList
              data={options}
              keyExtractor={(item) => item}
              renderItem={({ item }) => (
                <TouchableOpacity
                  style={styles.option}
                  onPress={() => {
                    onChange(item);
                    setOpen(false);
                  }}
                >
                  <Text style={[styles.optionText, item === value && styles.optionTextActive]}>
                    {item}
                  </Text>
                </TouchableOpacity>
              )}
            />
          </View>
        </TouchableOpacity>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: { gap: 6 },
  label: { fontSize: 13, color: colors.textMuted, textAlign: 'right' },
  field: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  valueText: { fontSize: 16, color: colors.text },
  placeholderText: { fontSize: 16, color: colors.textMuted },
  chevron: { color: colors.textMuted },
  backdrop: { flex: 1, backgroundColor: 'rgba(0,0,0,0.3)', justifyContent: 'flex-end' },
  sheet: {
    backgroundColor: colors.surface,
    maxHeight: '60%',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    paddingVertical: 8,
  },
  option: { paddingVertical: 14, paddingHorizontal: 20 },
  optionText: { fontSize: 16, color: colors.text, textAlign: 'right' },
  optionTextActive: { color: colors.primary, fontWeight: '600' },
});
