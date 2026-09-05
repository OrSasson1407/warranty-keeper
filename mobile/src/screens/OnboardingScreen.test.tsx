import { Dimensions, ScrollView } from 'react-native';
import { fireEvent, render, screen } from '@testing-library/react-native';

import OnboardingScreen from './OnboardingScreen';
import { createMockNavigation } from '../testUtils/navigation';

const { width } = Dimensions.get('window');

describe('OnboardingScreen', () => {
  it('shows the first slide and a "next" button initially', () => {
    const navigation = createMockNavigation();
    render(<OnboardingScreen navigation={navigation as any} route={{} as any} />);

    expect(screen.getByText('כל האחריות שלך, במקום אחד')).toBeTruthy();
    expect(screen.getByText('הבא')).toBeTruthy();
  });

  it('navigates to Login when "יש לי כבר חשבון" is pressed', () => {
    const navigation = createMockNavigation();
    render(<OnboardingScreen navigation={navigation as any} route={{} as any} />);

    fireEvent.press(screen.getByText('יש לי כבר חשבון'));
    expect(navigation.navigate).toHaveBeenCalledWith('Login');
  });

  it('advances slides on scroll and switches the button to "הרשמה" on the last one', () => {
    const navigation = createMockNavigation();
    render(<OnboardingScreen navigation={navigation as any} route={{} as any} />);
    const scrollView = screen.UNSAFE_getByType(ScrollView);

    fireEvent.scroll(scrollView, { nativeEvent: { contentOffset: { x: width } } });
    expect(screen.getByText('צלם קבלה — וזהו')).toBeTruthy();

    fireEvent.scroll(scrollView, { nativeEvent: { contentOffset: { x: width * 2 } } });
    expect(screen.getByText('תזכורת לפני שהזמן אוזל')).toBeTruthy();
    expect(screen.getByText('הרשמה')).toBeTruthy();
  });

  it("navigates to Register when the last slide's primary button is pressed", () => {
    const navigation = createMockNavigation();
    render(<OnboardingScreen navigation={navigation as any} route={{} as any} />);
    const scrollView = screen.UNSAFE_getByType(ScrollView);

    fireEvent.scroll(scrollView, { nativeEvent: { contentOffset: { x: width * 2 } } });
    fireEvent.press(screen.getByText('הרשמה'));

    expect(navigation.navigate).toHaveBeenCalledWith('Register');
  });
});
