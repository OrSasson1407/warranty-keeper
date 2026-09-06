import { makeRedirectUri, ResponseType, useAuthRequest } from 'expo-auth-session';
import type { AuthRequest, AuthSessionResult } from 'expo-auth-session';

// Same placeholder-client-ID trick as googleSignIn.ts: expo-auth-session
// throws synchronously if no client ID is configured for the platform, so a
// truthy placeholder avoids the crash. isGoogleSignInConfigured() (shared
// with sign-in, since both features reuse the one OAuth client) is what
// actually gates whether the "Connect Gmail" button appears.
const PLACEHOLDER_CLIENT_ID = 'google-oauth-not-configured';

const GMAIL_READONLY_SCOPE = 'https://www.googleapis.com/auth/gmail.readonly';

const discovery = {
  authorizationEndpoint: 'https://accounts.google.com/o/oauth2/v2/auth',
  tokenEndpoint: 'https://oauth2.googleapis.com/token',
  revocationEndpoint: 'https://oauth2.googleapis.com/revoke',
};

// Unlike sign-in (which only verifies an id_token), connecting Gmail needs
// offline, revocable access: an authorization *code* that our backend
// exchanges for a refresh token (using the client secret, which must never
// live in this app). access_type=offline + prompt=consent forces Google to
// hand back a refresh token even if the user connected once before.
export function useGmailConnectRequest() {
  return useAuthRequest(
    {
      clientId: process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID || PLACEHOLDER_CLIENT_ID,
      scopes: [GMAIL_READONLY_SCOPE],
      responseType: ResponseType.Code,
      redirectUri: makeRedirectUri(),
      extraParams: { access_type: 'offline', prompt: 'consent' },
    },
    discovery,
  );
}

export function extractAuthCode(response: AuthSessionResult | null): string | null {
  if (response?.type !== 'success') return null;
  return response.params.code ?? null;
}

// The PKCE code_verifier lives on the request object (generated when the
// hook builds it), not the response -- the backend needs it alongside the
// code to complete the token exchange.
export function extractCodeVerifier(request: AuthRequest | null): string | null {
  return request?.codeVerifier ?? null;
}
