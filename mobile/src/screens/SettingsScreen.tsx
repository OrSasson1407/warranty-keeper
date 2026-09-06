import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Share,
  StyleSheet,
  Switch,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { api, ApiError } from '../api/client';
import type { GmailStatus, Household } from '../api/types';
import { useAuth } from '../context/AuthContext';
import { colors } from '../theme/colors';
import { registerForExpiryPush } from '../notifications/registerPush';
import { extractAuthCode, extractCodeVerifier, useGmailConnectRequest } from '../auth/gmailConnect';
import { isGoogleSignInConfigured } from '../auth/googleSignIn';
import type { AppStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'Settings'>;

const PUSH_PREF_KEY = 'wk_push_enabled';

export default function SettingsScreen({ navigation }: Props) {
  const { user, logout } = useAuth();
  const [household, setHousehold] = useState<Household | null>(null);
  const [pushEnabled, setPushEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  const [upgrading, setUpgrading] = useState(false);
  const [gmailStatus, setGmailStatus] = useState<GmailStatus | null>(null);
  const [gmailBusy, setGmailBusy] = useState(false);
  const [gmailRequest, gmailResponse, promptGmailConnect] = useGmailConnectRequest();

  useEffect(() => {
    (async () => {
      const [h, pref] = await Promise.all([
        api.getMyHousehold(),
        AsyncStorage.getItem(PUSH_PREF_KEY),
      ]);
      setHousehold(h);
      setPushEnabled(pref === 'true');
      setLoading(false);
    })();
    if (isGoogleSignInConfigured()) {
      api
        .gmailStatus()
        .then(setGmailStatus)
        .catch(() => {});
    }
  }, []);

  useEffect(() => {
    const code = extractAuthCode(gmailResponse);
    if (!code) return;
    const codeVerifier = extractCodeVerifier(gmailRequest);
    // eslint-disable-next-line react-hooks/set-state-in-effect -- resetting loading state for the async call below, not derived state
    setGmailBusy(true);
    api
      .connectGmail(code, gmailRequest?.redirectUri ?? '', codeVerifier)
      .then((status) => {
        setGmailStatus(status);
        Alert.alert('Gmail מחובר', 'נסרוק את תיבת הדואר שלכם לאיתור קבלות בזמן הקרוב.');
      })
      .catch((e) => {
        Alert.alert('שגיאה', e instanceof ApiError ? e.message : 'לא הצלחנו לחבר את Gmail');
      })
      .finally(() => setGmailBusy(false));
  }, [gmailResponse, gmailRequest]);

  const onDisconnectGmail = async () => {
    setGmailBusy(true);
    try {
      await api.disconnectGmail();
      setGmailStatus({ connected: false, last_scan_at: null });
    } catch {
      Alert.alert('שגיאה', 'לא הצלחנו לנתק את Gmail כרגע, נסו שוב.');
    } finally {
      setGmailBusy(false);
    }
  };

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

  const onUpgrade = async () => {
    setUpgrading(true);
    try {
      await api.upgradeHousehold();
      const refreshed = await api.getMyHousehold();
      setHousehold(refreshed);
      Alert.alert('שודרג בהצלחה', 'משק הבית שלכם עבר ל-Premium — ללא הגבלת מוצרים.');
    } catch {
      Alert.alert('שגיאה', 'לא הצלחנו לשדרג כרגע, נסו שוב מאוחר יותר.');
    } finally {
      setUpgrading(false);
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
      <Text style={styles.sectionTitle}>{`משק בית: "${household.name}"`}</Text>
      <View style={styles.card}>
        {household.members.map((m) => (
          <Text key={m.id} style={styles.memberRow}>
            {m.full_name} {m.id === user?.id ? '(את/ה)' : ''}
          </Text>
        ))}
        {canInvite ? (
          <TouchableOpacity style={styles.inviteButton} onPress={onShareInvite}>
            <Text style={styles.inviteButtonText}>
              + הזמן בן/בת משפחה (קוד: {household.invite_code})
            </Text>
          </TouchableOpacity>
        ) : (
          <Text style={styles.fullNote}>משק הבית מלא (מקסימום 2 חברים)</Text>
        )}
      </View>

      <Text style={styles.sectionTitle}>מנוי</Text>
      <View style={styles.card}>
        {household.tier === 'premium' ? (
          <Text style={styles.premiumBadge}>⭐ Premium — ללא הגבלת מוצרים</Text>
        ) : (
          <>
            <Text style={styles.memberRow}>תוכנית חינמית — עד 20 מוצרים</Text>
            <TouchableOpacity style={styles.upgradeButton} onPress={onUpgrade} disabled={upgrading}>
              {upgrading ? (
                <ActivityIndicator color={colors.primaryText} />
              ) : (
                <Text style={styles.upgradeButtonText}>שדרגו ל-Premium</Text>
              )}
            </TouchableOpacity>
          </>
        )}
      </View>

      {isGoogleSignInConfigured() ? (
        <>
          <Text style={styles.sectionTitle}>ייבוא קבלות מ-Gmail</Text>
          <View style={styles.card}>
            <Text style={styles.gmailExplainer}>
              חיבור אופציונלי: נסרוק אך ורק מיילים מחנויות ידועות (כגון Amazon, KSP, איקאה) לאיתור
              אישורי הזמנה, ולא ניגע בשאר תיבת הדואר. ניתן לנתק בכל רגע.
            </Text>
            {gmailStatus?.connected ? (
              <>
                <Text style={styles.memberRow}>מחובר: {gmailStatus.gmail_address}</Text>
                <TouchableOpacity
                  style={styles.gmailLink}
                  onPress={() => navigation.navigate('GmailReceipts')}
                >
                  <Text style={styles.inviteButtonText}>צפייה בקבלות שנמצאו</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={styles.disconnectButton}
                  onPress={onDisconnectGmail}
                  disabled={gmailBusy}
                >
                  {gmailBusy ? (
                    <ActivityIndicator color={colors.danger} />
                  ) : (
                    <Text style={styles.disconnectButtonText}>ניתוק Gmail</Text>
                  )}
                </TouchableOpacity>
              </>
            ) : (
              <TouchableOpacity
                style={styles.upgradeButton}
                onPress={() => promptGmailConnect()}
                disabled={!gmailRequest || gmailBusy}
              >
                {gmailBusy ? (
                  <ActivityIndicator color={colors.primaryText} />
                ) : (
                  <Text style={styles.upgradeButtonText}>התחברות ל-Gmail</Text>
                )}
              </TouchableOpacity>
            )}
          </View>
        </>
      ) : null}

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
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: '700',
    color: colors.textMuted,
    textAlign: 'right',
    marginTop: 16,
  },
  card: { backgroundColor: colors.surface, borderRadius: 12, padding: 14, gap: 10 },
  memberRow: { fontSize: 16, color: colors.text, textAlign: 'right' },
  inviteButton: { paddingTop: 8, borderTopWidth: 1, borderTopColor: colors.border },
  inviteButtonText: { color: colors.primary, textAlign: 'right', fontWeight: '600' },
  fullNote: { color: colors.textMuted, textAlign: 'right', fontSize: 13 },
  premiumBadge: { color: colors.statusOk, fontWeight: '700', textAlign: 'right', fontSize: 15 },
  upgradeButton: {
    backgroundColor: colors.primary,
    borderRadius: 10,
    paddingVertical: 10,
    alignItems: 'center',
  },
  upgradeButtonText: { color: colors.primaryText, fontWeight: '600', fontSize: 14 },
  gmailExplainer: { fontSize: 13, color: colors.textMuted, textAlign: 'right', lineHeight: 18 },
  gmailLink: { paddingVertical: 4 },
  disconnectButton: { alignItems: 'center', paddingVertical: 8 },
  disconnectButtonText: { color: colors.danger, fontWeight: '600', fontSize: 14 },
  switchRow: { flexDirection: 'row-reverse', alignItems: 'center', gap: 10 },
  switchLabel: { fontSize: 15, color: colors.text },
  logoutButton: { marginTop: 32, alignItems: 'center', padding: 12 },
  logoutText: { color: colors.danger, fontSize: 15, fontWeight: '600' },
});
