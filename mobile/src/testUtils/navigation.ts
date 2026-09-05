// A minimal stand-in for React Navigation's screen props. Cast to `any` at
// the call site — building a real NativeStackScreenProps object is not
// worth the type gymnastics for tests that only care about which methods
// got called and with what arguments.
export function createMockNavigation() {
  return {
    navigate: jest.fn(),
    goBack: jest.fn(),
    replace: jest.fn(),
    setOptions: jest.fn(),
    addListener: jest.fn(() => jest.fn()),
    isFocused: jest.fn(() => true),
  };
}

export function createMockRoute(params?: unknown) {
  return { key: 'test', name: 'Test', params };
}
