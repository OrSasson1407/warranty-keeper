import { API_BASE_URL } from './config';
import * as tokenStore from './tokenStore';
import type {
  AuthResponse,
  Household,
  ManufacturerContact,
  Product,
  ProductCost,
  Receipt,
  ReceiptDraft,
  WarrantyClaim,
  WarrantyResolution,
} from './types';

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}, retry = true): Promise<T> {
  const token = await tokenStore.getAccessToken();
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string>),
  };
  if (token) headers.Authorization = `Bearer ${token}`;
  if (!(options.body instanceof FormData) && options.body) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(`${API_BASE_URL}${path}`, { ...options, headers });

  if (res.status === 401 && retry) {
    const refreshed = await tryRefresh();
    if (refreshed) return request<T>(path, options, false);
  }

  const text = await res.text();
  const data = text ? JSON.parse(text) : undefined;

  if (!res.ok) {
    throw new ApiError(res.status, data?.error ?? `Request failed with status ${res.status}`);
  }
  return data as T;
}

async function tryRefresh(): Promise<boolean> {
  const refreshToken = await tokenStore.getRefreshToken();
  if (!refreshToken) return false;
  try {
    const res = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!res.ok) return false;
    const data = await res.json();
    await tokenStore.setAccessToken(data.access_token);
    return true;
  } catch {
    return false;
  }
}

export const api = {
  register: (email: string, password: string, fullName: string, inviteCode?: string) =>
    request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, full_name: fullName, invite_code: inviteCode }),
    }),

  login: (email: string, password: string) =>
    request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  getMyHousehold: () => request<Household>('/households/me'),

  upgradeHousehold: () => request<{ tier: string }>('/households/me/upgrade', { method: 'POST' }),

  uploadReceipt: (file: { uri: string; name: string; type: string }) => {
    const form = new FormData();
    // React Native's FormData accepts this {uri,name,type} shape directly.
    form.append('image', file as unknown as Blob);
    return request<ReceiptDraft>('/receipts', { method: 'POST', body: form });
  },

  resolveWarranty: (category: string, brand: string, purchaseDate: string) => {
    const params = new URLSearchParams({ category, brand, purchase_date: purchaseDate });
    return request<WarrantyResolution>(`/warranty-rules/resolve?${params.toString()}`);
  },

  createProduct: (payload: {
    name: string;
    category: string;
    brand?: string;
    purchase_date: string;
    price?: number | null;
    room?: string;
    photo_url?: string;
    receipt_id?: string | null;
    warranty_expires_at?: string;
  }) => request<Product>('/products', { method: 'POST', body: JSON.stringify(payload) }),

  listProducts: (params?: {
    q?: string;
    room?: string;
    category?: string;
    status?: 'ok' | 'warning' | 'expired';
    price_min?: number;
    price_max?: number;
  }) => {
    const search = new URLSearchParams();
    if (params?.q) search.set('q', params.q);
    if (params?.room) search.set('room', params.room);
    if (params?.category) search.set('category', params.category);
    if (params?.status) search.set('status', params.status);
    if (params?.price_min != null) search.set('price_min', String(params.price_min));
    if (params?.price_max != null) search.set('price_max', String(params.price_max));
    const qs = search.toString();
    return request<Product[]>(`/products${qs ? `?${qs}` : ''}`);
  },

  getProduct: (id: string) => request<Product>(`/products/${id}`),

  updateProduct: (id: string, payload: Partial<Product>) =>
    request<Product>(`/products/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),

  createClaim: (productId: string, issueDescription: string) =>
    request<WarrantyClaim>(`/products/${productId}/claims`, {
      method: 'POST',
      body: JSON.stringify({ issue_description: issueDescription }),
    }),

  listClaims: (productId: string) => request<WarrantyClaim[]>(`/products/${productId}/claims`),

  createProductCost: (
    productId: string,
    payload: { amount: number; description?: string; incurred_at?: string },
  ) =>
    request<ProductCost>(`/products/${productId}/costs`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  listProductCosts: (productId: string) => request<ProductCost[]>(`/products/${productId}/costs`),

  reportWarrantyRule: (productId: string, note?: string) =>
    request<{ id: string }>(`/products/${productId}/warranty-report`, {
      method: 'POST',
      body: JSON.stringify({ note }),
    }),

  registerDevice: (expoPushToken: string) =>
    request<{ status: string }>('/devices', {
      method: 'POST',
      body: JSON.stringify({ expo_push_token: expoPushToken }),
    }),

  getReceipt: (id: string) => request<Receipt>(`/receipts/${id}`),

  listManufacturerContacts: () => request<ManufacturerContact[]>('/manufacturer-contacts'),
};
