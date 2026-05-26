import React from 'react';
import {fireEvent, render, waitFor} from '@testing-library/react-native';

jest.mock('../services/api');
jest.mock('../services/config', () => ({
  clearToken: jest.fn(),
}));

const ChatScreen = require('../screens/ChatScreen').default;
const api = require('../services/api');
const config = require('../services/config');

describe('ChatScreen', () => {
  const makeNavigation = () => ({
    navigate: jest.fn(),
    addListener: jest.fn(() => jest.fn()),
  });

  beforeEach(() => {
    jest.clearAllMocks();
    jest.spyOn(console, 'warn').mockImplementation(() => {});
    (api.createChatTempSession as jest.Mock).mockResolvedValue({
      id: 'temp-1',
      status: 'active',
      forced_new_topic: false,
    });
    (config.clearToken as jest.Mock).mockResolvedValue(undefined);
  });

  afterEach(() => {
    (console.warn as jest.Mock).mockRestore();
  });

  it('starts with a temp session and a low-pressure welcome message', async () => {
    const navigation = makeNavigation();
    const {findByText} = render(<ChatScreen navigation={navigation} />);

    await waitFor(() => {
      expect(api.createChatTempSession).toHaveBeenCalledWith(false);
    });
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();
  }, 10000);

  it('keeps recoverable chat initialization failures out of the warning overlay', async () => {
    (api.createChatTempSession as jest.Mock).mockRejectedValueOnce(new Error('gateway down'));

    const navigation = makeNavigation();
    const {findByText} = render(<ChatScreen navigation={navigation} />);

    expect(await findByText('现在连不上聊聊服务。你可以先把想说的放在这里，稍后再试。')).toBeTruthy();
    expect(console.warn).not.toHaveBeenCalled();
  });

  it('redirects to login when chat initialization gets an expired auth response', async () => {
    (api.createChatTempSession as jest.Mock).mockRejectedValueOnce({status: 401});
    const parentNavigation = {navigate: jest.fn()};
    const navigation = {
      ...makeNavigation(),
      getParent: jest.fn(() => parentNavigation),
    };

    render(<ChatScreen navigation={navigation} />);

    await waitFor(() => {
      expect(config.clearToken).toHaveBeenCalled();
      expect(parentNavigation.navigate).toHaveBeenCalledWith('login');
    });
  });

  it('renders reference cards and save suggestion after a successful temp-session reply', async () => {
    const navigation = makeNavigation();
    (api.sendTempChatMessage as jest.Mock).mockResolvedValue({
      user_message: {
        id: 'user-1',
        role: 'user',
        content: '我老是拖着不开始',
        created_at: '2026-05-26T00:00:00Z',
      },
      message: {
        id: 'assistant-1',
        role: 'assistant',
        content: '先别急着把自己说服，今天只把第一步做小。',
        created_at: '2026-05-26T00:00:01Z',
      },
      reference_cards: [
        {
          experience_id: 'exp-1',
          content: '行动不一定要等想清楚，先做一小步会带来新的判断。',
          is_collected: false,
        },
      ],
      note_suggestion: {
        should_show: true,
        suggested_text: '先做一小步，再用结果修正判断。',
        source_message_ids: ['user-1', 'assistant-1'],
      },
    });
    (api.setCollected as jest.Mock).mockResolvedValue({collected: true});

    const {findByLabelText, findByText, getByLabelText, getByPlaceholderText, getByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '我老是拖着不开始');
    fireEvent.press(getByText('发送'));

    expect(await findByText('先别急着把自己说服，今天只把第一步做小。')).toBeTruthy();
    expect(await findByText('参考经验')).toBeTruthy();
    await waitFor(() => {
      expect(api.recordExperienceEvent).toHaveBeenCalledWith(
        'exp-1',
        'chat_citation_show',
        'chat',
        expect.objectContaining({
          message_id: 'assistant-1',
          temp_session_id: 'temp-1',
        }),
      );
    });

    fireEvent.press(getByText('行动不一定要等想清楚，先做一小步会带来新的判断。'));
    expect(api.recordExperienceEvent).toHaveBeenCalledWith(
      'exp-1',
      'chat_citation_click',
      'chat',
      expect.objectContaining({
        message_id: 'assistant-1',
        temp_session_id: 'temp-1',
      }),
    );
    expect(navigation.navigate).toHaveBeenCalledWith('detail', {id: 'exp-1', from: 'chat'});

    fireEvent.press(getByLabelText('收藏参考经验'));
    await waitFor(() => {
      expect(api.setCollected).toHaveBeenCalledWith('exp-1', true);
    });
    expect(await findByLabelText('已收藏参考经验')).toBeTruthy();

    fireEvent.press(getByText('记下'));
    expect(navigation.navigate).toHaveBeenCalledWith('create', {
      prefillContent: '先做一小步，再用结果修正判断。',
      defaultVisibility: 'private',
      sourceScene: 'chat',
      sourceMessageIds: ['user-1', 'assistant-1'],
    });
  });

  it('clears expired auth when collecting a reference card returns 401', async () => {
    const parentNavigation = {navigate: jest.fn()};
    const navigation = {
      ...makeNavigation(),
      getParent: jest.fn(() => parentNavigation),
    };
    (api.sendTempChatMessage as jest.Mock).mockResolvedValue({
      user_message: {id: 'user-1', role: 'user', content: '我老是拖着不开始'},
      message: {id: 'assistant-1', role: 'assistant', content: '先把第一步做小。'},
      reference_cards: [{
        experience_id: 'exp-1',
        content: '行动不一定要等想清楚，先做一小步会带来新的判断。',
        is_collected: false,
      }],
    });
    (api.setCollected as jest.Mock).mockRejectedValueOnce({status: 401});

    const {findByText, getByLabelText, getByPlaceholderText, getByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '我老是拖着不开始');
    fireEvent.press(getByText('发送'));
    expect(await findByText('先把第一步做小。')).toBeTruthy();

    fireEvent.press(getByLabelText('收藏参考经验'));

    await waitFor(() => {
      expect(config.clearToken).toHaveBeenCalled();
      expect(parentNavigation.navigate).toHaveBeenCalledWith('login');
    });
  });

  it('redirects to login when opening recent topics gets an expired auth response', async () => {
    const parentNavigation = {navigate: jest.fn()};
    const navigation = {
      ...makeNavigation(),
      getParent: jest.fn(() => parentNavigation),
    };
    (api.fetchRecentChatTopics as jest.Mock).mockRejectedValueOnce({status: 401});

    const {findByLabelText, findByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.press(await findByLabelText('打开议题列表'));

    await waitFor(() => {
      expect(config.clearToken).toHaveBeenCalled();
      expect(parentNavigation.navigate).toHaveBeenCalledWith('login');
    });
  });

  it('keeps the user message and shows retry when the AI reply fails', async () => {
    const navigation = makeNavigation();
    (api.sendTempChatMessage as jest.Mock).mockRejectedValueOnce(new Error('gateway timeout'));

    const {findByText, getByPlaceholderText, getByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '我今天很乱');
    fireEvent.press(getByText('发送'));

    expect(await findByText('我今天很乱')).toBeTruthy();
    expect(await findByText('抱歉，对话服务暂时不可用，请稍后再试。')).toBeTruthy();
    expect(await findByText('重试')).toBeTruthy();
  });

  it('replaces a failed retry bubble with expired-auth copy when retry returns 401', async () => {
    const parentNavigation = {navigate: jest.fn()};
    const navigation = {
      ...makeNavigation(),
      getParent: jest.fn(() => parentNavigation),
    };
    (api.sendTempChatMessage as jest.Mock)
      .mockRejectedValueOnce(new Error('gateway timeout'))
      .mockRejectedValueOnce({status: 401});

    const {findByText, getByPlaceholderText, getByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '我今天很乱');
    fireEvent.press(getByText('发送'));

    expect(await findByText('抱歉，对话服务暂时不可用，请稍后再试。')).toBeTruthy();
    fireEvent.press(getByText('重试'));

    await waitFor(() => {
      expect(config.clearToken).toHaveBeenCalled();
      expect(parentNavigation.navigate).toHaveBeenCalledWith('login');
    });
    expect(await findByText('登录状态过期了，重新登录后可以继续聊。')).toBeTruthy();
  });
});
