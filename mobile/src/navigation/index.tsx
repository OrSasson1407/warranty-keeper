import { ActivityIndicator, View } from 'react-native';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';

import { useAuth } from '../context/AuthContext';
import { colors } from '../theme/colors';

import OnboardingScreen from '../screens/OnboardingScreen';
import LoginScreen from '../screens/LoginScreen';
import RegisterScreen from '../screens/RegisterScreen';
import DashboardScreen from '../screens/DashboardScreen';
import AddProductChooseScreen from '../screens/AddProductChooseScreen';
import ConfirmProductScreen from '../screens/ConfirmProductScreen';
import ProductDetailScreen from '../screens/ProductDetailScreen';
import ClaimScreen from '../screens/ClaimScreen';
import SearchScreen from '../screens/SearchScreen';
import SettingsScreen from '../screens/SettingsScreen';

import type { AppStackParamList, AuthStackParamList } from './types';

const AuthStack = createNativeStackNavigator<AuthStackParamList>();
const AppStack = createNativeStackNavigator<AppStackParamList>();

function AuthNavigator() {
  return (
    <AuthStack.Navigator screenOptions={{ headerShown: false }}>
      <AuthStack.Screen name="Onboarding" component={OnboardingScreen} />
      <AuthStack.Screen
        name="Login"
        component={LoginScreen}
        options={{ headerShown: true, title: 'התחברות' }}
      />
      <AuthStack.Screen
        name="Register"
        component={RegisterScreen}
        options={{ headerShown: true, title: 'הרשמה' }}
      />
    </AuthStack.Navigator>
  );
}

function AppNavigator() {
  return (
    <AppStack.Navigator screenOptions={{ headerShown: true }}>
      <AppStack.Screen
        name="Dashboard"
        component={DashboardScreen}
        options={{ headerShown: false }}
      />
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
      <AppStack.Screen name="Search" component={SearchScreen} options={{ title: 'חיפוש' }} />
      <AppStack.Screen name="Settings" component={SettingsScreen} options={{ title: 'הגדרות' }} />
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

  return <NavigationContainer>{user ? <AppNavigator /> : <AuthNavigator />}</NavigationContainer>;
}
