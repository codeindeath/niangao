import {apiGet, apiPost, apiPut, apiDelete, aiPost} from './config';

// ============================================================
// 类型
// ============================================================
export interface Experience {
  id: string;
  author_id: string;
  content: string;
  interpretation?: string;
  domain: string;
  sub_domain?: string;
  is_private?: boolean;
  is_official: boolean;
  source_type?: string;
  source_label?: string;
  creator_name?: string;
  score_reason?: string;
  like_count: number;
  bookmark_count: number;
  author_name?: string;
  author_avatar?: string;
  is_liked: boolean;
  is_bookmarked: boolean;
  review_status?: string;
  review_reason?: string;
  quality_score?: number;
  score_details?: string;
  created_at: string;
}

export interface UserProfile {
  id: string;
  nickname: string;
  avatar_url?: string;
  bio?: string;
  experience_count: number;
  bookmark_count: number;
  practiced_count: number;
}

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

// ============================================================
// 经验 API (Go 后端 :8080)
// ============================================================

export async function fetchExperiences(
  page: number = 1,
  domain?: string,
  sort: string = 'latest',
): Promise<{data: Experience[]; total: number; page: number}> {
  const params = new URLSearchParams({
    page: page.toString(),
    page_size: '20',
    sort,
  });
  if (domain) params.set('domain', domain);
  return apiGet(`/api/v1/experiences?${params}`);
}

export async function searchExperiences(
  keyword: string,
  page: number = 1,
): Promise<{data: Experience[]; total: number}> {
  const params = new URLSearchParams({
    search: keyword,
    page: page.toString(),
    page_size: '20',
  });
  return apiGet(`/api/v1/experiences?${params}`);
}

export async function fetchExperience(id: string): Promise<Experience> {
  return apiGet(`/api/v1/experiences/${id}`);
}

export async function createExperience(
  content: string,
  domain: string,
  sub_domain: string,
  is_private: boolean = false,
  interpretation?: string,
): Promise<Experience> {
  return apiPost('/api/v1/experiences', {
    content,
    domain,
    sub_domain,
    is_private,
    interpretation,
  });
}

export async function toggleLike(id: string): Promise<{liked: boolean}> {
  return apiPost(`/api/v1/experiences/${id}/like`, {});
}

export async function toggleBookmark(
  id: string,
): Promise<{bookmarked: boolean}> {
  return apiPost(`/api/v1/experiences/${id}/bookmark`, {});
}

export async function deleteExperience(id: string): Promise<{status: string}> {
  return apiDelete(`/api/v1/experiences/${id}`);
}

// ============================================================
// 用户个人 API
// ============================================================

export async function fetchMyExperiences(
  page: number = 1,
): Promise<{data: Experience[]; total: number; page: number}> {
  return apiGet(`/api/v1/me/experiences?page=${page}`);
}

export async function fetchMyBookmarks(
  page: number = 1,
): Promise<{data: Experience[]; total: number; page: number}> {
  return apiGet(`/api/v1/me/bookmarks?page=${page}`);
}

// ============================================================
// 用户 Profile API
// ============================================================

export async function fetchProfile(): Promise<UserProfile> {
  return apiGet('/api/v1/user/profile');
}

export async function updateProfile(fields: {
  nickname?: string;
  avatar_url?: string;
  bio?: string;
}): Promise<UserProfile> {
  return apiPut('/api/v1/user/profile', fields);
}

export async function deleteAccount(): Promise<{message: string}> {
  return apiDelete('/api/v1/user/account');
}

// ============================================================
// 推荐 API
// ============================================================

export async function fetchRecommendations(
  limit: number = 20,
  offset: number = 0,
): Promise<{data: Experience[]; total: number}> {
  return apiGet(`/api/v1/experiences/recommend?limit=${limit}&offset=${offset}`);
}

// ============================================================
// AI 对话 API (Python :8000)
// ============================================================

export async function sendMessage(
  message: string,
  userId: string,
  history: ChatMessage[] = [],
): Promise<{reply: string; referenced_experience_ids: string[]}> {
  return aiPost('/api/v1/chat/send', {
    message,
    user_id: userId,
    history,
  });
}

export async function generateInterpretation(
  content: string,
  domain: string,
): Promise<{interpretation: string}> {
  return aiPost('/api/v1/chat/generate-interpretation', {content, domain});
}

// ============================================================
// 登录 API
// ============================================================

export async function appleLogin(
  identityToken: string,
  fullName?: string,
): Promise<{token: string; refresh_token: string; user: any}> {
  return apiPost('/api/v1/auth/apple/login', {
    identity_token: identityToken,
    full_name: fullName,
  });
}

// 开发环境模拟登录
export async function devLogin(
  nickname?: string,
): Promise<{token: string; refresh_token: string; user: any}> {
  return apiPost('/api/v1/auth/dev/login', {
    nickname: nickname || '开发者',
  });
}

// 刷新 token
export async function refreshToken(
  refreshTokenValue: string,
): Promise<{token: string; refresh_token: string}> {
  return apiPost('/api/v1/auth/refresh', {
    refresh_token: refreshTokenValue,
  });
}

export {ApiError} from './config';
