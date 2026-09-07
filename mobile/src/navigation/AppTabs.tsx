import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';

import DashboardScreen from '../screens/DashboardScreen';
import SearchScreen from '../screens/SearchScreen';
import ExpiringScreen from '../screens/ExpiringScreen';
import SettingsScreen from '../screens/SettingsScreen';
import { colors } from '../theme/colors';
import { typography } from '../theme/typography';
import type { AppTabParamList } from './types';

const Tab = createBottomTabNavigator<AppTabParamList>();

const ICONS: Record<keyof AppTabParamList, string> = {
  DashboardTab: '▦',
  SearchTab: '⌕',
  AddTab: '+',
  ExpiringTab: '◔',
  SettingsTab: '⚙',
};

const LABELS: Record<keyof AppTabParamList, string> = {
  DashboardTab: 'ראשי',
  SearchTab: 'חיפוש',
  AddTab: '',
  ExpiringTab: 'תפוגה',
  SettingsTab: 'הגדרות',
};

// Placeholder screen for the center "add" tab -- tabPress is intercepted
// (see AddTab's listeners below) before this would ever render.
function AddPlaceholder() {
  return <View />;
}

export default function AppTabs() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarActiveTintColor: colors.primary,
        tabBarInactiveTintColor: colors.textMuted,
        tabBarStyle: styles.tabBar,
        tabBarLabelStyle: styles.tabLabel,
        tabBarIcon: ({ color }) =>
          route.name === 'AddTab' ? null : (
            <Text style={[styles.icon, { color }]}>{ICONS[route.name]}</Text>
          ),
        tabBarLabel: LABELS[route.name] || undefined,
      })}
    >
      <Tab.Screen name="DashboardTab" component={DashboardScreen} />
      <Tab.Screen name="SearchTab" component={SearchScreen} />
      <Tab.Screen
        name="AddTab"
        component={AddPlaceholder}
        options={{
          tabBarButton: (props) => (
            <View style={styles.addButtonWrap} pointerEvents="box-none">
              <TouchableOpacity
                style={styles.addButton}
                activeOpacity={0.85}
                onPress={(e) => props.onPress?.(e)}
              >
                <Text style={styles.addButtonText}>+</Text>
              </TouchableOpacity>
            </View>
          ),
        }}
        listeners={({ navigation }) => ({
          tabPress: (e) => {
            e.preventDefault();
            navigation.getParent()?.navigate('AddProductChoose');
          },
        })}
      />
      <Tab.Screen name="ExpiringTab" component={ExpiringScreen} />
      <Tab.Screen name="SettingsTab" component={SettingsScreen} />
    </Tab.Navigator>
  );
}

const styles = StyleSheet.create({
  tabBar: {
    backgroundColor: colors.surfaceContainerLow,
    borderTopWidth: 1,
    borderTopColor: colors.border,
    height: 64,
  },
  tabLabel: { ...typography.labelSm, textTransform: 'none' },
  icon: { fontSize: 20 },
  addButtonWrap: { flex: 1, alignItems: 'center', justifyContent: 'center', top: -12 },
  addButton: {
    width: 48,
    height: 48,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  addButtonText: { color: colors.primaryText, fontSize: 26, lineHeight: 28 },
});
