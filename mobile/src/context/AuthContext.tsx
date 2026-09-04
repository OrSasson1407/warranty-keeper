import AsyncStorage from '@react-native-async-storage/async-storage';
import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';

import { api } from '../api/client';
import * as tokenStore from '../api/tokenStore';
import type { User } from '../api/types';

const USER_KEY = 'wk_user';

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  register: (email: string, password: string, fullName: string, inviteCode?: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const [token, storedUser] = await Promise.all([
          tokenStore.getAccessToken(),
          AsyncStorage.getItem(USER_KEY),
        ]);
        if (token && storedUser) {
          setUser(JSON.parse(storedUser));
        }
      } finally {
        setIsLoading(false);
      }
    })();
  }, []);

  const persist = useCallback(async (accessToken: string, refreshToken: string, u: User) => {
    await tokenStore.setTokens(accessToken, refreshToken);
    await AsyncStorage.setItem(USER_KEY, JSON.stringify(u));
    setUser(u);
  }, []);

  const register = useCallback(
    async (email: string, password: string, fullName: string, inviteCode?: string) => {
      const res = await api.register(email, password, fullName, inviteCode);
      await persist(res.access_token, res.refresh_token, res.user);
    },
    [persist]
  );

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await api.login(email, password);
      await persist(res.access_token, res.refresh_token, res.user);
    },
    [persist]
  );

  const logout = useCallback(async () => {
    await tokenStore.clearTokens();
    await AsyncStorage.removeItem(USER_KEY);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, isLoading, register, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
