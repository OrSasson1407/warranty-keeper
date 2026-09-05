// Global Jest setup for the whole mobile test suite.
//
// @react-native-async-storage/async-storage has no native module under
// plain Jest (no simulator/device), so any real import of it crashes.
// The package ships an official in-memory mock for exactly this — use it
// globally so screens that pull in tokenStore.ts (directly or via api/client)
// don't each need their own ad-hoc mock.
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
);

// The first test in any given file pays for one-time module/transform
// cold-start, which can push a screen's initial async data load past RTL's
// default 1000ms findBy/waitFor timeout. Raise it globally instead of
// padding every first test by hand.
const { configure } = require('@testing-library/react-native');
configure({ asyncUtilTimeout: 5000 });
