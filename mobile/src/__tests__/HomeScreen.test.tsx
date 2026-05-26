import React from 'react';
import {Alert, Dimensions} from 'react-native';
import { fireEvent, render, waitFor } from '@testing-library/react-native';

jest.mock('../services/api');
jest.mock('../services/config');

const HomeScreen = require('../screens/HomeScreen').default;
const api = require('../services/api');
const config = require('../services/config');

describe('HomeScreen', () => {
  let consoleErrorSpy: jest.SpyInstance;

  const makeExperience = (id: string) => ({
    id,
    author_id: 'author-1',
    content: `第 ${id} 条经验`,
    domain: 'meaning',
    sub_domain: 'self',
    author_name: 'Tester',
    inspiration_count: 0,
    collection_count: 0,
    is_inspired: false,
    is_collected: false,
    is_official: false,
    created_at: '2026-01-01T00:00:00Z',
  });

  beforeEach(() => {
    jest.clearAllMocks();
    consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
    (config.getUserInfo as jest.Mock).mockResolvedValue(null);
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('checks auth token before initial feed load', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('fake-token');
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [{
        id: '1',
        author_id: 'author-1',
        content: '把想法先做成一个能运行的小版本，再判断它值不值得继续。',
        domain: 'meaning',
        sub_domain: 'self',
        author_name: 'Tester',
        inspiration_count: 0,
        collection_count: 0,
        is_inspired: false,
        is_collected: false,
        is_official: false,
        created_at: '2026-01-01T00:00:00Z',
      }],
      total: 1,
    });

    render(<HomeScreen navigation={{}} />);
    await waitFor(() => {
      expect(config.getToken).toHaveBeenCalled();
    });
  });

  it('falls back to public feed when not logged in', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [],
      total: 0,
    });

    render(<HomeScreen navigation={{}} />);
    await waitFor(() => {
      expect(api.fetchRecommendations).toHaveBeenCalledWith(20, 0);
    });
  });

  it('shows empty state when no experiences', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({ data: [], total: 0 });

    const { findByText } = render(<HomeScreen navigation={{}} />);
    expect(await findByText('暂无推荐内容')).toBeTruthy();
  });

  it('shows error state with retry button', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchRecommendations as jest.Mock).mockRejectedValue(new Error('Network error'));

    const { findByText } = render(<HomeScreen navigation={{}} />);
    expect(await findByText('加载失败')).toBeTruthy();
    expect(await findByText('重试')).toBeTruthy();
    expect(consoleErrorSpy).not.toHaveBeenCalled();
  });

  it('uses backend has_more to show recommend end state', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('fake-token');
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: Array.from({length: 20}, (_, index) => makeExperience(String(index + 1))),
      total: 21,
      has_more: false,
    });

    const { findByText } = render(<HomeScreen navigation={{}} />);

    expect(await findByText('这轮先看到这里')).toBeTruthy();
    expect(await findByText('刷新')).toBeTruthy();
  });

  it('keeps existing cards when loading more fails', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('fake-token');
    (api.fetchRecommendations as jest.Mock)
      .mockResolvedValueOnce({
        data: Array.from({length: 20}, (_, index) => makeExperience(String(index + 1))),
        total: 40,
        has_more: true,
      })
      .mockRejectedValueOnce(new Error('Network error'));

    const { findByText, getByTestId } = render(<HomeScreen navigation={{}} />);

    expect(await findByText('第 1 条经验')).toBeTruthy();
    fireEvent(getByTestId('home-feed-list'), 'onEndReached');

    expect(await findByText('网络不稳，点一下重试')).toBeTruthy();
    expect(await findByText('第 1 条经验')).toBeTruthy();
    expect(api.fetchRecommendations).toHaveBeenLastCalledWith(20, 20);
    expect(consoleErrorSpy).not.toHaveBeenCalled();
  });

  it('prompts login instead of loading personal tabs for guests', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [makeExperience('1')],
      total: 1,
      has_more: false,
    });

    const {findByText, getByLabelText} = render(<HomeScreen navigation={{}} />);

    expect(await findByText('第 1 条经验')).toBeTruthy();
    fireEvent.press(getByLabelText('收藏分页'));

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith(
        '先登录一下',
        '登录后可以查看收藏和自己记下的经验。',
        expect.any(Array),
      );
    });
    expect(api.fetchMyBookmarks).not.toHaveBeenCalled();
    expect(api.fetchMyExperiences).not.toHaveBeenCalled();
  });

  it('routes top-bar tap fallback for search and guest personal tabs', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [makeExperience('1')],
      total: 1,
      has_more: false,
    });
    const navigation = require('@react-navigation/native').__mockNavigation;
    navigation.navigate.mockClear();
    const screenWidth = Dimensions.get('window').width;

    const {findByText, getByTestId} = render(<HomeScreen />);
    await findByText('第 1 条经验');

    getByTestId('home-screen').props.onTouchStart({
      nativeEvent: {pageX: screenWidth - 12, pageY: 30},
    });
    getByTestId('home-screen').props.onTouchEnd({
      nativeEvent: {pageX: screenWidth - 12, pageY: 30},
    });
    expect(navigation.navigate).toHaveBeenCalledWith('searchPage');

    getByTestId('home-screen').props.onTouchStart({
      nativeEvent: {pageX: screenWidth * 0.5, pageY: 30},
    });
    getByTestId('home-screen').props.onTouchEnd({
      nativeEvent: {pageX: screenWidth * 0.5, pageY: 30},
    });
    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith(
        '先登录一下',
        '登录后可以查看收藏和自己记下的经验。',
        expect.any(Array),
      );
    });
  });

  it('prompts login before guest card actions reach the backend', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [makeExperience('1')],
      total: 1,
      has_more: false,
    });

    const {findByText, getByLabelText} = render(<HomeScreen navigation={{}} />);

    await findByText('第 1 条经验');
    fireEvent.press(getByLabelText('标记有启发'), {stopPropagation: jest.fn()});
    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith(
        '先登录一下',
        '登录后可以标记有启发，年糕也会更懂你的偏好。',
        expect.any(Array),
      );
    });
    expect(api.markInspired).not.toHaveBeenCalled();

    (Alert.alert as jest.Mock).mockClear();
    fireEvent.press(getByLabelText('收藏经验'), {stopPropagation: jest.fn()});
    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith(
        '先登录一下',
        '登录后可以收藏经验，之后在看看里随时翻回来。',
        expect.any(Array),
      );
    });
    expect(api.setCollected).not.toHaveBeenCalled();
  });

  it('clears expired auth when an authenticated card action returns 401', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('expired-token');
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [makeExperience('1')],
      total: 1,
      has_more: false,
    });
    (api.markInspired as jest.Mock).mockRejectedValue({status: 401});

    const navigation = require('@react-navigation/native').__mockNavigation;
    navigation.navigate.mockClear();
    const {findByText, getByLabelText} = render(<HomeScreen />);

    await findByText('第 1 条经验');
    fireEvent.press(getByLabelText('标记有启发'), {stopPropagation: jest.fn()});

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

  it('records a flip event when a card interpretation is opened', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('fake-token');
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [{...makeExperience('1'), interpretation: '先把它用起来，再判断方向。'}],
      total: 1,
      has_more: false,
    });

    const {findByText} = render(<HomeScreen navigation={{}} />);

    fireEvent.press(await findByText('第 1 条经验'));

    await waitFor(() => {
      expect(api.recordExperienceEvent).toHaveBeenCalledWith('1', 'flip', 'feed', {
        tab: 'recommend',
        side: 'back',
      });
    });
  });

  it('offers turn-private before deleting a V4-owned public card', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('fake-token');
    (config.getUserInfo as jest.Mock).mockResolvedValue({id: 'author-1'});
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [{
        ...makeExperience('1'),
        author_id: 'legacy-author',
        owner_user_id: 'author-1',
        visibility: 'public',
      }],
      total: 1,
      has_more: false,
    });
    (api.updateExperience as jest.Mock).mockResolvedValue({status: 'ok'});

    const {findByText} = render(<HomeScreen navigation={{}} />);

    fireEvent.press(await findByText('删除'), {stopPropagation: jest.fn()});
    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    const makePrivateButton = buttons.find((button: any) => button.text === '转为私密');
    const deleteButton = buttons.find((button: any) => button.text === '删除');
    expect(buttons.indexOf(makePrivateButton)).toBeLessThan(buttons.indexOf(deleteButton));

    makePrivateButton.onPress();

    await waitFor(() => {
      expect(api.updateExperience).toHaveBeenCalledWith(
        '1',
        '第 1 条经验',
        'meaning',
        'self',
        true,
        undefined,
        undefined,
      );
    });
    expect(api.deleteExperience).not.toHaveBeenCalled();
  });
});
