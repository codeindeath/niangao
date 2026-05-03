import AsyncStorage from '@react-native-async-storage/async-storage';

// 后端服务地址（ECS）
export const API_BASE = 'http://115.190.177.146:8080';
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
  return res.json();
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
  return res.json();
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
  return res.json();
}
