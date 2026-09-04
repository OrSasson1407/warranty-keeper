import { useEffect, useState } from 'react';
import { ActivityIndicator, Alert, Share, StyleSheet, Switch, Text, TouchableOpacity, View } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api } from '../api/client';
import type { Household } from '../api/types';
import { useAuth } from '../context/AuthContext';
import { colors } from '../theme/colors';
import { registerForExpiryPush } from '../notifications/registerPush';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'Settings'>;

const PUSH_PREF_KEY = 'wk_push_enabled';

export default function SettingsScreen(_props: Props) {
  const { user, logout } = useAuth();
  const [household, setHousehold] = useState<Household | null>(null);
  const [pushEnabled, setPushEnabled] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      const [h, pref] = await Promise.all([api.getMyHousehold(), AsyncStorage.getItem(PUSH_PREF_KEY)]);
      setHousehold(h);
      setPushEnabled(pref === 'true');
      setLoading(false);
    })();
  }, []);

  const onTogglePush = async (value: boolean) => {
    setPushEnabled(value);
    await AsyncStorage.setItem(PUSH_PREF_KEY, String(value));
    if (value) {
      const ok = await registerForExpiryPush();
      if (!ok) {
        Alert.alert('לא ניתן להפעיל התראות', 'נדרשת הרשאת התראות במכשיר.');
        setPushEnabled(false);
        await AsyncStorage.setItem(PUSH_PREF_KEY, 'false');
      }
    }
  };

  const onShareInvite = () => {
    if (!household) return;
    Share.share({
      message: `הצטרפו אליי ל-WarrantyKeeper! קוד ההזמנה למשק הבית "${household.name}": ${household.invite_code}`,
    });
  };

  if (loading || !household) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  const canInvite = household.members.length < 2;

  return (
    <View style={styles.container}>
      <Text style={styles.sectionTitle}>משק בית: "{household.name}"</Text>
      <View style={styles.card}>
        {household.members.map((m) => (
          <Text key={m.id} style={styles.memberRow}>
            {m.full_name} {m.id === user?.id ? '(את/ה)' : ''}
          </Text>
        ))}
        {canInvite ? (
          <TouchableOpacity style={styles.inviteButton} onPress={onShareInvite}>
            <Text style={styles.inviteButtonText}>+ הזמן בן/בת משפחה (קוד: {household.invite_code})</Text>
          </TouchableOpacity>
        ) : (
          <Text style={styles.fullNote}>משק הבית מלא (מקסימום 2 חברים)</Text>
        )}
      </View>

      <Text style={styles.sectionTitle}>התראות</Text>
      <View style={styles.card}>
        <View style={styles.switchRow}>
          <Switch value={pushEnabled} onValueChange={onTogglePush} />
          <Text style={styles.switchLabel}>30 ימים לפני תפוגת אחריות</Text>
        </View>
      </View>

      <TouchableOpacity style={styles.logoutButton} onPress={logout}>
        <Text style={styles.logoutText}>התנתקות</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background, padding: 20, gap: 8 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.background },
  sectionTitle: { fontSize: 14, fontWeight: '700', color: colors.textMuted, textAlign: 'right', marginTop: 16 },
  card: { backgroundColor: colors.surface, borderRadius: 12, padding: 14, gap: 10 },
  memberRow: { fontSize: 16, color: colors.text, textAlign: 'right' },
  inviteButton: { paddingTop: 8, borderTopWidth: 1, borderTopColor: colors.border },
  inviteButtonText: { color: colors.primary, textAlign: 'right', fontWeight: '600' },
  fullNote: { color: colors.textMuted, textAlign: 'right', fontSize: 13 },
  switchRow: { flexDirection: 'row-reverse', alignItems: 'center', gap: 10 },
  switchLabel: { fontSize: 15, color: colors.text },
  logoutButton: { marginTop: 32, alignItems: 'center', padding: 12 },
  logoutText: { color: colors.danger, fontSize: 15, fontWeight: '600' },
});
