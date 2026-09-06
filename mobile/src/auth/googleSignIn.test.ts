import { extractIdToken, isGoogleSignInConfigured } from './googleSignIn';

const ENV_KEYS = [
  'EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID',
  'EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID',
  'EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID',
] as const;

function clearGoogleEnv() {
  for (const key of ENV_KEYS) delete process.env[key];
}

describe('isGoogleSignInConfigured', () => {
  const original = { ...process.env };

  afterEach(() => {
    clearGoogleEnv();
    Object.assign(process.env, original);
  });

  it('returns false when no client ID env var is set', () => {
    clearGoogleEnv();
    expect(isGoogleSignInConfigured()).toBe(false);
  });

  it('returns true when the web client ID is set', () => {
    clearGoogleEnv();
    process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID = 'web-id.apps.googleusercontent.com';
    expect(isGoogleSignInConfigured()).toBe(true);
  });

  it('returns true when only the iOS or Android client ID is set', () => {
    clearGoogleEnv();
    process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID = 'ios-id.apps.googleusercontent.com';
    expect(isGoogleSignInConfigured()).toBe(true);
  });
});

describe('extractIdToken', () => {
  it('returns the id_token from a successful response', () => {
    const result = extractIdToken({
      type: 'success',
      params: { id_token: 'abc123' },
    } as any);
    expect(result).toBe('abc123');
  });

  it('returns null when the response is null', () => {
    expect(extractIdToken(null)).toBeNull();
  });

  it('returns null for a cancelled response', () => {
    expect(extractIdToken({ type: 'cancel' } as any)).toBeNull();
  });

  it('returns null for a dismissed response', () => {
    expect(extractIdToken({ type: 'dismiss' } as any)).toBeNull();
  });

  it('returns null for an error response', () => {
    expect(extractIdToken({ type: 'error', params: {} } as any)).toBeNull();
  });

  it('returns null when a success response is missing the id_token param', () => {
    expect(extractIdToken({ type: 'success', params: {} } as any)).toBeNull();
  });
});
