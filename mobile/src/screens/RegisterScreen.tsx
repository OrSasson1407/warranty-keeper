import { useState } from 'react';
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { useAuth } from '../context/AuthContext';
import { colors } from '../theme/colors';
import { fonts, typography } from '../theme/typography';
import { ApiError } from '../api/client';
import type { AuthStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AuthStackParamList, 'Register'>;

export default function RegisterScreen({ navigation, route }: Props) {
  const { register } = useAuth();
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [inviteCode, setInviteCode] = useState(route.params?.inviteCode ?? '');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async () => {
    setError(null);
    if (!fullName || !email || !password) {
      setError('נא למלא את כל השדות');
      return;
    }
    if (password.length < 8) {
      setError('הסיסמה חייבת להכיל לפחות 8 תווים');
      return;
    }
    setSubmitting(true);
    try {
      await register(email.trim(), password, fullName.trim(), inviteCode.trim() || undefined);
      // Navigation to the app stack happens automatically once `user` is set.
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'ההרשמה נכשלה, נסו שוב');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <ScrollView contentContainerStyle={styles.scroll}>
        <Text style={styles.title}>יצירת חשבון</Text>

        <TextInput
          style={styles.input}
          placeholder="שם מלא"
          placeholderTextColor={colors.outline}
          value={fullName}
          onChangeText={setFullName}
        />
        <TextInput
          style={styles.input}
          placeholder="אימייל"
          placeholderTextColor={colors.outline}
          autoCapitalize="none"
          keyboardType="email-address"
          value={email}
          onChangeText={setEmail}
        />
        <TextInput
          style={styles.input}
          placeholder="סיסמה (8 תווים לפחות)"
          placeholderTextColor={colors.outline}
          secureTextEntry
          value={password}
          onChangeText={setPassword}
        />
        <TextInput
          style={styles.input}
          placeholder="קוד הזמנה (אופציונלי, אם מצטרפים למשק בית קיים)"
          placeholderTextColor={colors.outline}
          autoCapitalize="characters"
          value={inviteCode}
          onChangeText={setInviteCode}
        />

        {error ? <Text style={styles.error}>{error}</Text> : null}

        <TouchableOpacity style={styles.button} onPress={onSubmit} disabled={submitting}>
          {submitting ? (
            <ActivityIndicator color={colors.primaryText} />
          ) : (
            <Text style={styles.buttonText}>הרשמה</Text>
          )}
        </TouchableOpacity>

        <TouchableOpacity onPress={() => navigation.navigate('Login')}>
          <Text style={styles.link}>כבר יש לי חשבון — התחברות</Text>
        </TouchableOpacity>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  scroll: { padding: 24, gap: 12 },
  title: {
    ...typography.headlineLg,
    color: colors.text,
    marginBottom: 16,
    textAlign: 'right',
  },
  input: {
    backgroundColor: colors.surfaceContainerLowest,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: 14,
    paddingVertical: 12,
    color: colors.text,
    ...typography.bodyLg,
    textAlign: 'right',
  },
  error: { color: colors.error, textAlign: 'right', ...typography.bodySm },
  button: {
    backgroundColor: colors.primary,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 8,
  },
  buttonText: { color: colors.primaryText, ...typography.headlineSm, fontFamily: fonts.headlineSm },
  link: { color: colors.primary, textAlign: 'center', marginTop: 8, ...typography.bodyMd },
});
