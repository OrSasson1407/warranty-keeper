import { useRef, useState } from 'react';
import {
  Dimensions,
  NativeScrollEvent,
  NativeSyntheticEvent,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { colors } from '../theme/colors';
import { fonts, typography } from '../theme/typography';
import type { AuthStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<AuthStackParamList, 'Onboarding'>;

const SLIDES = [
  {
    emoji: '🧾',
    title: 'כל האחריות שלך, במקום אחד',
    body: 'אף פעם לא תפספס תיקון בחינם כי לא זכרת שהמוצר עדיין באחריות.',
  },
  {
    emoji: '📷',
    title: 'צלם קבלה — וזהו',
    body: 'האפליקציה מזהה את המוצר, התאריך והמחיר, ומחשבת מתי פגה האחריות.',
  },
  {
    emoji: '🔔',
    title: 'תזכורת לפני שהזמן אוזל',
    body: 'נתריע 30 יום לפני שהאחריות פגה, כדי שתספיקו לפעול.',
  },
];

const { width } = Dimensions.get('window');

export default function OnboardingScreen({ navigation }: Props) {
  const [index, setIndex] = useState(0);
  const scrollRef = useRef<ScrollView>(null);

  const onScroll = (e: NativeSyntheticEvent<NativeScrollEvent>) => {
    const i = Math.round(e.nativeEvent.contentOffset.x / width);
    setIndex(i);
  };

  const isLast = index === SLIDES.length - 1;

  return (
    <View style={styles.container}>
      <ScrollView
        ref={scrollRef}
        horizontal
        pagingEnabled
        showsHorizontalScrollIndicator={false}
        onScroll={onScroll}
        scrollEventThrottle={16}
        style={styles.scroll}
      >
        {SLIDES.map((slide, i) => (
          <View key={i} style={[styles.slide, { width }]}>
            <Text style={styles.emoji}>{slide.emoji}</Text>
            <Text style={styles.title}>{slide.title}</Text>
            <Text style={styles.body}>{slide.body}</Text>
          </View>
        ))}
      </ScrollView>

      <View style={styles.dots}>
        {SLIDES.map((_, i) => (
          <View key={i} style={[styles.dot, i === index && styles.dotActive]} />
        ))}
      </View>

      <View style={styles.footer}>
        <TouchableOpacity
          style={styles.primaryButton}
          onPress={() => {
            if (isLast) {
              navigation.navigate('Register');
            } else {
              scrollRef.current?.scrollTo({ x: width * (index + 1), animated: true });
            }
          }}
        >
          <Text style={styles.primaryButtonText}>{isLast ? 'הרשמה' : 'הבא'}</Text>
        </TouchableOpacity>
        <TouchableOpacity onPress={() => navigation.navigate('Login')}>
          <Text style={styles.secondaryText}>יש לי כבר חשבון</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  scroll: { flexGrow: 0 },
  slide: { alignItems: 'center', justifyContent: 'center', padding: 32 },
  emoji: { fontSize: 64, marginBottom: 24 },
  title: {
    ...typography.headlineLg,
    color: colors.text,
    textAlign: 'center',
    marginBottom: 12,
  },
  body: { ...typography.bodyLg, color: colors.textMuted, textAlign: 'center' },
  dots: { flexDirection: 'row', justifyContent: 'center', gap: 8, marginTop: 8 },
  dot: { width: 8, height: 2, backgroundColor: colors.border },
  dotActive: { backgroundColor: colors.primary },
  footer: { padding: 24, gap: 16, alignItems: 'center' },
  primaryButton: {
    backgroundColor: colors.primary,
    paddingVertical: 16,
    width: '100%',
    alignItems: 'center',
  },
  primaryButtonText: { color: colors.primaryText, ...typography.headlineSm, fontFamily: fonts.headlineSm },
  secondaryText: { ...typography.bodyMd, color: colors.textMuted },
});
