import { useRef } from 'react';
import { Animated, Pressable, type PressableProps, type ViewStyle } from 'react-native';

interface Props extends PressableProps {
  style?: ViewStyle | ViewStyle[];
  children: React.ReactNode;
}

// Wraps any pressable content with a quick scale-down on press, matching the
// "instrument panel" tap feedback from the design system (no shadows/ripple,
// just a crisp scale).
export default function PressableScale({ style, children, onPressIn, onPressOut, ...rest }: Props) {
  const scale = useRef(new Animated.Value(1)).current;

  return (
    <Pressable
      onPressIn={(e) => {
        Animated.timing(scale, { toValue: 0.97, duration: 80, useNativeDriver: true }).start();
        onPressIn?.(e);
      }}
      onPressOut={(e) => {
        Animated.timing(scale, { toValue: 1, duration: 80, useNativeDriver: true }).start();
        onPressOut?.(e);
      }}
      {...rest}
    >
      <Animated.View style={[style, { transform: [{ scale }] }]}>{children}</Animated.View>
    </Pressable>
  );
}
