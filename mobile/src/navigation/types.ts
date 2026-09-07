import type { CompositeScreenProps, NavigatorScreenParams } from '@react-navigation/native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';

import type { ReceiptDraft } from '../api/types';

export type AuthStackParamList = {
  Onboarding: undefined;
  Login: undefined;
  Register: { inviteCode?: string } | undefined;
};

export type AppTabParamList = {
  DashboardTab: undefined;
  SearchTab: undefined;
  AddTab: undefined;
  ExpiringTab: undefined;
  SettingsTab: undefined;
};

export type AppStackParamList = {
  Tabs: NavigatorScreenParams<AppTabParamList> | undefined;
  AddProductChoose: undefined;
  ConfirmProduct: { draft?: ReceiptDraft } | undefined;
  ProductDetail: { productId: string };
  Claim: { productId: string };
  GmailReceipts: undefined;
};

// Screens rendered inside the bottom-tab navigator can navigate both to
// sibling tabs and to screens in the parent stack (e.g. ProductDetail).
export type AppTabScreenProps<T extends keyof AppTabParamList> = CompositeScreenProps<
  BottomTabScreenProps<AppTabParamList, T>,
  NativeStackScreenProps<AppStackParamList>
>;
