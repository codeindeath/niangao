import * as WeChat from 'react-native-wechat-lib';
import { apiPost, setToken, setUserInfo, clearToken } from './config';

// 替换为你的微信 AppID
const WECHAT_APP_ID = 'your_wechat_app_id';

export function registerWechat(): void {
  WeChat.registerApp(WECHAT_APP_ID);
}

// 微信登录
export async function loginWithWechat(): Promise<{ success: boolean; user?: any; error?: string }> {
  try {
    // 1. 向微信请求授权，拿到 code
    const authResp = await WeChat.sendAuthRequest('snsapi_userinfo', 'niangao_login');
    if (authResp.errCode !== 0) {
      // 用户取消或授权失败
      return { success: false, error: '用户取消登录' };
    }

    // 2. 把 code 发给后端，换取 JWT
    const data = await apiPost('/api/v1/auth/wechat/login', { code: authResp.code });

    if (data.error) {
      return { success: false, error: data.error };
    }

    // 3. 保存 token 和用户信息
    await setToken(data.token);
    await setUserInfo(data.user);

    return { success: true, user: data.user };
  } catch (e: any) {
    return { success: false, error: e.message || '登录失败' };
  }
}

export async function logout(): Promise<void> {
  await clearToken();
}

export async function isLoggedIn(): Promise<boolean> {
  const { getToken } = require('./config');
  const token = await getToken();
  return !!token;
}
