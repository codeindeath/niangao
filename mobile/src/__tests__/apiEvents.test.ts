import {recordSearchClick, recordView} from '../services/api';
import {apiPost, getToken} from '../services/config';

jest.mock('../services/config', () => ({
  apiPost: jest.fn(),
  getToken: jest.fn(),
}));

describe('experience event tracking', () => {
  let consoleWarnSpy: jest.SpyInstance;

  beforeEach(() => {
    jest.clearAllMocks();
    consoleWarnSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});
    (apiPost as jest.Mock).mockResolvedValue({});
  });

  afterEach(() => {
    consoleWarnSpy.mockRestore();
  });

  it('records search clicks through the V4 passive event endpoint', () => {
    recordSearchClick('exp-search-1', '姜文', 2);

    expect(apiPost).toHaveBeenCalledWith('/api/v1/experiences/exp-search-1/events', {
      event_type: 'search_click',
      source_context: 'search',
      metadata: {query: '姜文', rank: 2},
    });
  });

  it('records guest card exposure through the V4 passive event endpoint without legacy auth-only view', async () => {
    (getToken as jest.Mock).mockResolvedValue(null);

    recordView('exp-view-guest');

    expect(apiPost).toHaveBeenCalledWith('/api/v1/experiences/exp-view-guest/events', {
      event_type: 'expose',
      source_context: 'feed',
      metadata: {},
    });
    await Promise.resolve();
    expect(apiPost).not.toHaveBeenCalledWith('/api/v1/experiences/exp-view-guest/view', {});
  });

  it('records authenticated exposure only through the V4 passive event endpoint', async () => {
    (getToken as jest.Mock).mockResolvedValue('token-1');

    recordView('exp-view-auth');
    await Promise.resolve();

    expect(apiPost).toHaveBeenCalledWith('/api/v1/experiences/exp-view-auth/events', {
      event_type: 'expose',
      source_context: 'feed',
      metadata: {},
    });
    expect(apiPost).not.toHaveBeenCalledWith('/api/v1/experiences/exp-view-auth/view', {});
  });

  it('keeps passive event failures out of the React Native warning overlay', async () => {
    (apiPost as jest.Mock).mockRejectedValueOnce(new Error('network down'));

    recordSearchClick('exp-search-fail', '关系', 0);
    await Promise.resolve();

    expect(consoleWarnSpy).not.toHaveBeenCalled();
  });

  it('does not call the legacy auth-only view endpoint on exposure failure paths', async () => {
    (getToken as jest.Mock).mockResolvedValue('token-1');
    (apiPost as jest.Mock).mockResolvedValueOnce({});

    recordView('exp-view-no-legacy');
    await Promise.resolve();
    await Promise.resolve();

    expect(apiPost).not.toHaveBeenCalledWith('/api/v1/experiences/exp-view-no-legacy/view', {});
    expect(consoleWarnSpy).not.toHaveBeenCalled();
  });

  it('allows exposure to retry after a passive event failure', async () => {
    (apiPost as jest.Mock)
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce({});

    recordView('exp-view-retry-after-failure');
    await Promise.resolve();

    recordView('exp-view-retry-after-failure');
    await Promise.resolve();

    expect(apiPost).toHaveBeenCalledTimes(2);
    expect(apiPost).toHaveBeenNthCalledWith(2, '/api/v1/experiences/exp-view-retry-after-failure/events', {
      event_type: 'expose',
      source_context: 'feed',
      metadata: {},
    });
  });
});
