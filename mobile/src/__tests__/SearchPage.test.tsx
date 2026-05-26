import React from 'react';
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
    (config.getToken as jest.Mock).mockResolvedValue('token-1');
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  it('shows creator_display_name in search results', async () => {
    (api.searchExperiences as jest.Mock).mockResolvedValue({
      data: [{
        id: 'exp-1',
        author_id: 'author-1',
        content: '人要活出一点自己的劲儿。',
        domain: 'meaning',
        sub_domain: 'self',
        creator_display_name: '姜文',
        creator_name: undefined,
        author_name: undefined,
        source_type: 'platform',
        inspiration_count: 0,
        collection_count: 0,
        is_inspired: false,
        is_collected: false,
        is_official: true,
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
        author_id: 'author-1',
        content: '人要活出一点自己的劲儿。',
        domain: 'meaning',
        sub_domain: 'self',
        creator_display_name: '姜文',
        source_type: 'platform',
        inspiration_count: 0,
        collection_count: 0,
        is_inspired: false,
        is_collected: false,
        is_official: true,
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
});
