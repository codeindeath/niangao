import React from 'react';
import {Alert} from 'react-native';
import {act, fireEvent, render, waitFor} from '@testing-library/react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import ProfileScreen from '../screens/ProfileScreen';
import * as api from '../services/api';
import {logout} from '../services/auth';

jest.mock('../services/api');
jest.mock('../services/auth', () => ({
  logout: jest.fn(),
}));

function collectText(node: any, out: string[] = []): string[] {
  if (!node) return out;
  if (typeof node === 'string') {
    out.push(node);
    return out;
  }
  if (Array.isArray(node)) {
    node.forEach(child => collectText(child, out));
    return out;
  }
  collectText(node.children, out);
  return out;
}

function collectAccessibilityLabels(node: any, out: string[] = []): string[] {
  if (!node || typeof node === 'string') return out;
  if (Array.isArray(node)) {
    node.forEach(child => collectAccessibilityLabels(child, out));
    return out;
  }
  if (node.props?.accessibilityLabel) {
    out.push(node.props.accessibilityLabel);
  }
  collectAccessibilityLabels(node.children, out);
  return out;
}

async function flushProfileLoad(): Promise<void> {
  await act(async () => {
    await Promise.resolve();
  });
  await act(async () => {
    await Promise.resolve();
  });
}

describe('ProfileScreen', () => {
  const navigation = {
    navigate: jest.fn(),
    addListener: jest.fn(() => jest.fn()),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
    (api.fetchProfile as jest.Mock).mockResolvedValue({
      id: 'user-1',
      display_name: '阿年',
      free_description: '生活有态度',
    });
    (api.fetchAssetStats as jest.Mock).mockResolvedValue({
      my_experiences: 3,
      collections: 5,
      month_added: 2,
      public_experiences: 2,
      private_experiences: 1,
      from_note: 2,
      from_chat: 1,
    });
    (api.fetchContributionStats as jest.Mock).mockResolvedValue({
      inspired_users: 8,
      collected_count: 4,
      month_inspired_users: 3,
      month_collected: 2,
    });
    (api.fetchChangeStats as jest.Mock).mockResolvedValue({
      chat_topics: 2,
      clearer_count: 1,
      month_chat_experiences: 1,
    });
    (api.fetchRecentHarvestStats as jest.Mock).mockResolvedValue({
      range: '30d',
      note_added: 2,
      chat_experiences: 1,
      inspired_users: 3,
      collected_count: 2,
    });
    (api.fetchRecentRespondedExperiences as jest.Mock).mockResolvedValue({
      data: [],
    });
    (AsyncStorage.getItem as jest.Mock).mockResolvedValue(null);
    (AsyncStorage.setItem as jest.Mock).mockResolvedValue(undefined);
    (AsyncStorage.removeItem as jest.Mock).mockResolvedValue(undefined);
  });

  afterEach(() => {
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('keeps the profile visible when a stats request fails', async () => {
    (api.fetchAssetStats as jest.Mock).mockRejectedValueOnce(new Error('stats down'));

    render(<ProfileScreen navigation={navigation} />);

    await waitFor(() => {
      expect(api.fetchProfile).toHaveBeenCalled();
    });
    expect(api.fetchAssetStats).toHaveBeenCalled();
    expect(api.fetchContributionStats).toHaveBeenCalled();
    expect(api.fetchChangeStats).toHaveBeenCalled();
    expect(api.fetchRecentHarvestStats).toHaveBeenCalled();
    expect(api.fetchRecentRespondedExperiences).toHaveBeenCalled();
  });

  it('shows placeholders and retry when stats are unavailable', async () => {
    (api.fetchAssetStats as jest.Mock).mockRejectedValueOnce(new Error('assets down'));
    (api.fetchContributionStats as jest.Mock).mockRejectedValueOnce(new Error('contribution down'));
    (api.fetchChangeStats as jest.Mock).mockRejectedValueOnce(new Error('change down'));
    (api.fetchRecentHarvestStats as jest.Mock).mockRejectedValueOnce(new Error('recent down'));
    (api.fetchRecentRespondedExperiences as jest.Mock).mockRejectedValueOnce(new Error('responded down'));

    const {findByText, findAllByText} = render(<ProfileScreen navigation={navigation} />);

    expect(await findByText('部分信息暂时没取到')).toBeTruthy();
    expect(await findByText('重试')).toBeTruthy();
    expect((await findAllByText('—')).length).toBeGreaterThan(0);
  });

  it('uses cached stats instead of zeroes when live stats fail', async () => {
    (AsyncStorage.getItem as jest.Mock).mockResolvedValue(JSON.stringify({
      assets: {
        my_experiences: 7,
        collections: 4,
        month_added: 1,
        public_experiences: 5,
        private_experiences: 2,
        from_note: 6,
        from_chat: 1,
      },
      contribution: {
        inspired_users: 11,
        collected_count: 3,
        month_inspired_users: 2,
        month_collected: 1,
      },
      change: {
        chat_topics: 9,
        clearer_count: 0,
        month_chat_experiences: 0,
      },
      recentHarvestByRange: {
        '30d': {
          range: '30d',
          note_added: 6,
          chat_experiences: 1,
          inspired_users: 2,
          collected_count: 1,
        },
      },
      respondedCards: [],
    }));
    (api.fetchAssetStats as jest.Mock).mockRejectedValueOnce(new Error('assets down'));
    (api.fetchContributionStats as jest.Mock).mockRejectedValueOnce(new Error('contribution down'));
    (api.fetchChangeStats as jest.Mock).mockRejectedValueOnce(new Error('change down'));
    (api.fetchRecentHarvestStats as jest.Mock).mockRejectedValueOnce(new Error('recent down'));
    (api.fetchRecentRespondedExperiences as jest.Mock).mockRejectedValueOnce(new Error('responded down'));

    const {findByText, queryByText} = render(<ProfileScreen navigation={navigation} />);

    expect(await findByText('部分信息暂时没取到')).toBeTruthy();
    expect(await findByText('7')).toBeTruthy();
    expect(await findByText('4 条收藏')).toBeTruthy();
    expect(queryByText('— 条收藏')).toBeNull();
  });

  it('caches successfully loaded stats for later weak-network states', async () => {
    render(<ProfileScreen navigation={navigation} />);

    await waitFor(() => {
      expect(AsyncStorage.setItem).toHaveBeenCalledWith(
        'niangao:v4:me-stats-cache',
        expect.stringContaining('"my_experiences":3'),
      );
    });
  });

  it('shows the login gate when profile is unauthorized', async () => {
    (api.fetchProfile as jest.Mock).mockRejectedValueOnce({status: 401});

    const {findByText} = render(<ProfileScreen navigation={navigation} />);

    expect(await findByText('Apple登录')).toBeTruthy();
    await waitFor(() => {
      expect(logout).toHaveBeenCalledTimes(1);
      expect(api.fetchAssetStats).not.toHaveBeenCalled();
    });
  });

  it('orders responded experiences before long-term accumulation', async () => {
    const rendered = render(<ProfileScreen navigation={navigation} />);

    await rendered.findByText('最近有回应的经验');
    const text = collectText(rendered.toJSON()).join('|');

    expect(text.indexOf('最近收获')).toBeLessThan(text.indexOf('最近有回应的经验'));
    expect(text.indexOf('最近有回应的经验')).toBeLessThan(text.indexOf('我的积累'));
  });

  it('uses first-phase bottom actions without future settings placeholders', async () => {
    const rendered = render(<ProfileScreen navigation={navigation} />);
    await flushProfileLoad();
    await waitFor(() => {
      expect(collectText(rendered.toJSON()).join('')).toContain('我的积累');
    });

    const labels = collectAccessibilityLabels(rendered.toJSON());
    expect(labels).toEqual(expect.arrayContaining(['意见反馈', '退出登录', '注销账号']));
    const text = collectText(rendered.toJSON()).join('');
    expect(text).toContain('我的积累');
    expect(text).toContain('意见反馈');
    expect(text).toContain('注销账号');
    expect(text).not.toContain('对话风格');
    expect(text).not.toContain('设置');
    expect(text).not.toContain('□');
    expect(text).not.toContain('↪');
    expect(text).not.toContain('×');
    expect(text).not.toContain('›');
  });

  it('confirms before logging out', async () => {
    const {findByText} = render(<ProfileScreen navigation={navigation} />);

    fireEvent.press(await findByText('退出登录'));
    expect(logout).not.toHaveBeenCalled();

    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    await act(async () => {
      await buttons[1].onPress();
    });
    expect(logout).toHaveBeenCalled();
  });

  it('routes expired auth during feedback submission back to login without a generic failure alert', async () => {
    (api.submitFeedback as jest.Mock).mockRejectedValueOnce({status: 401});

    const rendered = render(<ProfileScreen navigation={navigation} />);
    fireEvent.press(await rendered.findByText('意见反馈'));
    fireEvent.changeText(rendered.getByPlaceholderText('写下你的反馈'), '这里提交起来不顺');

    await act(async () => {
      fireEvent.press(rendered.getByText('提交'));
    });

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith(
        '登录状态过期',
        expect.any(String),
        expect.any(Array),
      );
    });
    expect(Alert.alert).not.toHaveBeenCalledWith('提交失败', expect.any(String));

    const expiredCall = (Alert.alert as jest.Mock).mock.calls.find(
      call => call[0] === '登录状态过期',
    );
    const buttons = expiredCall[2];
    buttons.find((button: any) => button.text === 'Apple登录').onPress();

    expect(navigation.navigate).toHaveBeenCalledWith('login');
  });

  it('sanitizes technical feedback submission failures before showing the alert', async () => {
    (api.submitFeedback as jest.Mock).mockRejectedValueOnce(new Error('network down'));

    const rendered = render(<ProfileScreen navigation={navigation} />);
    fireEvent.press(await rendered.findByText('意见反馈'));
    fireEvent.changeText(rendered.getByPlaceholderText('写下你的反馈'), '这里提交起来不顺');

    await act(async () => {
      fireEvent.press(rendered.getByText('提交'));
    });

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith('提交失败', '网络不稳，请稍后再试');
    });
  });

  it('routes expired auth during account deletion back to login without a generic failure alert', async () => {
    (api.deleteAccount as jest.Mock).mockRejectedValueOnce({status: 401});

    const {findByText} = render(<ProfileScreen navigation={navigation} />);
    fireEvent.press(await findByText('注销账号'));

    let buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    await act(async () => {
      await buttons.find((button: any) => button.text === '继续').onPress();
    });
    buttons = (Alert.alert as jest.Mock).mock.calls[1][2];
    await act(async () => {
      await buttons.find((button: any) => button.text === '确认注销').onPress();
    });

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith(
        '登录状态过期',
        expect.any(String),
        expect.any(Array),
      );
    });
    expect(Alert.alert).not.toHaveBeenCalledWith('注销失败', expect.any(String));
  });

  it('sanitizes technical account deletion failures before showing the alert', async () => {
    (api.deleteAccount as jest.Mock).mockRejectedValueOnce(new Error('network down'));

    const {findByText} = render(<ProfileScreen navigation={navigation} />);
    fireEvent.press(await findByText('注销账号'));

    let buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    await act(async () => {
      await buttons.find((button: any) => button.text === '继续').onPress();
    });
    buttons = (Alert.alert as jest.Mock).mock.calls[1][2];
    await act(async () => {
      await buttons.find((button: any) => button.text === '确认注销').onPress();
    });

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith('注销失败', '网络不稳，请稍后再试');
    });
  });
});
