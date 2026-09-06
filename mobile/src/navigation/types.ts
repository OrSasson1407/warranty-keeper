import type { ReceiptDraft } from '../api/types';

export type AuthStackParamList = {
  Onboarding: undefined;
  Login: undefined;
  Register: { inviteCode?: string } | undefined;
};

export type AppStackParamList = {
  Dashboard: undefined;
  AddProductChoose: undefined;
  ConfirmProduct: { draft?: ReceiptDraft } | undefined;
  ProductDetail: { productId: string };
  Claim: { productId: string };
  Search: undefined;
  Settings: undefined;
  GmailReceipts: undefined;
};
