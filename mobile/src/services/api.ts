import {apiGet, apiPost, apiPut, apiPatch, apiDelete, ApiError} from './config';
import {reportHandledError} from '../utils/logging';

// ============================================================
// 类型
// ============================================================
export interface Experience {
  id: string;
  author_id: string;
  owner_user_id?: string;
  content: string;
  interpretation?: string;
  domain: string;
  sub_domain?: string;
  topics?: string;
  is_private?: boolean;
  is_official: boolean;
  experience_type?: string;
  visibility?: string;
  source_type?: string;
  source_label?: string;
  creator_name?: string;
  creator_display_name?: string;
  score_reason?: string;
  like_count: number;
  bookmark_count: number;
  author_name?: string;
  author_avatar?: string;
  author_title?: string;
  is_liked: boolean;
  is_bookmarked: boolean;
  review_status?: string;
  review_reason?: string;
  quality_score?: number;
  score_details?: string;
  original_text?: string;
  created_at: string;
}

export interface ExperienceCard {
  id: string;
  owner_user_id?: string;
  content?: string;
  experience_type?: 'platform_selected' | 'user_original' | string;
  visibility?: 'public' | 'private' | string;
  lifecycle_status?: string;
  domain?: string;
  sub_domain?: string;
  topic?: string;
  creator_display_name?: string;
  interpretation_status?: string;
  interpretation_summary_available?: boolean;
  quality_tier?: string;
  star_rating?: number;
  inspiration_count?: number;
  collection_count?: number;
  is_collected?: boolean;
  is_inspired?: boolean;
  unavailable_reason?: string;
}

export interface FeedPage {
  data: Experience[];
  next_cursor?: string;
  session_id?: string;
  has_more?: boolean;
  total: number;
}

function normalizeFeedCard(card: ExperienceCard): Experience {
  const starRating = card.star_rating ?? 0;
  return {
    id: card.id,
    author_id: card.owner_user_id || '',
    owner_user_id: card.owner_user_id,
    content: card.content || '',
    domain: card.domain || '',
    sub_domain: card.sub_domain || undefined,
    topics: card.topic || '',
    is_private: card.visibility === 'private',
    is_official: card.experience_type === 'platform_selected',
    source_type: card.experience_type === 'platform_selected' ? 'platform' : 'user',
    creator_name: card.creator_display_name || undefined,
    like_count: card.inspiration_count || 0,
    bookmark_count: card.collection_count || 0,
    author_name: card.creator_display_name || undefined,
    is_liked: Boolean(card.is_inspired),
    is_bookmarked: Boolean(card.is_collected),
    review_status: card.lifecycle_status,
    quality_score: starRating > 0 ? starRating * 2 : undefined,
    created_at: '',
  };
}

function normalizeFeedPage(page: {
  data?: ExperienceCard[];
  next_cursor?: string;
  session_id?: string;
  has_more?: boolean;
}, offset: number): FeedPage {
  const data = Array.isArray(page.data) ? page.data.map(normalizeFeedCard) : [];
  return {
    data,
    next_cursor: page.next_cursor,
    session_id: page.session_id,
    has_more: page.has_more,
    total: offset + data.length + (page.has_more ? 1 : 0),
  };
}

export interface UserProfile {
  id?: string;
  display_name?: string;
  nickname?: string;
  avatar_url?: string;
  bio?: string;
  title?: string;
  career_stage?: string;
  relationship_status?: string;
  is_parent?: boolean;
  common_issues?: string[];
  free_description?: string;
  profile_version?: number;
  experience_count?: number;
  bookmark_count?: number;
  practiced_count?: number;
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
  const offset = Math.max(page - 1, 0) * 20;
  const params = new URLSearchParams({
    q: keyword,
    limit: '20',
  });
  if (offset > 0) params.set('cursor', offset.toString());
  const result = await apiGet(`/api/v1/search/experiences?${params.toString()}`);
  return normalizeFeedPage(result, offset);
}

export async function fetchExperience(id: string): Promise<Experience> {
  return apiGet(`/api/v1/experiences/${id}`);
}

export async function createExperience(
  content: string,
  domain?: string,
  sub_domain?: string,
  is_private: boolean = false,
  interpretation?: string,
  topics?: string,
  options: {
    source_scene?: 'note' | 'chat';
    source_message_ids?: string[];
  } = {},
): Promise<Experience> {
  const body: Record<string, unknown> = {
    content,
    is_private,
    visibility: is_private ? 'private' : 'public',
    interpretation,
    topics,
    source_scene: options.source_scene || 'note',
  };
  if (options.source_message_ids) {
    body.source_message_ids = options.source_message_ids;
  }
  if (domain) body.domain = domain;
  if (sub_domain) body.sub_domain = sub_domain;
  const result = await apiPost('/api/v1/experiences', body);
  return result?.experience ?? result;
}

export async function updateExperience(
  id: string,
  content: string,
  domain?: string,
  sub_domain?: string,
  is_private: boolean = false,
  interpretation?: string,
  topics?: string,
): Promise<{status: string}> {
  const body: Record<string, unknown> = {
    content,
    is_private,
    visibility: is_private ? 'private' : 'public',
    interpretation,
    topics,
  };
  if (domain) body.domain = domain;
  if (sub_domain) body.sub_domain = sub_domain;
  return apiPut('/api/v1/experiences/' + id, body);
}

export async function markInspired(id: string): Promise<{inspired: boolean}> {
  try {
    await apiPost(`/api/v1/experiences/${id}/inspire`, {});
    return {inspired: true};
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      return {inspired: true};
    }
    throw err;
  }
}

export async function setCollected(
  id: string,
  collected: boolean = true,
): Promise<{collected: boolean}> {
  if (collected) {
    const result = await apiPost(`/api/v1/experiences/${id}/collect`, {});
    return {collected: Boolean(result?.collected ?? true)};
  }
  const result = await apiDelete(`/api/v1/experiences/${id}/collect`);
  return {collected: Boolean(result?.collected ?? false)};
}

export async function deleteExperience(id: string): Promise<{status: string}> {
  return apiDelete(`/api/v1/experiences/${id}`);
}

// ============================================================
// 用户个人 API
// ============================================================

export async function fetchMyExperiences(
  page: number = 1,
): Promise<{data: Experience[]; total: number; page: number; has_more?: boolean}> {
  const offset = Math.max(page - 1, 0) * 20;
  const result = await apiGet(`/api/v1/feed/mine?limit=20&cursor=${offset}`);
  const normalized = normalizeFeedPage(result, offset);
  return {...normalized, page};
}

export async function fetchMyBookmarks(
  page: number = 1,
): Promise<{data: Experience[]; total: number; page: number; has_more?: boolean}> {
  const offset = Math.max(page - 1, 0) * 20;
  const result = await apiGet(`/api/v1/feed/collections?limit=20&cursor=${offset}`);
  const normalized = normalizeFeedPage(result, offset);
  return {...normalized, page};
}

// ============================================================
// 用户 Profile API
// ============================================================

export async function fetchProfile(): Promise<UserProfile> {
  const profile = await apiGet('/api/v1/me/profile');
  return {
    ...profile,
    nickname: profile.display_name || profile.nickname || '',
    bio: profile.free_description || profile.bio || '',
  };
}

export async function updateProfile(fields: {
  nickname?: string;
  display_name?: string;
  avatar_url?: string;
  bio?: string;
  title?: string;
  career_stage?: string;
  relationship_status?: string;
  is_parent?: boolean;
  common_issues?: string[];
  free_description?: string;
}): Promise<UserProfile> {
  const patch: Record<string, unknown> = {};
  if (fields.display_name !== undefined || fields.nickname !== undefined) {
    patch.display_name = fields.display_name ?? fields.nickname;
  }
  if (fields.free_description !== undefined || fields.bio !== undefined) {
    patch.free_description = fields.free_description ?? fields.bio;
  }
  if (fields.career_stage !== undefined) patch.career_stage = fields.career_stage;
  if (fields.relationship_status !== undefined) patch.relationship_status = fields.relationship_status;
  if (fields.is_parent !== undefined) patch.is_parent = fields.is_parent;
  if (fields.common_issues !== undefined) patch.common_issues = fields.common_issues;
  const profile = await apiPatch('/api/v1/me/profile', patch);
  return {
    ...profile,
    nickname: profile.display_name || profile.nickname || '',
    bio: profile.free_description || profile.bio || '',
  };
}

export async function deleteAccount(): Promise<{message: string}> {
  return apiDelete('/api/v1/me/account');
}

export async function submitFeedback(fields: {
  type?: string;
  content: string;
  app_version?: string;
  device?: string;
  os_version?: string;
}): Promise<{status: string}> {
  return apiPost('/api/v1/me/feedback', fields);
}

// ============================================================
// 推荐 API
// ============================================================

export async function fetchRecommendations(
  limit: number = 20,
  offset: number = 0,
): Promise<{data: Experience[]; total: number; has_more?: boolean}> {
  const params = new URLSearchParams({limit: limit.toString()});
  if (offset > 0) params.set('cursor', offset.toString());
  const result = await apiGet(`/api/v1/feed/recommend?${params.toString()}`);
  return normalizeFeedPage(result, offset);
}

// ============================================================
// AI 对话 API (Go 后端编排，不再直连 Python)
// ============================================================

export interface ChatMessageItem {
  id: string;
  conversation_id?: string;
  topic_id?: string;
  temp_session_id?: string;
  role: 'user' | 'assistant';
  content: string;
  referenced_experience_ids?: string[];
  created_at: string;
}

export interface ChatTempSession {
  id: string;
  status: string;
  forced_new_topic: boolean;
  promoted_topic_id?: string;
}

export interface ChatTopic {
  id: string;
  status: string;
  title: string;
  domain?: string;
  sub_domain?: string;
  topic?: string;
  summary?: string;
}

export interface ChatReferenceCard {
  experience_id: string;
  content: string;
  is_collected: boolean;
  citation_sentence?: string;
}

export interface ChatNoteSuggestion {
  should_show: boolean;
  suggested_text?: string | null;
  source_message_ids: string[];
}

export interface SendChatMessageResult {
  user_message: ChatMessageItem;
  message: ChatMessageItem;
  reference_cards: ChatReferenceCard[];
  note_suggestion: ChatNoteSuggestion;
  retryable?: boolean;
}

export async function createChatTempSession(
  forcedNewTopic: boolean = false,
): Promise<ChatTempSession> {
  return apiPost('/api/v1/chat/temp-sessions', {forced_new_topic: forcedNewTopic});
}

export async function fetchRecentChatTopics(): Promise<{data: ChatTopic[]}> {
  return apiGet('/api/v1/chat/recent-topics');
}

export async function fetchChatTopicMessages(
  topicId: string,
): Promise<{data: ChatMessageItem[]; next_cursor?: string; has_more?: boolean}> {
  return apiGet(`/api/v1/chat/topics/${topicId}/messages`);
}

export async function sendChatTopicMessage(
  topicId: string,
  content: string,
  clientMessageId?: string,
): Promise<SendChatMessageResult> {
  return apiPost(`/api/v1/chat/topics/${topicId}/messages`, {
    content,
    client_message_id: clientMessageId,
  });
}

export async function sendTempChatMessage(
  tempSessionId: string,
  content: string,
  clientMessageId?: string,
): Promise<SendChatMessageResult> {
  return apiPost(`/api/v1/chat/temp-sessions/${tempSessionId}/messages`, {
    content,
    client_message_id: clientMessageId,
  });
}

export async function rewriteExperience(
  content: string,
  options: {
    source?: 'manual_note' | 'chat_note';
    default_visibility?: 'public' | 'private';
    user_selected_domain?: string;
    user_selected_sub_domain?: string;
    topic_context?: string;
    source_message_ids?: string[];
  } = {},
): Promise<{
  can_rewrite: boolean;
  rewritten_content: string;
  domain?: string;
  sub_domain?: string;
  topic?: string;
  retryable?: boolean;
}> {
  return apiPost('/api/v1/experiences/rewrite', {
    content,
    ...options,
  });
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

// ============================================================
// 统计
// ============================================================
export interface AssetStats {
  my_experiences: number;
  collections: number;
  month_added: number;
  public_experiences: number;
  private_experiences: number;
  from_note: number;
  from_chat: number;
}
export interface ContributionStats {
  inspired_users: number;
  collected_count: number;
  month_inspired_users: number;
  month_collected: number;
}
export interface ChangeStats {
  chat_topics: number;
  clearer_count: number;
  month_chat_experiences: number;
}
export type RecentHarvestRange = '7d' | '30d' | 'all';
export interface RecentHarvestStats {
  range: RecentHarvestRange;
  note_added: number;
  chat_experiences: number;
  inspired_users: number;
  collected_count: number;
}
export interface RespondedExperienceCard {
  id: string;
  content: string;
  domain: string;
  sub_domain?: string;
  star_rating: number;
  inspiration_count: number;
  collection_count: number;
  last_responded_at: string;
}

export async function fetchAssetStats(): Promise<AssetStats> {
  return apiGet('/api/v1/me/stats/assets');
}

export async function fetchContributionStats(): Promise<ContributionStats> {
  return apiGet('/api/v1/me/stats/contribution');
}

export async function fetchChangeStats(): Promise<ChangeStats> {
  return apiGet('/api/v1/me/stats/change');
}

export async function fetchRecentHarvestStats(
  range: RecentHarvestRange = '30d',
): Promise<RecentHarvestStats> {
  return apiGet(`/api/v1/me/stats/recent-harvest?range=${range}`);
}

export async function fetchRecentRespondedExperiences(
  limit: number = 3,
): Promise<{data: RespondedExperienceCard[]}> {
  return apiGet(`/api/v1/me/recent-responded-experiences?limit=${limit}`);
}

const _viewedIds = new Set<string>();

export function recordView(experienceId: string): void {
  if (_viewedIds.has(experienceId)) return;
  _viewedIds.add(experienceId);
  recordExperienceEvent(experienceId, 'expose', 'feed');
}

export function recordExperienceEvent(
  experienceId: string,
  eventType: 'expose' | 'flip' | 'search_click' | 'chat_citation_show' | 'chat_citation_click',
  sourceContext: string,
  metadata: Record<string, unknown> = {},
): void {
  apiPost(`/api/v1/experiences/${experienceId}/events`, {
    event_type: eventType,
    source_context: sourceContext,
    metadata,
  }).catch((err: any) => {
    reportHandledError(`recordExperienceEvent.${eventType}:${experienceId.slice(0, 8)}`, err);
  });
}

export function recordSearchClick(experienceId: string, keyword: string, rank: number): void {
  recordExperienceEvent(experienceId, 'search_click', 'search', {
    query: keyword,
    rank,
  });
}
