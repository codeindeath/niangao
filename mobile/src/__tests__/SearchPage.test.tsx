import React from 'react';
import {Alert} from 'react-native';
import {fireEvent, render, waitFor} from '@testing-library/react-native';
import SearchPage from '../screens/SearchPage';
import * as api from '../services/api';
import * as config from '../services/config';

jest.mock('../services/api');
jest.mock('../services/config');

describe('SearchPage', () => {
  const makeNavigation = () => ({
    navigate: jest.fn(),
    goBack: jest.fn(),
  });
  let consoleErrorSpy: jest.SpyInstance;

  beforeEach(() => {
    jest.clearAllMocks();
    consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
    (config.getToken as jest.Mock).mockResolvedValue('token-1');
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('shows creator_display_name in search results', async () => {
    (api.searchExperiences as jest.Mock).mockResolvedValue({
      data: [{
        id: 'exp-1',
        owner_user_id: 'author-1',
        content: '人要活出一点自己的劲儿。',
        domain: 'meaning',
        sub_domain: 'self',
        creator_display_name: '姜文',
        experience_type: 'platform_selected',
        inspiration_count: 0,
        collection_count: 0,
        is_inspired: false,
        is_collected: false,
        quality_score: 9,
        created_at: '2026-05-26T00:00:00Z',
      }],
    });

    const {getByPlaceholderText, getByText, findByText} = render(
      <SearchPage navigation={makeNavigation()} />,
    );

    fireEvent.changeText(getByPlaceholderText('输入你想找的经验、创作者...'), '姜文');
    fireEvent.press(getByText('搜索'));

    expect(await findByText(/姜文/)).toBeTruthy();
  });

  it('records a search click before opening result cards', async () => {
    const navigation = makeNavigation();
    (api.searchExperiences as jest.Mock).mockResolvedValue({
      data: [{
        id: 'exp-1',
        owner_user_id: 'author-1',
        content: '人要活出一点自己的劲儿。',
        domain: 'meaning',
        sub_domain: 'self',
        creator_display_name: '姜文',
        experience_type: 'platform_selected',
        inspiration_count: 0,
        collection_count: 0,
        is_inspired: false,
        is_collected: false,
        created_at: '2026-05-26T00:00:00Z',
      }],
    });

    const {getByPlaceholderText, getByText, findByText} = render(
      <SearchPage navigation={navigation} />,
    );

    fireEvent.changeText(getByPlaceholderText('输入你想找的经验、创作者...'), ' 姜文 ');
    fireEvent.press(getByText('搜索'));

    fireEvent.press(await findByText('人要活出一点自己的劲儿。'));

    expect(api.recordSearchClick).toHaveBeenCalledWith('exp-1', '姜文', 0);
    expect(navigation.navigate).toHaveBeenCalledWith('searchCard', {
      results: expect.any(Array),
      initialIndex: 0,
      keyword: '姜文',
    });
  });

  it('offers a chat entry when search returns no results', async () => {
    const navigation = makeNavigation();
    (api.searchExperiences as jest.Mock).mockResolvedValue({data: []});

    const {getByPlaceholderText, getByText, findByText} = render(
      <SearchPage navigation={navigation} />,
    );

    fireEvent.changeText(getByPlaceholderText('输入你想找的经验、创作者...'), '我不知道怎么和老板说');
    fireEvent.press(getByText('搜索'));

    expect(await findByText('没找到特别贴近的，换个说法试试')).toBeTruthy();
    fireEvent.press(getByText('去聊聊'));

    await waitFor(() => {
      expect(navigation.navigate).toHaveBeenCalledWith('main', {screen: 'chat'});
    });
  });

  it('clears expired auth when search returns 401', async () => {
    const navigation = makeNavigation();
    (api.searchExperiences as jest.Mock).mockRejectedValueOnce({status: 401});

    const {getByPlaceholderText, getByText, queryByText} = render(
      <SearchPage navigation={navigation} />,
    );

    fireEvent.changeText(getByPlaceholderText('输入你想找的经验、创作者...'), '关系');
    fireEvent.press(getByText('搜索'));

    await waitFor(() => {
      expect(config.clearToken).toHaveBeenCalledTimes(1);
      expect(Alert.alert).toHaveBeenCalledWith(
        '登录状态过期',
        '重新登录后可以继续。',
        expect.any(Array),
      );
    });
    expect(queryByText('搜索失败，请检查网络连接')).toBeNull();

    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    buttons.find((button: any) => button.text === 'Apple登录').onPress();
    expect(navigation.navigate).toHaveBeenCalledWith('login');
  });
});
