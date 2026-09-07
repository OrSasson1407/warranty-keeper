// Dark "telemetry" design system — see the Stitch mockups this was built from.
export const colors = {
  background: '#0b1326',
  surfaceDim: '#0b1326',
  surfaceBright: '#31394d',
  surfaceContainerLowest: '#060e20',
  surfaceContainerLow: '#131b2e',
  surfaceContainer: '#171f33',
  surfaceContainerHigh: '#222a3d',
  surfaceContainerHighest: '#2d3449',
  surface: '#171f33',

  text: '#dae2fd',
  textMuted: '#bfc7d2',
  outline: '#89929b',
  border: '#3f4850',

  primary: '#93ccff',
  primaryText: '#003351',
  primaryContainer: '#3198dc',
  onPrimaryContainer: '#002c47',

  secondary: '#b8c4ff',
  onSecondary: '#002584',
  secondaryContainer: '#173bab',
  onSecondaryContainer: '#a0b1ff',

  tertiary: '#4edea3',
  onTertiary: '#003824',
  tertiaryContainer: '#00a572',
  onTertiaryContainer: '#00311f',

  error: '#ffb4ab',
  onError: '#690005',
  errorContainer: '#93000a',
  onErrorContainer: '#ffdad6',

  // back-compat alias used by destructive buttons (logout, delete, claim CTA)
  danger: '#ffb4ab',

  // back-compat aliases for screens not yet migrated to statusColor()/statusAccent()
  statusOk: '#4edea3',
  statusOkBg: '#00a572',
  statusWarning: '#b8c4ff',
  statusWarningBg: '#173bab',
  statusExpired: '#ffb4ab',
  statusExpiredBg: '#93000a',
};

export type WarrantyStatus = 'ok' | 'warning' | 'expired';

// Badge fill (bg) + text/icon (fg) pair per status, matching the mockups exactly.
export function statusColor(status: WarrantyStatus) {
  switch (status) {
    case 'ok':
      return { fg: colors.onTertiaryContainer, bg: colors.tertiaryContainer };
    case 'warning':
      return { fg: colors.onSecondaryContainer, bg: colors.secondaryContainer };
    case 'expired':
      return { fg: colors.onErrorContainer, bg: colors.errorContainer };
  }
}

// Single accent hex per status — used for card "edge beacon" bars and icons,
// as opposed to the badge fill/text pair above.
export function statusAccent(status: WarrantyStatus) {
  switch (status) {
    case 'ok':
      return colors.tertiary;
    case 'warning':
      return colors.secondary;
    case 'expired':
      return colors.error;
  }
}
