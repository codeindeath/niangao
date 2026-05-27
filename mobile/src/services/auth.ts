import {
  apiFetchWithTimeout,
  clearToken,
  getToken,
  isTokenExpired,
  refreshStoredAuthToken,
} from './config';

// 登出：先调服务端吊销，再清本地
export async function logout(): Promise<void> {
  try {
    const token = await getToken();
    if (token) {
      await apiFetchWithTimeout('/api/v1/auth/logout', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
    }
  } catch {
    // 网络问题也继续清本地
  }
  await clearToken();
}

// 刷新 token
export async function refreshToken(): Promise<boolean> {
  return refreshStoredAuthToken();
}

// 检查登录状态：token 存在且未过期
export async function isLoggedIn(): Promise<boolean> {
  try {
    const token = await getToken();
    if (!token) return false;

    // 检查是否过期
    if (isTokenExpired(token)) {
      // 尝试刷新
      const refreshed = await refreshToken();
      return refreshed;
    }

    return true;
  } catch {
    return false;
  }
}
