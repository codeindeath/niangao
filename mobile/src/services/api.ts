import {apiGet, apiPost, aiPost} from './config';

// ============================================================
// 类型
// ============================================================
export interface Experience {
  id: string;
  author_id: string;
  content: string;
  interpretation?: string;
  domain: string;
  is_official: boolean;
  source_label?: string;
  like_count: number;
  bookmark_count: number;
  author_name?: string;
  author_avatar?: string;
  is_liked: boolean;
  is_bookmarked: boolean;
  created_at: string;
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
  interpretation?: string,
): Promise<Experience> {
  return apiPost('/api/v1/experiences', {content, domain, interpretation});
}

export async function toggleLike(id: string): Promise<{liked: boolean}> {
  return apiPost(`/api/v1/experiences/${id}/like`, {});
}

export async function toggleBookmark(
  id: string,
): Promise<{bookmarked: boolean}> {
  return apiPost(`/api/v1/experiences/${id}/bookmark`, {});
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
