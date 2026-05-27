jest.mock('../services/config', () => ({
  getToken: jest.fn(),
  clearToken: jest.fn(),
  apiFetchWithTimeout: jest.fn(),
}));

jest.mock('@react-navigation/native-stack', () => ({
  createNativeStackNavigator: () => ({
    Navigator: ({children}: any) => children,
    Screen: ({children}: any) => (typeof children === 'function' ? children() : null),
  }),
}));

jest.mock('../navigation/BottomTabNavigator', () => () => null);
jest.mock('../screens/DetailScreen', () => () => null);
jest.mock('../screens/LoginScreen', () => () => null);
jest.mock('../screens/SearchPage', () => () => null);
jest.mock('../screens/SearchCardScreen', () => () => null);
jest.mock('../screens/ProfileEditScreen', () => () => null);
jest.mock('../screens/CreateScreen', () => () => null);

const config = require('../services/config');
const {checkAndValidateToken} = require('../../App');

describe('startup auth validation', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    config.getToken.mockResolvedValue('token-1');
    config.clearToken.mockResolvedValue(undefined);
  });

  it('keeps the local session when server validation has a transient failure', async () => {
    config.apiFetchWithTimeout.mockResolvedValue({status: 500, ok: false});

    await expect(checkAndValidateToken()).resolves.toBe(true);

    expect(config.clearToken).not.toHaveBeenCalled();
  });

  it('clears the local session only when server validation returns unauthorized', async () => {
    config.apiFetchWithTimeout.mockResolvedValue({status: 401, ok: false});

    await expect(checkAndValidateToken()).resolves.toBe(false);

    expect(config.clearToken).toHaveBeenCalledTimes(1);
  });
});
