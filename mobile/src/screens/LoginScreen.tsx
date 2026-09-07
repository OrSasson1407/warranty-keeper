import { useEffect, useState } from 'react';
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
import { extractIdToken, isGoogleSignInConfigured, useGoogleSignIn } from '../auth/googleSignIn';
import { colors } from '../theme/colors';
import { fonts, typography } from '../theme/typography';
import { ApiError } from '../api/client';
import type { AuthStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AuthStackParamList, 'Login'>;

export default function LoginScreen({ navigation }: Props) {
  const { login, loginWithGoogle } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [googleRequest, googleResponse, promptGoogleSignIn] = useGoogleSignIn();

  const onSubmit = async () => {
    setError(null);
    setSubmitting(true);
    try {
      await login(email.trim(), password);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'ההתחברות נכשלה, נסו שוב');
    } finally {
      setSubmitting(false);
    }
  };

  useEffect(() => {
    const idToken = extractIdToken(googleResponse);
    if (!idToken) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- resetting error/loading state for the async call below, not derived state
    setError(null);
    setSubmitting(true);
    loginWithGoogle(idToken)
      .catch((e) => setError(e instanceof ApiError ? e.message : 'ההתחברות עם Google נכשלה'))
      .finally(() => setSubmitting(false));
  }, [googleResponse, loginWithGoogle]);

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <ScrollView contentContainerStyle={styles.scroll}>
        <Text style={styles.title}>התחברות</Text>

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
          placeholder="סיסמה"
          placeholderTextColor={colors.outline}
          secureTextEntry
          value={password}
          onChangeText={setPassword}
        />

        {error ? <Text style={styles.error}>{error}</Text> : null}

        <TouchableOpacity style={styles.button} onPress={onSubmit} disabled={submitting}>
          {submitting ? (
            <ActivityIndicator color={colors.primaryText} />
          ) : (
            <Text style={styles.buttonText}>התחברות</Text>
          )}
        </TouchableOpacity>

        {isGoogleSignInConfigured() ? (
          <TouchableOpacity
            style={styles.googleButton}
            onPress={() => promptGoogleSignIn()}
            disabled={!googleRequest || submitting}
          >
            <Text style={styles.googleButtonText}>התחברות עם Google</Text>
          </TouchableOpacity>
        ) : null}

        <TouchableOpacity onPress={() => navigation.navigate('Register')}>
          <Text style={styles.link}>אין לי חשבון — הרשמה</Text>
        </TouchableOpacity>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  scroll: { padding: 24, gap: 12, flexGrow: 1, justifyContent: 'center' },
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
  googleButton: {
    backgroundColor: colors.surfaceContainer,
    borderWidth: 1,
    borderColor: colors.border,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 8,
  },
  googleButtonText: { color: colors.text, ...typography.bodyMd, fontFamily: fonts.bodyMdSemiBold },
  link: { color: colors.primary, textAlign: 'center', marginTop: 8, ...typography.bodyMd },
});
