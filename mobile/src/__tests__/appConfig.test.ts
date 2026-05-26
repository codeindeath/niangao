const appConfig = require('../../app.json');

function loadDefaultApiBase(): string {
  let apiBase = '';
  jest.isolateModules(() => {
    const originalApiBase = process.env.EXPO_PUBLIC_API_BASE;
    delete process.env.EXPO_PUBLIC_API_BASE;
    apiBase = require('../services/config').API_BASE;
    if (originalApiBase === undefined) {
      delete process.env.EXPO_PUBLIC_API_BASE;
    } else {
      process.env.EXPO_PUBLIC_API_BASE = originalApiBase;
    }
  });
  return apiBase;
}

describe('Expo production app configuration', () => {
  beforeEach(() => {
    jest.resetModules();
  });

  it('keeps Apple login and the production iOS bundle identifier configured', () => {
    const ios = appConfig.expo.ios;

    expect(ios.bundleIdentifier).toBe('com.swt.niangaogao');
    expect(ios.usesAppleSignIn).toBe(true);
    expect(appConfig.expo.plugins).toEqual(
      expect.arrayContaining([
        'expo-apple-authentication',
        './plugins/withDynamicIosBundleUrl',
      ]),
    );
  });

  it('keeps the default HTTP API host covered by an iOS ATS exception without arbitrary loads', () => {
    const defaultApiBase = loadDefaultApiBase();
    const defaultHost = new URL(defaultApiBase).hostname;
    const ats = appConfig.expo.ios?.infoPlist?.NSAppTransportSecurity;

    expect(defaultApiBase).toBe('http://115.190.177.146');
    expect(ats?.NSAllowsArbitraryLoads).toBe(false);
    expect(ats?.NSAllowsLocalNetworking).toBe(true);
    expect(ats?.NSExceptionDomains?.[defaultHost]?.NSExceptionAllowsInsecureHTTPLoads).toBe(true);
  });
});
