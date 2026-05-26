import React from 'react';
import {Alert} from 'react-native';
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
    Object.values(api).forEach((mockFn: unknown) => {
      if (jest.isMockFunction(mockFn)) mockFn.mockReset();
    });
    (config.clearToken as jest.Mock).mockReset();
    jest.spyOn(console, 'warn').mockImplementation(() => {});
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
    (api.createChatTempSession as jest.Mock).mockResolvedValue({
      id: 'temp-1',
      status: 'active',
      forced_new_topic: false,
    });
    (api.fetchRecentChatTopics as jest.Mock).mockResolvedValue({data: []});
    (api.fetchChatTopicMessages as jest.Mock).mockResolvedValue({data: []});
    (config.clearToken as jest.Mock).mockResolvedValue(undefined);
  });

  afterEach(() => {
    (console.warn as jest.Mock).mockRestore();
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('starts with a temp session and a low-pressure welcome message', async () => {
    (api.fetchRecentChatTopics as jest.Mock).mockResolvedValueOnce({data: []});
    const navigation = makeNavigation();
    const {findByText} = render(<ChatScreen navigation={navigation} />);

    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。', {}, {timeout: 10000})).toBeTruthy();
    expect(api.createChatTempSession).toHaveBeenCalledWith(false);
  }, 10000);

  it('resumes a recently active stable topic when entering chat', async () => {
    (api.fetchRecentChatTopics as jest.Mock).mockResolvedValueOnce({
      data: [{
        id: 'topic-recent',
        status: 'active',
        title: '工作里的不甘心',
        domain: 'work',
        updated_at: new Date().toISOString(),
      }],
    });
    (api.fetchChatTopicMessages as jest.Mock).mockResolvedValueOnce({
      data: [{
        id: 'assistant-history',
        topic_id: 'topic-recent',
        role: 'assistant',
        content: '上次我们聊到，被当众否定之后你一直有点过不去。',
        created_at: '2026-05-26T00:00:01Z',
      }],
    });

    const navigation = makeNavigation();
    const {findByText} = render(<ChatScreen navigation={navigation} />);

    expect(await findByText('上次我们聊到，被当众否定之后你一直有点过不去。')).toBeTruthy();
    expect(await findByText('工作里的不甘心')).toBeTruthy();
    expect(api.createChatTempSession).not.toHaveBeenCalled();
  });

  it('starts a temp session when the latest stable topic is stale', async () => {
    (api.fetchRecentChatTopics as jest.Mock).mockResolvedValueOnce({
      data: [{
        id: 'topic-old',
        status: 'active',
        title: '很久以前的议题',
        updated_at: '2026-01-01T00:00:00Z',
      }],
    });

    const navigation = makeNavigation();
    const {findByText} = render(<ChatScreen navigation={navigation} />);

    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();
    expect(api.createChatTempSession).toHaveBeenCalledWith(false);
    expect(api.fetchChatTopicMessages).not.toHaveBeenCalled();
  });

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
      sourceChatMessageId: 'assistant-1',
      sourceChatMessageSnapshot: 'user-1,assistant-1',
    });
  });

  it('switches to the promoted stable topic after a temp-session reply is classified', async () => {
    const navigation = makeNavigation();
    (api.sendTempChatMessage as jest.Mock).mockResolvedValueOnce({
      user_message: {
        id: 'user-1',
        topic_id: 'topic-promoted',
        role: 'user',
        content: '我觉得在会上被上级当众否定这事过不去',
        created_at: '2026-05-26T00:00:00Z',
      },
      message: {
        id: 'assistant-1',
        topic_id: 'topic-promoted',
        role: 'assistant',
        content: '这听起来不是一句评价那么简单，更像是边界被当众推开了。',
        created_at: '2026-05-26T00:00:01Z',
      },
      reference_cards: [
        {
          experience_id: 'exp-1',
          content: '公开场合被评价时，先分清事实和被侵犯的边界。',
          is_collected: false,
        },
      ],
      note_suggestion: {should_show: false, source_message_ids: []},
      session_state: 'stable_topic',
      promoted_topic: {
        id: 'topic-promoted',
        status: 'active',
        title: '工作里的不甘心',
        domain: 'work',
        sub_domain: 'work-comm',
        topic: '和上级沟通',
      },
    });
    (api.sendChatTopicMessage as jest.Mock).mockResolvedValueOnce({
      user_message: {
        id: 'user-2',
        topic_id: 'topic-promoted',
        role: 'user',
        content: '后来我一直在想同事怎么看我',
        created_at: '2026-05-26T00:00:02Z',
      },
      message: {
        id: 'assistant-2',
        topic_id: 'topic-promoted',
        role: 'assistant',
        content: '那种反复猜测，可能是被公开评价后留下的紧绷。',
        created_at: '2026-05-26T00:00:03Z',
      },
      reference_cards: [],
      note_suggestion: {should_show: false, source_message_ids: []},
    });

    const {findByText, getByPlaceholderText, getByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '我觉得在会上被上级当众否定这事过不去');
    fireEvent.press(getByText('发送'));

    expect(await findByText('这听起来不是一句评价那么简单，更像是边界被当众推开了。')).toBeTruthy();
    expect(await findByText('工作里的不甘心')).toBeTruthy();
    await waitFor(() => {
      expect(api.recordExperienceEvent).toHaveBeenCalledWith(
        'exp-1',
        'chat_citation_show',
        'chat',
        expect.objectContaining({
          message_id: 'assistant-1',
          topic_id: 'topic-promoted',
        }),
      );
    });

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '后来我一直在想同事怎么看我');
    fireEvent.press(getByText('发送'));

    await waitFor(() => {
      expect(api.sendChatTopicMessage).toHaveBeenCalledWith(
        'topic-promoted',
        '后来我一直在想同事怎么看我',
        expect.any(String),
      );
    });
    expect(api.sendTempChatMessage).toHaveBeenCalledTimes(1);
  });

  it('restores historical unavailable reference cards without leaking content', async () => {
    const navigation = makeNavigation();
    (api.fetchRecentChatTopics as jest.Mock)
      .mockResolvedValueOnce({data: []})
      .mockResolvedValueOnce({
        data: [{id: 'topic-1', title: '旧议题', domain: 'work'}],
      });
    (api.fetchChatTopicMessages as jest.Mock).mockResolvedValue({
      data: [{
        id: 'assistant-history',
        role: 'assistant',
        content: '当时我们说，可以先把第一步做小。',
        created_at: '2026-05-26T00:00:01Z',
        reference_cards: [{
          experience_id: 'exp-hidden',
          content: '这段历史引用原文不应该显示',
          is_collected: false,
          unavailable_reason: 'experience_unavailable',
        }],
      }],
    });

    const {findByLabelText, findByText, queryByLabelText, queryByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.press(await findByLabelText('打开议题列表'));
    fireEvent.press(await findByText('旧议题'));

    expect(await findByText('当时我们说，可以先把第一步做小。')).toBeTruthy();
    expect(await findByText('该经验已不可见')).toBeTruthy();
    expect(await findByText('它可能已经被删除、转为私密，或正在重新处理。')).toBeTruthy();
    expect(queryByText('这段历史引用原文不应该显示')).toBeNull();
    expect(queryByLabelText('收藏参考经验')).toBeNull();
  });

  it('starts a forced new temp session from the recent-topic panel', async () => {
    const navigation = makeNavigation();
    (api.fetchRecentChatTopics as jest.Mock)
      .mockResolvedValueOnce({data: []})
      .mockResolvedValueOnce({
        data: [{
          id: 'topic-1',
          status: 'active',
          title: '旧议题',
          domain: 'work',
        }],
      });

    const {findByLabelText, findByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.press(await findByLabelText('打开议题列表'));
    fireEvent.press(await findByText('换个事聊'));

    await waitFor(() => {
      expect(api.createChatTempSession).toHaveBeenLastCalledWith(true);
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

  it('shows action failure feedback when collecting a reference card fails for a non-auth error', async () => {
    const navigation = makeNavigation();
    (api.sendTempChatMessage as jest.Mock).mockResolvedValue({
      user_message: {id: 'user-1', role: 'user', content: '我老是拖着不开始'},
      message: {id: 'assistant-1', role: 'assistant', content: '先把第一步做小。'},
      reference_cards: [{
        experience_id: 'exp-1',
        content: '行动不一定要等想清楚，先做一小步会带来新的判断。',
        is_collected: false,
      }],
    });
    (api.setCollected as jest.Mock).mockRejectedValueOnce(new Error('network down'));

    const {findByText, getByLabelText, getByPlaceholderText, getByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '我老是拖着不开始');
    fireEvent.press(getByText('发送'));
    expect(await findByText('先把第一步做小。')).toBeTruthy();

    fireEvent.press(getByLabelText('收藏参考经验'));

    await waitFor(() => {
      expect(Alert.alert).toHaveBeenCalledWith('操作失败', 'network down');
    });
    expect(config.clearToken).not.toHaveBeenCalled();
  });

  it('redirects to login when opening recent topics gets an expired auth response', async () => {
    const parentNavigation = {navigate: jest.fn()};
    const navigation = {
      ...makeNavigation(),
      getParent: jest.fn(() => parentNavigation),
    };
    (api.fetchRecentChatTopics as jest.Mock)
      .mockResolvedValueOnce({data: []})
      .mockRejectedValueOnce({status: 401});

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

  it('uses the backend quota message when the AI reply returns 429', async () => {
    const navigation = makeNavigation();
    (api.sendTempChatMessage as jest.Mock).mockRejectedValueOnce({
      status: 429,
      message: '今日对话已达上限（50轮），明天再来聊吧。',
    });

    const {findByText, getByPlaceholderText, getByText, queryByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '我还想再聊一轮');
    fireEvent.press(getByText('发送'));

    expect(await findByText('今日对话已达上限（50轮），明天再来聊吧。')).toBeTruthy();
    expect(queryByText('今日对话已达上限（100轮），明天再来聊吧。')).toBeNull();
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

  it('uses the backend quota message when retrying an AI reply returns 429', async () => {
    const navigation = makeNavigation();
    (api.sendTempChatMessage as jest.Mock)
      .mockRejectedValueOnce(new Error('gateway timeout'))
      .mockRejectedValueOnce({
        status: 429,
        message: '今日对话已达上限（50轮），明天再来聊吧。',
      });

    const {findByText, getByPlaceholderText, getByText, queryByText} = render(<ChatScreen navigation={navigation} />);
    expect(await findByText('我在。你可以从任何一点开始说，不用先想清楚。')).toBeTruthy();

    fireEvent.changeText(getByPlaceholderText('输入你想聊的...'), '我今天很乱');
    fireEvent.press(getByText('发送'));

    expect(await findByText('抱歉，对话服务暂时不可用，请稍后再试。')).toBeTruthy();
    fireEvent.press(getByText('重试'));

    expect(await findByText('今日对话已达上限（50轮），明天再来聊吧。')).toBeTruthy();
    expect(queryByText('还是没连上。你这条消息已经保留了，可以稍后再试。')).toBeNull();
  });
});
