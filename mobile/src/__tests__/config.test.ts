import {apiGet, apiPost, ApiError} from '../services/config';
import AsyncStorage from '@react-native-async-storage/async-storage';

describe('config API errors', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useRealTimers();
    (AsyncStorage.getItem as jest.Mock).mockResolvedValue('token-1');
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('preserves structured backend error codes', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
      status: 400,
      headers: {get: jest.fn().mockReturnValue('server-request-1')},
      text: jest.fn().mockResolvedValue(JSON.stringify({
        error: {
          code: 'display_name_required',
          message: '需要先设置展示名',
        },
      })),
    });

    await expect(apiPost('/api/v1/experiences', {})).rejects.toMatchObject({
      status: 400,
      message: '需要先设置展示名',
      code: 'display_name_required',
      requestId: 'server-request-1',
    });
    await expect(apiPost('/api/v1/experiences', {})).rejects.toBeInstanceOf(ApiError);
  });

  it('preserves top-level structured error action fields', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
      status: 503,
      headers: {get: jest.fn().mockReturnValue('chat-request-1')},
      text: jest.fn().mockResolvedValue(JSON.stringify({
        error: {
          code: 'chat_service_unavailable',
          message: '暂时聊不了，请稍后再试。',
          request_id: 'chat-error-request-1',
        },
        retryable: true,
        user_message_id: 'msg-user-1',
      })),
    });

    await expect(apiPost('/api/v1/chat/temp-sessions/temp-1/messages', {content: '刚才那句继续'})).rejects.toMatchObject({
      status: 503,
      message: '暂时聊不了，请稍后再试。',
      code: 'chat_service_unavailable',
      requestId: 'chat-error-request-1',
      retryable: true,
      userMessageId: 'msg-user-1',
    });
  });

  it('sends a request id header for backend traceability', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      text: jest.fn().mockResolvedValue(JSON.stringify({status: 'ok'})),
    });

    await apiPost('/api/v1/experiences', {content: '今天先把事情做小一点'});

    expect(global.fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: expect.objectContaining({
          'X-Request-ID': expect.stringMatching(/^app-/),
        }),
      }),
    );
  });

  it('uses sibling backend message when legacy string errors include a user-facing message', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
      status: 429,
      headers: {get: jest.fn().mockReturnValue('quota-request-1')},
      text: jest.fn().mockResolvedValue(JSON.stringify({
        error: 'chat_quota_exceeded',
        message: '今日对话已达上限（50轮），明天再来聊吧。',
        retryable: false,
      })),
    });

    await expect(apiPost('/api/v1/chat/temp-sessions/temp-1/messages', {content: '再聊一句'})).rejects.toMatchObject({
      status: 429,
      message: '今日对话已达上限（50轮），明天再来聊吧。',
      code: 'chat_quota_exceeded',
      requestId: 'quota-request-1',
    });
  });

  it('does not expose plain technical HTTP error bodies to the App', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
      status: 502,
      headers: {get: jest.fn().mockReturnValue('proxy-request-1')},
      text: jest.fn().mockResolvedValue('Bad Gateway'),
    });

    await expect(apiGet('/api/v1/feed/recommend?limit=2')).rejects.toMatchObject({
      status: 502,
      message: '请求失败',
      requestId: 'proxy-request-1',
    });
  });

  it('aborts slow normal API requests with a structured timeout error', async () => {
    jest.useFakeTimers();
    let signal: AbortSignal | undefined;
    (global.fetch as jest.Mock).mockImplementation((_url: string, init?: {signal?: AbortSignal}) => {
      signal = init?.signal;
      if (!signal) return Promise.reject(new Error('missing abort signal'));
      return new Promise((_resolve, reject) => {
        signal!.addEventListener('abort', () => {
          const err = new Error('Aborted');
          err.name = 'AbortError';
          reject(err);
        });
      });
    });

    const request = apiGet('/api/v1/feed/recommend?limit=2');
    await jest.advanceTimersByTimeAsync(0);
    expect(signal).toBeDefined();
    const requestInit = (global.fetch as jest.Mock).mock.calls[0][1];
    const requestId = requestInit.headers['X-Request-ID'];

    jest.advanceTimersByTime(15_000);

    await expect(request).rejects.toMatchObject({
      status: 0,
      code: 'request_timeout',
      requestId,
    });
  });

  it('wraps network failures with the same weak-network user copy', async () => {
    (global.fetch as jest.Mock).mockRejectedValue(new Error('Network request failed'));

    const request = apiGet('/api/v1/feed/recommend?limit=2');

    await expect(request).rejects.toMatchObject({
      status: 0,
      code: 'network_failed',
      message: '网络不稳，请稍后再试',
      requestId: expect.stringMatching(/^app-/),
    });
    await expect(request).rejects.toBeInstanceOf(ApiError);
  });

  it.each([
    ['chat', '/api/v1/chat/temp-sessions/session-1/messages', {content: '今天有点乱'}],
    ['rewrite', '/api/v1/experiences/rewrite', {content: '今天有点乱'}],
  ])('gives %s API calls a longer timeout budget', async (_name, path, body) => {
    jest.useFakeTimers();
    let signal: AbortSignal | undefined;
    (global.fetch as jest.Mock).mockImplementation((_url: string, init?: {signal?: AbortSignal}) => {
      signal = init?.signal;
      if (!signal) return Promise.reject(new Error('missing abort signal'));
      return new Promise((_resolve, reject) => {
        signal!.addEventListener('abort', () => {
          const err = new Error('Aborted');
          err.name = 'AbortError';
          reject(err);
        });
      });
    });

    const request = apiPost(path as string, body);
    await jest.advanceTimersByTimeAsync(0);
    expect(signal).toBeDefined();
    const requestInit = (global.fetch as jest.Mock).mock.calls[0][1];
    const requestId = requestInit.headers['X-Request-ID'];

    jest.advanceTimersByTime(15_000);
    expect(signal!.aborted).toBe(false);

    jest.advanceTimersByTime(45_000);

    await expect(request).rejects.toMatchObject({
      status: 0,
      code: 'request_timeout',
      requestId,
    });
  });
});
