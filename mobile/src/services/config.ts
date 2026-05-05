import AsyncStorage from '@react-native-async-storage/async-storage';

// 后端服务地址（ECS）
export const API_BASE = 'http://115.190.177.146';  // Nginx 反代，不需要直连端口
export const AI_BASE = 'http://115.190.177.146:8000';

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'user_info';

// ---------- Token 管理 ----------
export async function getToken(): Promise<string | null> {
  return AsyncStorage.getItem(TOKEN_KEY);
}

export async function setToken(token: string): Promise<void> {
  await AsyncStorage.setItem(TOKEN_KEY, token);
}

export async function clearToken(): Promise<void> {
  await AsyncStorage.removeItem(TOKEN_KEY);
  await AsyncStorage.removeItem(USER_KEY);
}

// ---------- 用户信息 ----------
export async function getUserInfo(): Promise<any | null> {
  const data = await AsyncStorage.getItem(USER_KEY);
  return data ? JSON.parse(data) : null;
}

export async function setUserInfo(user: any): Promise<void> {
  await AsyncStorage.setItem(USER_KEY, JSON.stringify(user));
}

// ---------- Go 后端 HTTP 请求 ----------
export async function apiGet(path: string): Promise<any> {
  const token = await getToken();
  const res = await fetch(`${API_BASE}${path}`, {
    headers: token ? {Authorization: `Bearer ${token}`} : {},
  });
  const text = await res.text();
  if (!res.ok) {
    let message = text;
    try {
      const json = JSON.parse(text);
      message = json.error || json.message || text;
    } catch {}
    throw new ApiError(res.status, message);
  }
  try { return JSON.parse(text); } catch { return text; }
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = 'ApiError';
  }
}

export async function apiPost(path: string, body: any): Promise<any> {
  const token = await getToken();
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? {Authorization: `Bearer ${token}`} : {}),
    },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  if (!res.ok) {
    let message = text;
    try {
      const json = JSON.parse(text);
      message = json.error || json.message || text;
    } catch {}
    throw new ApiError(res.status, message);
  }
  try { return JSON.parse(text); } catch { return text; }
}

// ---------- AI 服务 HTTP 请求 ----------
export async function aiPost(path: string, body: any): Promise<any> {
  const token = await getToken();
  const res = await fetch(`${AI_BASE}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? {Authorization: `Bearer ${token}`} : {}),
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const errBody = await res.text();
    throw new Error(`HTTP ${res.status}: ${errBody}`);
  }
  return res.json();
}
