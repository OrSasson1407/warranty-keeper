import AsyncStorage from '@react-native-async-storage/async-storage';

const ACCESS_KEY = 'wk_access_token';
const REFRESH_KEY = 'wk_refresh_token';

export async function getAccessToken() {
  return AsyncStorage.getItem(ACCESS_KEY);
}

export async function getRefreshToken() {
  return AsyncStorage.getItem(REFRESH_KEY);
}

export async function setTokens(accessToken: string, refreshToken: string) {
  await AsyncStorage.multiSet([
    [ACCESS_KEY, accessToken],
    [REFRESH_KEY, refreshToken],
  ]);
}

export async function setAccessToken(accessToken: string) {
  await AsyncStorage.setItem(ACCESS_KEY, accessToken);
}

export async function clearTokens() {
  await AsyncStorage.multiRemove([ACCESS_KEY, REFRESH_KEY]);
}
