import { apiGet, apiPost, apiPut, apiDel } from './client';

export interface Dashboard {
  total_users: number;
  total_experiences: number;
  today_new_users: number;
  today_new_exps: number;
  today_active_users: number;
  today_ai_chats: number;
  pending_reviews: number;
  today_approved: number;
  today_rejected: number;
  yesterday_new_users?: number;
  yesterday_new_exps?: number;
  review_preview?: { id: string; content: string; domain: string; submitted_at: string }[];
}

export interface Trend {
  date: string;
  count: number;
}

export interface Trends {
  days: number;
  users: Trend[];
  experiences: Trend[];
}

export interface ReviewItem {
  id: string;
  content: string;
  domain: string;
  sub_domain?: string;
  source_type: string;
  review_status: string;
  ai_verdict?: string;
  ai_score?: number;
  ai_score_detail?: string;
  ai_interpretation?: string;
  hard_policy_result?: string;
  author_name: string;
  submitted_at: string;
}

export interface PaginatedData<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
}

export const DOMAIN_LABELS: Record<string, string> = {
  vitality: '生命',
  living: '生活',
  work: '工作',
  relationship: '关系',
  cognition: '认知',
  meaning: '意义',
};

// ── Auth ──

export async function adminLogin(username: string, password: string): Promise<{ token: string; user: unknown }> {
  const res = await fetch(
    (window.location.hostname === 'localhost' ? 'http://115.190.177.146' : '') + '/api/v1/auth/admin/login',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    }
  );
  if (!res.ok) throw { status: res.status };
  return res.json();
}

// ── Dashboard ──

export async function fetchDashboard(): Promise<Dashboard> {
  return apiGet('/api/v1/admin/dashboard');
}

export async function fetchTrends(days: number = 7): Promise<Trends> {
  return apiGet(`/api/v1/admin/trends?days=${days}`);
}

// ── AI Status ──

export interface AIStatus {
  model: string;
  healthy: boolean;
  latency_ms: number;
  success_rate: number;
  last_hour_calls: number;
  error_msg?: string;
  tier_stats?: {
    review: { today: number; total: number };
    chat: { today: number; total: number };
    interpretation: { today: number; total: number };
  };
  daily_cost?: { today_estimated: number; month_estimated: number };
  prompt_config?: { review_prompt_length: number; chat_system_prompt_length: number };
  batch_tasks?: { action_type: string; result: string; created_at: string }[];
}

export async function fetchAIStatus(): Promise<AIStatus> {
  return apiGet('/api/v1/admin/ai-status');
}

// ── Review Queue ──

export async function fetchReviews(params?: Record<string, string>): Promise<PaginatedData<ReviewItem>> {
  const qs = params ? '?' + new URLSearchParams(params).toString() : '';
  return apiGet('/api/v1/admin/reviews' + qs);
}

export async function approveReview(id: string) {
  return apiPost(`/api/v1/admin/reviews/${id}/approve`);
}

export async function rejectReview(id: string, reason: string) {
  return apiPost(`/api/v1/admin/reviews/${id}/reject`, { reason });
}

export async function batchReview(ids: string[], action: string, reason?: string) {
  return apiPost('/api/v1/admin/reviews/batch', { ids, action, reason });
}

export async function retryReview(id: string) {
  return apiPost(`/api/v1/admin/reviews/${id}/retry`);
}

export async function misjudgeReview(id: string, reason: string) {
  return apiPost(`/api/v1/admin/reviews/${id}/misjudge`, { reason });
}

// ── Content Management ──

export interface Experience {
  id: string;
  content: string;
  domain: string;
  sub_domain?: string;
  source_type: string;
  creator_name: string;
  review_status: string;
  score_reason?: string;
  likes: number;
  bookmarks: number;
  created_at: string;
}

export async function fetchExperiences(params?: Record<string, string>): Promise<PaginatedData<Experience>> {
  const qs = params ? '?' + new URLSearchParams(params).toString() : '';
  return apiGet('/api/v1/admin/experiences' + qs);
}

export async function updateExperience(id: string, body: Partial<Experience>) {
  return apiPut(`/api/v1/admin/experiences/${id}`, body);
}

export async function deleteExperience(id: string) {
  return apiDel(`/api/v1/admin/experiences/${id}`);
}

export async function unpublishExperience(id: string) {
  return apiPost(`/api/v1/admin/experiences/${id}/unpublish`);
}

export async function restoreExperience(id: string) {
  return apiPost(`/api/v1/admin/experiences/${id}/restore`);
}

export async function hardDeleteExperience(id: string) {
  return apiPost(`/api/v1/admin/experiences/${id}/hard-delete`);
}

export async function updateReviewStatus(id: string, review_status: string, reason?: string) {
  return apiPut(`/api/v1/admin/experiences/${id}/review-status`, { review_status, reason });
}

// ── User Management ──

export interface AdminUser {
  id: string;
  nickname: string;
  auth_provider: string;
  is_admin: boolean;
  is_active: boolean;
  exp_count: number;
  created_at: string;
  avatar_url?: string;
  title?: string;
  like_received?: number;
  bookmark_received?: number;
  viewed_count?: number;
  liked_count?: number;
  bookmarked_count?: number;
  chat_count?: number;
  msg_count?: number;
  bio?: string;
  domain_distribution?: Record<string, number>;
}

export async function fetchUsers(params?: Record<string, string>): Promise<PaginatedData<AdminUser>> {
  const qs = params ? '?' + new URLSearchParams(params).toString() : '';
  return apiGet('/api/v1/admin/users' + qs);
}

export async function fetchUserDetail(id: string): Promise<AdminUser> {
  return apiGet(`/api/v1/admin/users/${id}`);
}

export async function toggleUserEnabled(id: string, active: boolean) {
  return apiPut(`/api/v1/admin/users/${id}/status`, { active });
}

export async function batchUpdateUserStatus(ids: string[], active: boolean, reason?: string) {
  return apiPost('/api/v1/admin/users/batch-status', { ids, active, reason });
}

// ── Platform Content ──

export interface PlatformExperience {
  id: string;
  content: string;
  domain: string;
  sub_domain?: string;
  creator_name: string;
  source_label: string;
  quality_score?: number;
  has_interpretation: boolean;
  like_count: number;
  bookmark_count: number;
  score_reason?: string;
  created_at: string;
  author_name: string;
  status?: string;
}

export async function fetchPlatformExperiences(params?: Record<string, string>): Promise<PaginatedData<PlatformExperience>> {
  const qs = params ? '?' + new URLSearchParams(params).toString() : '';
  return apiGet('/api/v1/admin/platform-experiences' + qs);
}

export async function createPlatformExperience(body: Partial<PlatformExperience>) {
  return apiPost('/api/v1/admin/platform-experiences', body);
}

export async function updatePlatformExperience(id: string, body: Partial<PlatformExperience>) {
  return apiPut(`/api/v1/admin/platform-experiences/${id}`, body);
}

export async function togglePublishPlatformExperience(id: string) {
  return apiPost(`/api/v1/admin/platform-experiences/${id}/publish`);
}

export async function rescorePlatformExperience(id: string) {
  return apiPost(`/api/v1/admin/platform-experiences/${id}/rescore`);
}

export async function importCSVPlatformExperiences(csvData: string) {
  return apiPost('/api/v1/admin/platform-experiences/import-csv', { data: csvData });
}

export async function batchAIScore(ids: string[]) {
  return apiPost('/api/v1/admin/platform-experiences/batch-ai', { ids });
}

// ── Domain Management ──

export interface DomainItem {
  name: string;
  display_name: string;
  icon: string;
  parent_name?: string;
  exp_count: number;
  sort_order: number;
  active: boolean;
  sub_domains?: { name: string; display_name: string }[];
}

export async function fetchDomains(): Promise<{ domains: DomainItem[] }> {
  return apiGet('/api/v1/admin/domains');
}

export async function fetchDomainStats(): Promise<{ stats: { domain: string; count: number }[] }> {
  return apiGet('/api/v1/admin/domains/stats');
}

export async function updateDomain(name: string, body: { display_name: string; icon: string }) {
  return apiPut(`/api/v1/admin/domains/${name}`, body);
}

export async function toggleDomainActive(name: string, active: boolean) {
  return apiPost(`/api/v1/admin/domains/${name}/${active ? 'enable' : 'disable'}`);
}

export async function createDomain(body: { name: string; display_name: string; icon: string }) {
  return apiPost('/api/v1/admin/domains', body);
}

export async function addSubDomain(parentName: string, body: { name: string; display_name: string }) {
  return apiPost(`/api/v1/admin/domains/${parentName}/sub`, body);
}

export async function reorderDomains(body: { parent_name?: string; names: string[] }) {
  return apiPut('/api/v1/admin/domains/reorder', body);
}

// ── Statistics ──

export async function fetchUserStats(days: number = 7) {
  return apiGet(`/api/v1/admin/stats/users?days=${days}`);
}

export async function fetchExperienceStats(days: number = 7, sourceType?: string) {
  return apiGet(`/api/v1/admin/stats/experiences?days=${days}&source_type=${sourceType || 'all'}`);
}

export async function fetchInteractionStats(days: number = 7) {
  return apiGet(`/api/v1/admin/stats/interactions?days=${days}`);
}

export async function fetchReviewStats(days: number = 7) {
  return apiGet(`/api/v1/admin/stats/reviews?days=${days}`);
}

export async function fetchDomainDistribution() {
  return apiGet('/api/v1/admin/stats/domains');
}

export async function fetchAIStats(days: number = 7) {
  return apiGet(`/api/v1/admin/stats/ai?days=${days}`);
}

export async function fetchRetention(days: number = 30) {
  return apiGet(`/api/v1/admin/stats/retention?days=${days}`);
}

// ── System Config ──

export interface SystemConfig {
  [key: string]: unknown;
}

export async function fetchConfig(): Promise<SystemConfig> {
  return apiGet('/api/v1/admin/config');
}

export async function updateConfig(key: string, value: unknown) {
  return apiPut('/api/v1/admin/config', { key, value });
}

export async function fetchConfigDefaults() {
  return apiGet('/api/v1/admin/config/defaults');
}

export async function fetchSensitiveWords(): Promise<{ id: number; word: string }[]> {
  return apiGet('/api/v1/admin/config/sensitive-words');
}

export async function addSensitiveWord(word: string) {
  return apiPost('/api/v1/admin/config/sensitive-words', { word });
}

export async function deleteSensitiveWord(id: number) {
  return apiDel(`/api/v1/admin/config/sensitive-words/${id}`);
}

// ── Admin Logs ──

export interface AdminLog {
  id: string;
  admin_id: string;
  admin_name: string;
  action_type: string;
  target_type?: string;
  target_id?: string;
  detail?: string;
  result: string;
  created_at: string;
}

export async function fetchLogs(params?: Record<string, string>): Promise<PaginatedData<AdminLog>> {
  const qs = params ? '?' + new URLSearchParams(params).toString() : '';
  return apiGet('/api/v1/admin/logs' + qs);
}

// ── CSV Export ──

export function exportUsersCSV(): string {
  const token = localStorage.getItem('admin_token');
  const base = window.location.hostname === 'localhost' ? 'http://115.190.177.146' : '';
  return `${base}/api/v1/admin/export/users?format=csv&token=${token}`;
}

export function exportExperiencesCSV(): string {
  const token = localStorage.getItem('admin_token');
  const base = window.location.hostname === 'localhost' ? 'http://115.190.177.146' : '';
  return `${base}/api/v1/admin/export/experiences?format=csv&token=${token}`;
}
