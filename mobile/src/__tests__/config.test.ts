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
    });
    await expect(apiPost('/api/v1/experiences', {})).rejects.toBeInstanceOf(ApiError);
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
    const assertion = expect(request).rejects.toMatchObject({
      status: 0,
      code: 'request_timeout',
    });
    await jest.advanceTimersByTimeAsync(0);
    expect(signal).toBeDefined();

    jest.advanceTimersByTime(15_000);

    await assertion;
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
    const assertion = expect(request).rejects.toMatchObject({
      status: 0,
      code: 'request_timeout',
    });
    await jest.advanceTimersByTimeAsync(0);
    expect(signal).toBeDefined();

    jest.advanceTimersByTime(15_000);
    expect(signal!.aborted).toBe(false);

    jest.advanceTimersByTime(45_000);

    await assertion;
  });
});
