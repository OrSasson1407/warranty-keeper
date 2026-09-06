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
          autoCapitalize="none"
          keyboardType="email-address"
          value={email}
          onChangeText={setEmail}
        />
        <TextInput
          style={styles.input}
          placeholder="סיסמה"
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
    fontSize: 24,
    fontWeight: '700',
    color: colors.text,
    marginBottom: 16,
    textAlign: 'right',
  },
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
  error: { color: colors.danger, textAlign: 'right' },
  button: {
    backgroundColor: colors.primary,
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 8,
  },
  buttonText: { color: colors.primaryText, fontSize: 16, fontWeight: '600' },
  googleButton: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 8,
  },
  googleButtonText: { color: colors.text, fontSize: 15, fontWeight: '600' },
  link: { color: colors.primary, textAlign: 'center', marginTop: 8 },
});
