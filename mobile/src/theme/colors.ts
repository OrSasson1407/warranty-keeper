export const colors = {
  background: '#F7F7FA',
  surface: '#FFFFFF',
  text: '#1C1C1E',
  textMuted: '#6B7280',
  border: '#E5E7EB',
  primary: '#2563EB',
  primaryText: '#FFFFFF',
  danger: '#DC2626',

  statusOk: '#16A34A',
  statusOkBg: '#DCFCE7',
  statusWarning: '#EA580C',
  statusWarningBg: '#FFEDD5',
  statusExpired: '#DC2626',
  statusExpiredBg: '#FEE2E2',
};

export type WarrantyStatus = 'ok' | 'warning' | 'expired';

export function statusColor(status: WarrantyStatus) {
  switch (status) {
    case 'ok':
      return { fg: colors.statusOk, bg: colors.statusOkBg };
    case 'warning':
      return { fg: colors.statusWarning, bg: colors.statusWarningBg };
    case 'expired':
      return { fg: colors.statusExpired, bg: colors.statusExpiredBg };
  }
}
