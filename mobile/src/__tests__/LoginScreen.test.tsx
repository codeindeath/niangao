import React from 'react';
import {Alert} from 'react-native';
import {fireEvent, render, waitFor} from '@testing-library/react-native';
import * as AppleAuthentication from 'expo-apple-authentication';
import LoginScreen from '../screens/LoginScreen';
import * as api from '../services/api';
import * as config from '../services/config';

jest.mock('../services/api');
jest.mock('../services/config');
jest.mock('expo-apple-authentication', () => ({
  AppleAuthenticationScope: {
    FULL_NAME: 'FULL_NAME',
    EMAIL: 'EMAIL',
  },
  signInAsync: jest.fn(),
}));

describe('LoginScreen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
  });

  afterEach(() => {
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('enters guest browsing without touching auth APIs', () => {
    const onSkip = jest.fn();
    const {getByText} = render(<LoginScreen onSkip={onSkip} />);

    expect(getByText('年糕')).toBeTruthy();
    expect(getByText('生活有态度')).toBeTruthy();

    fireEvent.press(getByText('先看看'));

    expect(onSkip).toHaveBeenCalledTimes(1);
    expect(api.appleLogin).not.toHaveBeenCalled();
    expect(api.devLogin).not.toHaveBeenCalled();
    expect(config.setToken).not.toHaveBeenCalled();
  });

  it('stores the dev-login session before entering the app', async () => {
    const onLoginSuccess = jest.fn();
    (api.devLogin as jest.Mock).mockResolvedValue({
      token: 'token-1',
      refresh_token: 'refresh-1',
      user: {id: 'user-1', display_name: '阿年'},
    });

    const {getByText} = render(<LoginScreen onLoginSuccess={onLoginSuccess} />);
    fireEvent.press(getByText('开发模拟登录'));

    await waitFor(() => {
      expect(api.devLogin).toHaveBeenCalledTimes(1);
      expect(config.setToken).toHaveBeenCalledWith('token-1');
      expect(config.setRefreshToken).toHaveBeenCalledWith('refresh-1');
      expect(config.setUserInfo).toHaveBeenCalledWith({id: 'user-1', display_name: '阿年'});
      expect(onLoginSuccess).toHaveBeenCalledTimes(1);
    });
  });

  it('shows a useful message when dev login is unavailable on the current backend', async () => {
    (api.devLogin as jest.Mock).mockRejectedValue({
      status: 404,
      message: '404 page not found',
    });

    const {getByText} = render(<LoginScreen />);
    fireEvent.press(getByText('开发模拟登录'));

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith(
        '开发登录不可用',
        '当前后端还没有启用开发登录接口，请切换到 V4 测试后端再试。',
      );
    });
    expect(config.setToken).not.toHaveBeenCalled();
  });

  it('silently stays on the login page when Apple login is cancelled', async () => {
    const onLoginSuccess = jest.fn();
    (AppleAuthentication.signInAsync as jest.Mock).mockRejectedValue({code: 'ERR_CANCELED'});

    const {getByText} = render(<LoginScreen onLoginSuccess={onLoginSuccess} />);
    fireEvent.press(getByText('Apple登录'));

    await waitFor(() => {
      expect(AppleAuthentication.signInAsync).toHaveBeenCalledTimes(1);
    });
    expect(Alert.alert).not.toHaveBeenCalled();
    expect(api.appleLogin).not.toHaveBeenCalled();
    expect(onLoginSuccess).not.toHaveBeenCalled();
  });

  it('starts Apple login only once while the native prompt is in flight', () => {
    (AppleAuthentication.signInAsync as jest.Mock).mockReturnValue(new Promise(() => {}));

    const {getByText} = render(<LoginScreen />);
    fireEvent.press(getByText('Apple登录'));
    fireEvent.press(getByText('Apple登录'));

    expect(AppleAuthentication.signInAsync).toHaveBeenCalledTimes(1);
    expect(api.appleLogin).not.toHaveBeenCalled();
  });

  it('rejects Apple login locally when identity token is missing', async () => {
    (AppleAuthentication.signInAsync as jest.Mock).mockResolvedValue({
      identityToken: null,
      fullName: {givenName: '年', familyName: '糕'},
    });

    const {getByText} = render(<LoginScreen />);
    fireEvent.press(getByText('Apple登录'));

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith('登录失败', 'Apple登录凭证无效，请重试');
    });
    expect(api.appleLogin).not.toHaveBeenCalled();
    expect(config.setToken).not.toHaveBeenCalled();
  });
});
