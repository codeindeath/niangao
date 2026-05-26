// Jest setup: mock native modules not available in test environment
jest.mock('@react-native-async-storage/async-storage', () => ({
  setItem: jest.fn(),
  getItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
}));

jest.mock('react-native-safe-area-context', () => {
  const React = require('react');
  const {View} = require('react-native');

  return {
    SafeAreaProvider: ({children}) => React.createElement(View, null, children),
    SafeAreaView: ({children, ...props}) => React.createElement(View, props, children),
    useSafeAreaInsets: () => ({top: 0, right: 0, bottom: 0, left: 0}),
  };
});

const mockNavigation = {
  navigate: jest.fn(),
  goBack: jest.fn(),
  addListener: jest.fn(() => jest.fn()),
};

jest.mock('@react-navigation/native', () => ({
  __mockNavigation: mockNavigation,
  useFocusEffect: jest.fn((callback) => callback()),
  useNavigation: jest.fn(() => mockNavigation),
}));
