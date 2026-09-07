import { ActivityIndicator, View } from 'react-native';
import { NavigationContainer, DarkTheme } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';

import { useAuth } from '../context/AuthContext';
import { colors } from '../theme/colors';

import OnboardingScreen from '../screens/OnboardingScreen';
import LoginScreen from '../screens/LoginScreen';
import RegisterScreen from '../screens/RegisterScreen';
import AppTabs from './AppTabs';
import AddProductChooseScreen from '../screens/AddProductChooseScreen';
import ConfirmProductScreen from '../screens/ConfirmProductScreen';
import ProductDetailScreen from '../screens/ProductDetailScreen';
import ClaimScreen from '../screens/ClaimScreen';
import GmailReceiptsScreen from '../screens/GmailReceiptsScreen';

import type { AppStackParamList, AuthStackParamList } from './types';

const AuthStack = createNativeStackNavigator<AuthStackParamList>();
const AppStack = createNativeStackNavigator<AppStackParamList>();

const navTheme = {
  ...DarkTheme,
  colors: {
    ...DarkTheme.colors,
    background: colors.background,
    card: colors.surfaceContainer,
    text: colors.text,
    border: colors.border,
    primary: colors.primary,
  },
};

const stackHeaderOptions = {
  headerStyle: { backgroundColor: colors.surfaceContainer },
  headerTintColor: colors.text,
  headerTitleStyle: { color: colors.text },
  headerShadowVisible: false,
};

function AuthNavigator() {
  return (
    <AuthStack.Navigator screenOptions={{ headerShown: false }}>
      <AuthStack.Screen name="Onboarding" component={OnboardingScreen} />
      <AuthStack.Screen
        name="Login"
        component={LoginScreen}
        options={{ headerShown: true, title: 'התחברות', ...stackHeaderOptions }}
      />
      <AuthStack.Screen
        name="Register"
        component={RegisterScreen}
        options={{ headerShown: true, title: 'הרשמה', ...stackHeaderOptions }}
      />
    </AuthStack.Navigator>
  );
}

function AppNavigator() {
  return (
    <AppStack.Navigator screenOptions={{ headerShown: true, ...stackHeaderOptions }}>
      <AppStack.Screen name="Tabs" component={AppTabs} options={{ headerShown: false }} />
      <AppStack.Screen
        name="AddProductChoose"
        component={AddProductChooseScreen}
        options={{ title: 'הוספת מוצר' }}
      />
      <AppStack.Screen
        name="ConfirmProduct"
        component={ConfirmProductScreen}
        options={{ title: 'אישור פרטים' }}
      />
      <AppStack.Screen
        name="ProductDetail"
        component={ProductDetailScreen}
        options={{ title: 'פרטי מוצר' }}
      />
      <AppStack.Screen name="Claim" component={ClaimScreen} options={{ title: 'תביעת אחריות' }} />
      <AppStack.Screen
        name="GmailReceipts"
        component={GmailReceiptsScreen}
        options={{ title: 'קבלות מ-Gmail' }}
      />
    </AppStack.Navigator>
  );
}

export default function RootNavigator() {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    return (
      <View
        style={{
          flex: 1,
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: colors.background,
        }}
      >
        <ActivityIndicator size="large" />
      </View>
    );
  }

  return (
    <NavigationContainer theme={navTheme}>
      {user ? <AppNavigator /> : <AuthNavigator />}
    </NavigationContainer>
  );
}
