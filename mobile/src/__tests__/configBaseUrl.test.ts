function loadConfigModule(): typeof import('../services/config') {
  let mod: typeof import('../services/config') | undefined;
  jest.isolateModules(() => {
    mod = require('../services/config');
  });
  if (!mod) throw new Error('config module was not loaded');
  return mod;
}

describe('API base URL configuration', () => {
  const originalApiBase = process.env.EXPO_PUBLIC_API_BASE;

  beforeEach(() => {
    jest.resetModules();
    jest.clearAllMocks();
    delete process.env.EXPO_PUBLIC_API_BASE;
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetModules();
    if (originalApiBase === undefined) {
      delete process.env.EXPO_PUBLIC_API_BASE;
    } else {
      process.env.EXPO_PUBLIC_API_BASE = originalApiBase;
    }
  });

  it('defaults to the production API host when no app env override is set', () => {
    const {API_BASE} = loadConfigModule();

    expect(API_BASE).toBe('http://115.190.177.146');
  });

  it('uses EXPO_PUBLIC_API_BASE and removes trailing slashes for simulator and staging checks', () => {
    process.env.EXPO_PUBLIC_API_BASE = 'http://127.0.0.1:8080///';

    const {API_BASE} = loadConfigModule();

    expect(API_BASE).toBe('http://127.0.0.1:8080');
  });

  it('auth calls use the same configurable API base', async () => {
    process.env.EXPO_PUBLIC_API_BASE = 'http://127.0.0.1:8080/';
    (global.fetch as jest.Mock).mockResolvedValue({ok: true});
    let logout: typeof import('../services/auth').logout | undefined;
    jest.isolateModules(() => {
      const asyncStorage = require('@react-native-async-storage/async-storage');
      asyncStorage.getItem.mockResolvedValue('token-1');
      logout = require('../services/auth').logout;
    });

    if (!logout) throw new Error('logout was not loaded');
    await logout();

    expect(global.fetch).toHaveBeenCalledWith(
      'http://127.0.0.1:8080/api/v1/auth/logout',
      expect.objectContaining({
        signal: expect.any(Object),
      }),
    );
  });
});
