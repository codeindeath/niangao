import {Alert} from 'react-native';
import {handleAuthExpired, requireLogin} from '../utils/authGate';
import * as config from '../services/config';

jest.mock('../services/config');

describe('requireLogin', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
  });

  afterEach(() => {
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('allows the action without an alert when a token exists', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('token-1');
    const navigation = {navigate: jest.fn(), getParent: jest.fn()};

    await expect(requireLogin(navigation, '需要登录')).resolves.toBe(true);

    expect(Alert.alert).not.toHaveBeenCalled();
    expect(navigation.navigate).not.toHaveBeenCalled();
  });

  it('opens root login from nested navigation when the user chooses Apple login', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    const parent = {navigate: jest.fn()};
    const navigation = {
      navigate: jest.fn(),
      getParent: jest.fn(() => parent),
    };

    await expect(requireLogin(navigation, '登录后可以继续')).resolves.toBe(false);
    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    buttons[1].onPress();

    expect(parent.navigate).toHaveBeenCalledWith('login');
    expect(navigation.navigate).not.toHaveBeenCalled();
  });

  it('falls back to local navigation when no parent navigator exists', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    const navigation = {
      navigate: jest.fn(),
      getParent: jest.fn(() => undefined),
    };

    await requireLogin(navigation, '登录后可以继续');
    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    buttons[1].onPress();

    expect(navigation.navigate).toHaveBeenCalledWith('login');
  });

  it('clears expired auth and routes nested screens back to login', async () => {
    const parent = {navigate: jest.fn()};
    const navigation = {
      navigate: jest.fn(),
      getParent: jest.fn(() => parent),
    };

    await expect(handleAuthExpired(navigation, {status: 401})).resolves.toBe(true);

    expect(config.clearToken).toHaveBeenCalledTimes(1);
    expect(Alert.alert).toHaveBeenCalledWith(
      '登录状态过期',
      '重新登录后可以继续。',
      expect.any(Array),
    );

    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    buttons.find((button: any) => button.text === 'Apple登录').onPress();

    expect(parent.navigate).toHaveBeenCalledWith('login');
    expect(navigation.navigate).not.toHaveBeenCalled();
  });

  it('ignores non-auth failures', async () => {
    const navigation = {navigate: jest.fn(), getParent: jest.fn()};

    await expect(handleAuthExpired(navigation, {status: 500})).resolves.toBe(false);

    expect(config.clearToken).not.toHaveBeenCalled();
    expect(Alert.alert).not.toHaveBeenCalled();
    expect(navigation.navigate).not.toHaveBeenCalled();
  });
});
