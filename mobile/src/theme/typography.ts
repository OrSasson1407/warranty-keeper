import type { TextStyle } from 'react-native';

// Matches the Stitch design system: Space Grotesk for numeric/status/heading
// telemetry, Work Sans for Hebrew body copy. Load both via useFonts() in
// App.tsx before rendering — these family names must match exactly.
export const fonts = {
  displayLg: 'SpaceGrotesk_700Bold',
  headlineLg: 'SpaceGrotesk_700Bold',
  headlineSm: 'SpaceGrotesk_600SemiBold',
  labelNumeric: 'SpaceGrotesk_700Bold',
  labelSm: 'SpaceGrotesk_600SemiBold',
  bodyLg: 'WorkSans_400Regular',
  bodyMd: 'WorkSans_400Regular',
  bodySm: 'WorkSans_400Regular',
  bodyMdMedium: 'WorkSans_500Medium',
  bodyMdSemiBold: 'WorkSans_600SemiBold',
};

export const typography: Record<string, TextStyle> = {
  displayLg: {
    fontFamily: fonts.displayLg,
    fontSize: 32,
    lineHeight: 40,
    letterSpacing: -0.64,
  },
  headlineLg: {
    fontFamily: fonts.headlineLg,
    fontSize: 24,
    lineHeight: 32,
    letterSpacing: -0.24,
  },
  headlineSm: {
    fontFamily: fonts.headlineSm,
    fontSize: 18,
    lineHeight: 24,
  },
  bodyLg: {
    fontFamily: fonts.bodyLg,
    fontSize: 16,
    lineHeight: 24,
  },
  bodyMd: {
    fontFamily: fonts.bodyMd,
    fontSize: 14,
    lineHeight: 20,
  },
  bodySm: {
    fontFamily: fonts.bodySm,
    fontSize: 12,
    lineHeight: 16,
    letterSpacing: 0.12,
  },
  labelNumeric: {
    fontFamily: fonts.labelNumeric,
    fontSize: 20,
    lineHeight: 24,
    letterSpacing: 1,
  },
  labelSm: {
    fontFamily: fonts.labelSm,
    fontSize: 11,
    lineHeight: 14,
    letterSpacing: 0.66,
    textTransform: 'uppercase',
  },
};
