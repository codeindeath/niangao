import AsyncStorage from '@react-native-async-storage/async-storage';

// 后端服务地址（ECS）
declare const process: {env?: Record<string, string | undefined>};

const DEFAULT_API_BASE = 'http://115.190.177.146';

function normalizeApiBase(value?: string): string {
  const trimmed = value?.trim();
  if (!trimmed) return DEFAULT_API_BASE;
  return trimmed.replace(/\/+$/, '');
}

export const API_BASE = normalizeApiBase(process.env?.EXPO_PUBLIC_API_BASE);

const TOKEN_KEY = 'auth_token';
const REFRESH_KEY = 'refresh_token';
const USER_KEY = 'user_info';
const STANDARD_REQUEST_TIMEOUT_MS = 15_000;
const AI_REQUEST_TIMEOUT_MS = 60_000;

type ParsedErrorPayload = {
  message: string;
  code?: string;
  requestId?: string;
  retryable?: boolean;
  userMessageId?: string;
};

// ---------- Token 管理 ----------
export async function getToken(): Promise<string | null> {
  return AsyncStorage.getItem(TOKEN_KEY);
}

export async function setToken(token: string): Promise<void> {
  await AsyncStorage.setItem(TOKEN_KEY, token);
}

export async function clearToken(): Promise<void> {
  await AsyncStorage.removeItem(TOKEN_KEY);
  await AsyncStorage.removeItem(REFRESH_KEY);
  await AsyncStorage.removeItem(USER_KEY);
}

// ---------- Refresh Token ----------
export async function getRefreshToken(): Promise<string | null> {
  return AsyncStorage.getItem(REFRESH_KEY);
}

export async function setRefreshToken(token: string): Promise<void> {
  await AsyncStorage.setItem(REFRESH_KEY, token);
}

// ---------- 用户信息 ----------
export async function getUserInfo(): Promise<any | null> {
  const data = await AsyncStorage.getItem(USER_KEY);
  return data ? JSON.parse(data) : null;
}

export async function setUserInfo(user: any): Promise<void> {
  await AsyncStorage.setItem(USER_KEY, JSON.stringify(user));
}

// ---------- Token 过期检查 ----------
export function isTokenExpired(token: string): boolean {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return true;
    const payload = JSON.parse(atob(parts[1]));
    if (!payload.exp) return false;
    return Date.now() >= payload.exp * 1000;
  } catch {
    return true; // 解析失败视为过期
  }
}

function parseErrorPayload(text: string): ParsedErrorPayload {
  try {
    const json = JSON.parse(text);
    const actionFields = {
      retryable: typeof json.retryable === 'boolean' ? json.retryable : undefined,
      userMessageId: typeof json.user_message_id === 'string' ? json.user_message_id : undefined,
    };
    if (typeof json.error === 'string') {
      return {
        message: typeof json.message === 'string' ? json.message : json.error,
        code: typeof json.message === 'string' ? json.error : json.code,
        requestId: json.request_id,
        ...actionFields,
      };
    }
    if (json.error && typeof json.error.message === 'string') {
      return {
        message: json.error.message,
        code: json.error.code,
        requestId: json.error.request_id || json.request_id,
        ...actionFields,
      };
    }
    if (typeof json.message === 'string') {
      return {message: json.message, code: json.code, requestId: json.request_id, ...actionFields};
    }
  } catch {}
  return {message: text || '请求失败'};
}

export class ApiError extends Error {
  status: number;
  code?: string;
  requestId?: string;
  retryable?: boolean;
  userMessageId?: string;
  constructor(status: number, message: string, code?: string, requestId?: string, retryable?: boolean, userMessageId?: string) {
    super(message);
    Object.defineProperty(this, 'message', {value: message, enumerable: true, configurable: true});
    this.status = status;
    this.code = code;
    this.requestId = requestId;
    this.retryable = retryable;
    this.userMessageId = userMessageId;
    this.name = 'ApiError';
  }
}

function requestTimeoutForPath(path: string): number {
  if (
    path.includes('/api/v1/chat/') ||
    path.includes('/api/v1/experiences/rewrite')
  ) {
    return AI_REQUEST_TIMEOUT_MS;
  }
  return STANDARD_REQUEST_TIMEOUT_MS;
}

function isAbortError(error: unknown): boolean {
  return Boolean(error && typeof error === 'object' && (error as {name?: string}).name === 'AbortError');
}

function newRequestId(): string {
  const random = Math.random().toString(36).slice(2, 10);
  return `app-${Date.now().toString(36)}-${random}`;
}

function responseRequestId(res: Response): string | undefined {
  return res.headers?.get?.('X-Request-ID') || res.headers?.get?.('x-request-id') || undefined;
}

function apiErrorFromResponse(res: Response, text: string): ApiError {
  const error = parseErrorPayload(text);
  return new ApiError(
    res.status,
    error.message,
    error.code,
    error.requestId || responseRequestId(res),
    error.retryable,
    error.userMessageId,
  );
}

function withRequestIdHeader(init: Record<string, any>): Record<string, any> {
  return {
    ...init,
    headers: {
      ...(init.headers || {}),
      'X-Request-ID': init.headers?.['X-Request-ID'] || init.headers?.['x-request-id'] || newRequestId(),
    },
  };
}

export async function apiFetchWithTimeout(path: string, init: Record<string, any> = {}): Promise<Response> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), requestTimeoutForPath(path));

  try {
    const requestInit = withRequestIdHeader(init);
    return await fetch(`${API_BASE}${path}`, {
      ...requestInit,
      signal: controller.signal,
    });
  } catch (error) {
    if (isAbortError(error)) {
      throw new ApiError(0, '网络不稳，请稍后再试', 'request_timeout');
    }
    throw error;
  } finally {
    clearTimeout(timeout);
  }
}

// ---------- Go 后端 HTTP 请求 ----------
export async function apiGet(path: string): Promise<any> {
  const token = await getToken();
  const res = await apiFetchWithTimeout(path, {
    headers: token ? {Authorization: `Bearer ${token}`} : {},
  });
  const text = await res.text();
  if (!res.ok) {
    throw apiErrorFromResponse(res, text);
  }
  try { return JSON.parse(text); } catch { return text; }
}

export async function apiPost(path: string, body: any): Promise<any> {
  const token = await getToken();
  const res = await apiFetchWithTimeout(path, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? {Authorization: `Bearer ${token}`} : {}),
    },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  if (!res.ok) {
    throw apiErrorFromResponse(res, text);
  }
  try { return JSON.parse(text); } catch { return text; }
}

export async function apiPut(path: string, body: any): Promise<any> {
  const token = await getToken();
  const res = await apiFetchWithTimeout(path, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? {Authorization: `Bearer ${token}`} : {}),
    },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  if (!res.ok) {
    throw apiErrorFromResponse(res, text);
  }
  try { return JSON.parse(text); } catch { return text; }
}

export async function apiPatch(path: string, body: any): Promise<any> {
  const token = await getToken();
  const res = await apiFetchWithTimeout(path, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? {Authorization: `Bearer ${token}`} : {}),
    },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  if (!res.ok) {
    throw apiErrorFromResponse(res, text);
  }
  try { return JSON.parse(text); } catch { return text; }
}

export async function apiDelete(path: string): Promise<any> {
  const token = await getToken();
  const res = await apiFetchWithTimeout(path, {
    method: 'DELETE',
    headers: token ? {Authorization: `Bearer ${token}`} : {},
  });
  const text = await res.text();
  if (!res.ok) {
    throw apiErrorFromResponse(res, text);
  }
  try { return JSON.parse(text); } catch { return text; }
}
