import React from 'react';
import {act, fireEvent, render, waitFor} from '@testing-library/react-native';
import ProfileEditScreen from '../screens/ProfileEditScreen';
import * as api from '../services/api';

jest.mock('../services/api');

describe('ProfileEditScreen', () => {
  const navigation = {goBack: jest.fn()};
  let consoleErrorSpy: jest.SpyInstance;

  beforeEach(() => {
    jest.clearAllMocks();
    (api.fetchProfile as jest.Mock).mockReset();
    (api.updateProfile as jest.Mock).mockReset();
    consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
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
});
