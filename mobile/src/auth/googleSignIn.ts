import type { AuthSessionResult } from 'expo-auth-session';
import * as Google from 'expo-auth-session/providers/google';
import * as WebBrowser from 'expo-web-browser';

// Required once per app for the auth session to close its browser tab
// correctly after redirecting back. Expo's own docs example calls this at
// module scope, not inside a component.
WebBrowser.maybeCompleteAuthSession();

// expo-auth-session's Google provider throws synchronously (inside a
// useMemo, on every render) if no client ID is configured for the current
// platform -- it only checks presence, not validity, so a placeholder
// string is enough to avoid the crash without a real Google Cloud project.
// isGoogleSignInConfigured() is what actually gates whether the sign-in
// button appears.
const PLACEHOLDER_CLIENT_ID = 'google-oauth-not-configured';

export function isGoogleSignInConfigured(): boolean {
  return Boolean(
    process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID ||
    process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID ||
    process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  );
}

// expo-auth-session/providers/google's config type is marked @deprecated in
// SDK 57 in favor of a generic AuthRequest against Google's own discovery
// document (see docs.expo.dev/guides/google-authentication), but it's still
// functional and is by far the smallest amount of code to get an id_token
// out of Google Sign-In -- which is exactly what POST /auth/google expects.
// Revisit if a future SDK actually removes it.
export function useGoogleSignIn() {
  return Google.useIdTokenAuthRequest({
    webClientId: process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID || PLACEHOLDER_CLIENT_ID,
    iosClientId: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID || PLACEHOLDER_CLIENT_ID,
    androidClientId: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID || PLACEHOLDER_CLIENT_ID,
  });
}

export function extractIdToken(response: AuthSessionResult | null): string | null {
  if (response?.type !== 'success') return null;
  return response.params.id_token ?? null;
}
