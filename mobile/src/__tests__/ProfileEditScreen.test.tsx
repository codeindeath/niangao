import React from 'react';
import {Alert} from 'react-native';
import {act, fireEvent, render, waitFor} from '@testing-library/react-native';
import ProfileEditScreen from '../screens/ProfileEditScreen';
import * as api from '../services/api';
import * as config from '../services/config';

jest.mock('../services/api');
jest.mock('../services/config');

describe('ProfileEditScreen', () => {
  const navigation = {goBack: jest.fn(), navigate: jest.fn(), getParent: jest.fn(() => undefined)};
  let consoleErrorSpy: jest.SpyInstance;

  beforeEach(() => {
    jest.clearAllMocks();
    (api.fetchProfile as jest.Mock).mockReset();
    (api.updateProfile as jest.Mock).mockReset();
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
    consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    (Alert.alert as jest.Mock).mockRestore();
    consoleErrorSpy.mockRestore();
  });

  it('shows a retryable weak state when profile loading fails', async () => {
    (api.fetchProfile as jest.Mock).mockRejectedValue(new Error('network down'));

    const rendered = render(<ProfileEditScreen navigation={navigation} />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });

    await waitFor(() => {
      expect(rendered.getByTestId('profile-edit-load-error').props.children).toBe('个人信息暂时没取到');
    });
    expect(rendered.getByTestId('profile-edit-retry')).toBeTruthy();
    expect(consoleErrorSpy).not.toHaveBeenCalled();
  });

  it('can retry profile loading without leaving the user on an empty form', async () => {
    (api.fetchProfile as jest.Mock)
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce({
        display_name: '阿年',
        free_description: '认真生活',
        common_issues: ['工作', '情绪'],
      });

    const rendered = render(<ProfileEditScreen navigation={navigation} />);

    await waitFor(() => {
      expect(rendered.getByTestId('profile-edit-retry')).toBeTruthy();
    });
    fireEvent.press(rendered.getByTestId('profile-edit-retry'));

    await waitFor(() => {
      expect(api.fetchProfile).toHaveBeenCalledTimes(2);
    });
    expect((await rendered.findByPlaceholderText('输入名字')).props.value).toBe('阿年');
    expect((await rendered.findByPlaceholderText('比如：正在认真生活的人')).props.value).toBe('认真生活');
    expect((await rendered.findByPlaceholderText('用顿号分隔，比如：工作、关系、情绪')).props.value).toBe('工作、情绪');
  });

  it('clears expired auth when saving profile returns 401', async () => {
    (api.fetchProfile as jest.Mock).mockResolvedValue({
      display_name: '阿年',
      free_description: '认真生活',
      common_issues: [],
    });
    (api.updateProfile as jest.Mock).mockRejectedValue({status: 401});

    const rendered = render(<ProfileEditScreen navigation={navigation} />);

    fireEvent.press(await rendered.findByText('保存'));

    await waitFor(() => {
      expect(config.clearToken).toHaveBeenCalledTimes(1);
      expect(Alert.alert).toHaveBeenCalledWith(
        '登录状态过期',
        '重新登录后可以继续。',
        expect.any(Array),
      );
    });

    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    buttons.find((button: any) => button.text === 'Apple登录').onPress();
    expect(navigation.navigate).toHaveBeenCalledWith('login');
  });
});
