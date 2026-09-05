import { api, ApiError } from './client';
import { API_BASE_URL } from './config';
import * as tokenStore from './tokenStore';

// Mocked with an explicit factory (not automock, which would still evaluate
// the real module) so each test can control exactly what token is "stored"
// without touching AsyncStorage at all (see jest.setup.js for the global
// AsyncStorage mock other tests rely on instead).
jest.mock('./tokenStore', () => ({
  getAccessToken: jest.fn(),
  getRefreshToken: jest.fn(),
  setTokens: jest.fn(),
  setAccessToken: jest.fn(),
  clearTokens: jest.fn(),
}));

const mockedTokenStore = tokenStore as jest.Mocked<typeof tokenStore>;

function jsonResponse(status: number, body: unknown) {
  return {
    status,
    ok: status >= 200 && status < 300,
    text: async () => JSON.stringify(body),
    json: async () => body,
  };
}

beforeEach(() => {
  globalThis.fetch = jest.fn();
  mockedTokenStore.getAccessToken.mockResolvedValue(null);
  mockedTokenStore.getRefreshToken.mockResolvedValue(null);
  mockedTokenStore.setAccessToken.mockResolvedValue(undefined);
});

afterEach(() => {
  jest.resetAllMocks();
});

describe('api.login', () => {
  it('POSTs credentials and returns the parsed response', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce(
      jsonResponse(200, { access_token: 'a', refresh_token: 'b', user: { id: '1' } }),
    );

    const result = await api.login('a@example.com', 'pw');
    expect(result.access_token).toBe('a');

    const [url, options] = (globalThis.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe(`${API_BASE_URL}/auth/login`);
    expect(options.method).toBe('POST');
    expect(JSON.parse(options.body)).toEqual({ email: 'a@example.com', password: 'pw' });
  });
});

describe('request error handling', () => {
  it('throws ApiError with the server-provided message on failure', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValue(jsonResponse(400, { error: 'bad request' }));
    await expect(api.login('x', 'y')).rejects.toThrow(ApiError);
    await expect(api.login('x', 'y')).rejects.toThrow('bad request');
  });

  it('falls back to a generic message when the body has no error field', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValue(jsonResponse(500, {}));
    await expect(api.login('x', 'y')).rejects.toThrow('Request failed with status 500');
  });

  it('exposes the HTTP status on the thrown error', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValue(jsonResponse(404, { error: 'not found' }));
    try {
      await api.getProduct('missing-id');
      throw new Error('expected api.getProduct to reject');
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(404);
    }
  });
});

describe('authorization header', () => {
  it('omits the Authorization header when no token is stored', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce(jsonResponse(200, {}));
    await api.getMyHousehold();
    const [, options] = (globalThis.fetch as jest.Mock).mock.calls[0];
    expect(options.headers.Authorization).toBeUndefined();
  });

  it('includes a Bearer header when a token is stored', async () => {
    mockedTokenStore.getAccessToken.mockResolvedValue('my-token');
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce(jsonResponse(200, {}));
    await api.getMyHousehold();
    const [, options] = (globalThis.fetch as jest.Mock).mock.calls[0];
    expect(options.headers.Authorization).toBe('Bearer my-token');
  });
});

describe('401 refresh-and-retry', () => {
  it('refreshes the access token once and retries the original request', async () => {
    mockedTokenStore.getAccessToken
      .mockResolvedValueOnce('expired-token')
      .mockResolvedValueOnce('new-token');
    mockedTokenStore.getRefreshToken.mockResolvedValue('a-refresh-token');

    (globalThis.fetch as jest.Mock)
      .mockResolvedValueOnce(jsonResponse(401, { error: 'expired' })) // original request
      .mockResolvedValueOnce(jsonResponse(200, { access_token: 'new-token' })) // refresh call
      .mockResolvedValueOnce(jsonResponse(200, { id: 'h1' })); // retried request

    const result = await api.getMyHousehold();
    expect(result).toEqual({ id: 'h1' });
    expect(globalThis.fetch).toHaveBeenCalledTimes(3);
    expect(mockedTokenStore.setAccessToken).toHaveBeenCalledWith('new-token');
  });

  it('does not loop forever if the retried request is also a 401', async () => {
    mockedTokenStore.getAccessToken.mockResolvedValue('expired-token');
    mockedTokenStore.getRefreshToken.mockResolvedValue('a-refresh-token');

    (globalThis.fetch as jest.Mock)
      .mockResolvedValueOnce(jsonResponse(401, { error: 'expired' }))
      .mockResolvedValueOnce(jsonResponse(200, { access_token: 'still-bad' }))
      .mockResolvedValueOnce(jsonResponse(401, { error: 'still expired' }));

    await expect(api.getMyHousehold()).rejects.toThrow('still expired');
    expect(globalThis.fetch).toHaveBeenCalledTimes(3); // original, refresh attempt, one retry — not more
  });

  it('gives up immediately if there is no refresh token to use', async () => {
    mockedTokenStore.getAccessToken.mockResolvedValue('expired-token');
    mockedTokenStore.getRefreshToken.mockResolvedValue(null);
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce(jsonResponse(401, { error: 'expired' }));

    await expect(api.getMyHousehold()).rejects.toThrow('expired');
    expect(globalThis.fetch).toHaveBeenCalledTimes(1);
  });
});

describe('resolveWarranty', () => {
  it('builds the query string from category, brand, and purchase date', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce(
      jsonResponse(200, {
        warranty_expires_at: '2027-01-01',
        duration_months: 12,
        uncertain: false,
        source: 'default',
      }),
    );

    await api.resolveWarranty('מזגן', 'טורנדו', '2026-01-01');

    const [url] = (globalThis.fetch as jest.Mock).mock.calls[0];
    const parsed = new URL(url);
    expect(parsed.pathname).toBe('/warranty-rules/resolve');
    expect(parsed.searchParams.get('category')).toBe('מזגן');
    expect(parsed.searchParams.get('brand')).toBe('טורנדו');
    expect(parsed.searchParams.get('purchase_date')).toBe('2026-01-01');
  });
});

describe('listProducts', () => {
  it('omits the q param entirely when no search term is given', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce(jsonResponse(200, []));
    await api.listProducts();
    const [url] = (globalThis.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe(`${API_BASE_URL}/products`);
  });

  it('URL-encodes the search term', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce(jsonResponse(200, []));
    const term = 'מזגן & אחריות';
    await api.listProducts(term);
    const [url] = (globalThis.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe(`${API_BASE_URL}/products?q=${encodeURIComponent(term)}`);
  });
});

describe('uploadReceipt', () => {
  it('sends a multipart FormData body without a manually-set Content-Type', async () => {
    (globalThis.fetch as jest.Mock).mockResolvedValueOnce(jsonResponse(200, { receipt_id: 'r1' }));
    await api.uploadReceipt({ uri: 'file:///x.jpg', name: 'x.jpg', type: 'image/jpeg' });

    const [, options] = (globalThis.fetch as jest.Mock).mock.calls[0];
    expect(options.body).toBeInstanceOf(FormData);
    expect(options.headers['Content-Type']).toBeUndefined();
  });
});
