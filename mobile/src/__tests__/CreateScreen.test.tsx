import React from 'react';
import {Alert, Animated} from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import {fireEvent, render, waitFor} from '@testing-library/react-native';
import CreateScreen from '../screens/CreateScreen';
import * as api from '../services/api';

jest.mock('../services/api');

describe('CreateScreen', () => {
  const makeNavigation = () => ({
    navigate: jest.fn(),
    goBack: jest.fn(),
    addListener: jest.fn(() => jest.fn()),
  });
  let consoleErrorSpy: jest.SpyInstance;
  let animatedTimingSpy: jest.SpyInstance;

  beforeEach(() => {
    jest.clearAllMocks();
    consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
    animatedTimingSpy = jest.spyOn(Animated, 'timing').mockImplementation(() => ({
      start: (callback?: (result: {finished: boolean}) => void) => callback?.({finished: true}),
      stop: jest.fn(),
      reset: jest.fn(),
    } as any));
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
    (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);
    (AsyncStorage.removeItem as jest.Mock).mockResolvedValue(undefined);
    (api.createExperience as jest.Mock).mockResolvedValue({id: 'exp-1'});
  });

  afterEach(() => {
    animatedTimingSpy.mockRestore();
    consoleErrorSpy.mockRestore();
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('prefills chat note suggestions as private chat-sourced notes', async () => {
    const navigation = makeNavigation();
    const route = {
      params: {
        prefillContent: '先做一小步，再用结果修正判断。',
        defaultVisibility: 'private',
        sourceScene: 'chat',
        sourceMessageIds: ['user-1', 'assistant-1'],
      },
    };

    const {getByDisplayValue, getByText} = render(
      <CreateScreen navigation={navigation} route={route} />,
    );

    expect(await waitFor(() => getByDisplayValue('先做一小步，再用结果修正判断。'))).toBeTruthy();
    expect(getByText('🔒 私密')).toBeTruthy();

    fireEvent.press(getByText('保存'));

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith(
        '要匿名贡献给相似处境的人吗？',
        '不会公开原聊天内容、议题标题或上下文。',
        expect.any(Array),
      );
    });
    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    await buttons.find((button: any) => button.text === '保持私密').onPress();

    await waitFor(() => {
      expect(api.createExperience).toHaveBeenCalledWith(
        '先做一小步，再用结果修正判断。',
        '',
        '',
        true,
        undefined,
        '',
        {
          source_scene: 'chat',
          source_message_ids: ['user-1', 'assistant-1'],
        },
      );
    });
  });

  it('can anonymously contribute a chat-sourced note as public without exposing chat context', async () => {
    const navigation = makeNavigation();
    const route = {
      params: {
        prefillContent: '先做一小步，再用结果修正判断。',
        defaultVisibility: 'private',
        sourceScene: 'chat',
        sourceMessageIds: ['user-1', 'assistant-1'],
      },
    };

    const {getByText} = render(<CreateScreen navigation={navigation} route={route} />);

    fireEvent.press(getByText('保存'));
    await waitFor(() => expect(Alert.alert).toHaveBeenCalled());
    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    await buttons.find((button: any) => button.text === '匿名贡献').onPress();

    await waitFor(() => {
      expect(api.createExperience).toHaveBeenCalledWith(
        '先做一小步，再用结果修正判断。',
        '',
        '',
        false,
        undefined,
        '',
        {
          source_scene: 'chat',
          source_message_ids: ['user-1', 'assistant-1'],
        },
      );
    });
  });

  it('saves normal notes as public note-sourced experiences by default', async () => {
    const navigation = makeNavigation();
    const {getByPlaceholderText, getByText} = render(
      <CreateScreen navigation={navigation} route={{params: {}}} />,
    );

    fireEvent.changeText(getByPlaceholderText('此刻你有什么想说的？'), '今天先把事情做小一点');
    expect(getByText('公开')).toBeTruthy();
    fireEvent.press(getByText('保存'));

    await waitFor(() => {
      expect(api.createExperience).toHaveBeenCalledWith(
        '今天先把事情做小一点',
        '',
        '',
        false,
        undefined,
        '',
        {
          source_scene: 'note',
          source_message_ids: undefined,
        },
      );
    });
  });

  it('asks for a display name and retries the public save when first public note requires it', async () => {
    const navigation = makeNavigation();
    (api.createExperience as jest.Mock)
      .mockRejectedValueOnce({
        status: 400,
        code: 'display_name_required',
        message: '需要先设置展示名',
      })
      .mockResolvedValueOnce({id: 'exp-1'});
    (api.updateProfile as jest.Mock).mockResolvedValue({
      id: 'user-1',
      display_name: '小年糕',
    });

    const {getByPlaceholderText, getByText, queryByText} = render(
      <CreateScreen navigation={navigation} route={{params: {}}} />,
    );

    fireEvent.changeText(getByPlaceholderText('此刻你有什么想说的？'), '今天先把事情做小一点');
    fireEvent.press(getByText('保存'));

    expect(await waitFor(() => getByText('先取个名字'))).toBeTruthy();
    expect(queryByText('发布失败')).toBeNull();

    fireEvent.changeText(getByPlaceholderText('别人会在经验卡上看到这个名字'), ' 小年糕 ');
    fireEvent.press(getByText('保存并继续'));

    await waitFor(() => {
      expect(api.updateProfile).toHaveBeenCalledWith({display_name: '小年糕'});
      expect(api.createExperience).toHaveBeenCalledTimes(2);
      expect(api.createExperience).toHaveBeenLastCalledWith(
        '今天先把事情做小一点',
        '',
        '',
        false,
        undefined,
        '',
        {
          source_scene: 'note',
          source_message_ids: undefined,
        },
      );
    });
  });

  it('keeps the draft editable when display-name setup is cancelled', async () => {
    (api.createExperience as jest.Mock).mockRejectedValueOnce({
      status: 400,
      code: 'display_name_required',
      message: '需要先设置展示名',
    });

    const {getByPlaceholderText, getByText, getByDisplayValue, queryByText} = render(
      <CreateScreen navigation={makeNavigation()} route={{params: {}}} />,
    );

    fireEvent.changeText(getByPlaceholderText('此刻你有什么想说的？'), '今天先把事情做小一点');
    fireEvent.press(getByText('保存'));
    expect(await waitFor(() => getByText('先取个名字'))).toBeTruthy();

    fireEvent.press(getByText('取消'));

    expect(queryByText('先取个名字')).toBeNull();
    expect(getByDisplayValue('今天先把事情做小一点')).toBeTruthy();
    expect(getByText('公开')).toBeTruthy();
  });
});
