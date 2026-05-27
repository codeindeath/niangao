import React from 'react';
import {Alert} from 'react-native';
import {act, fireEvent, render, waitFor} from '@testing-library/react-native';
import SearchCardScreen from '../screens/SearchCardScreen';
import * as api from '../services/api';
import * as config from '../services/config';

jest.mock('../services/api');
jest.mock('../services/config');

describe('SearchCardScreen', () => {
  const experience = {
    id: '1',
    owner_user_id: 'author-1',
    content: '第 1 条经验',
    domain: 'meaning',
    sub_domain: 'self',
    creator_display_name: 'Tester',
    inspiration_count: 0,
    collection_count: 0,
    is_inspired: false,
    is_collected: false,
    visibility: 'public',
    created_at: '2026-01-01T00:00:00Z',
  };

  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
    (config.getUserInfo as jest.Mock).mockResolvedValue({id: 'author-1'});
    (config.getToken as jest.Mock).mockResolvedValue('token-1');
  });

  afterEach(() => {
    act(() => {
      jest.runOnlyPendingTimers();
    });
    jest.useRealTimers();
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('offers turn-private before deleting a V4-owned public search card', async () => {
    (api.updateExperience as jest.Mock).mockResolvedValue({status: 'ok'});

    const {findByText} = render(
      <SearchCardScreen
        route={{params: {results: [experience], initialIndex: 0, keyword: '测试'}}}
        navigation={{goBack: jest.fn(), navigate: jest.fn()}}
      />,
    );

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

  it('shows action failure feedback when inspiring fails for a non-auth error', async () => {
    (api.markInspired as jest.Mock).mockRejectedValueOnce(new Error('network down'));

    const {findByLabelText} = render(
      <SearchCardScreen
        route={{params: {results: [experience], initialIndex: 0, keyword: '测试'}}}
        navigation={{goBack: jest.fn(), navigate: jest.fn()}}
      />,
    );

    fireEvent.press(await findByLabelText('标记有启发'), {stopPropagation: jest.fn()});

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith('操作失败', '网络不稳，请稍后再试');
    });
  });

  it('marks inspiration only once while the search-card action is in flight', async () => {
    (api.markInspired as jest.Mock).mockReturnValue(new Promise(() => {}));

    const {findByLabelText} = render(
      <SearchCardScreen
        route={{params: {results: [experience], initialIndex: 0, keyword: '测试'}}}
        navigation={{goBack: jest.fn(), navigate: jest.fn()}}
      />,
    );

    const inspireButton = await findByLabelText('标记有启发');
    fireEvent.press(inspireButton, {stopPropagation: jest.fn()});
    fireEvent.press(inspireButton, {stopPropagation: jest.fn()});

    await waitFor(() => {
      expect(api.markInspired).toHaveBeenCalledTimes(1);
    });
  });

  it('updates collection only once while the search-card action is in flight', async () => {
    (api.setCollected as jest.Mock).mockReturnValue(new Promise(() => {}));

    const {findByLabelText} = render(
      <SearchCardScreen
        route={{params: {results: [experience], initialIndex: 0, keyword: '测试'}}}
        navigation={{goBack: jest.fn(), navigate: jest.fn()}}
      />,
    );

    const collectButton = await findByLabelText('收藏经验');
    fireEvent.press(collectButton, {stopPropagation: jest.fn()});
    fireEvent.press(collectButton, {stopPropagation: jest.fn()});

    await waitFor(() => {
      expect(api.setCollected).toHaveBeenCalledTimes(1);
      expect(api.setCollected).toHaveBeenCalledWith('1', true);
    });
  });
});
