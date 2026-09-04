import { Platform } from 'react-native';

// The Go API's base URL. `localhost` only reaches the dev machine itself —
// on a physical device or emulator you must replace this with your
// computer's LAN IP (e.g. http://192.168.1.20:8080). Web preview and
// iOS simulator can keep localhost.
const DEV_API_HOST = Platform.OS === 'android' ? 'http://10.0.2.2:8080' : 'http://localhost:8080';

export const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL ?? DEV_API_HOST;
