import { extractAuthCode, extractCodeVerifier } from './gmailConnect';

describe('extractAuthCode', () => {
  it('returns the code from a successful response', () => {
    expect(extractAuthCode({ type: 'success', params: { code: 'abc123' } } as any)).toBe('abc123');
  });

  it('returns null when the response is null', () => {
    expect(extractAuthCode(null)).toBeNull();
  });

  it('returns null for a cancelled response', () => {
    expect(extractAuthCode({ type: 'cancel' } as any)).toBeNull();
  });

  it('returns null when a success response is missing the code param', () => {
    expect(extractAuthCode({ type: 'success', params: {} } as any)).toBeNull();
  });
});

describe('extractCodeVerifier', () => {
  it('returns the request codeVerifier when present', () => {
    expect(extractCodeVerifier({ codeVerifier: 'verifier-1' } as any)).toBe('verifier-1');
  });

  it('returns null when the request is null', () => {
    expect(extractCodeVerifier(null)).toBeNull();
  });

  it('returns null when the request has no codeVerifier', () => {
    expect(extractCodeVerifier({} as any)).toBeNull();
  });
});
