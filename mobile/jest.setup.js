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

// Jest's own per-test timeout defaults to 5000ms too — the same value as
// asyncUtilTimeout above. On a slower/colder CI runner that lets the outer
// test timeout race a findBy/waitFor call and win, producing a generic
// "Exceeded timeout of 5000 ms for a test" instead of a useful error. Give
// it real headroom over asyncUtilTimeout so waitFor always gets to finish
// (and report clearly) first.
jest.setTimeout(15000);
