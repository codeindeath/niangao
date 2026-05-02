import { apiGet, apiPost } from './config';

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

export async function fetchExperiences(
  page: number = 1,
  domain?: string,
  sort: string = 'latest',
): Promise<{ data: Experience[]; total: number }> {
  const params = new URLSearchParams({ page: page.toString(), page_size: '20', sort });
  if (domain) params.set('domain', domain);
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
  return apiPost('/api/v1/experiences', { content, domain, interpretation });
}

export async function toggleLike(id: string): Promise<{ liked: boolean }> {
  return apiPost(`/api/v1/experiences/${id}/like`, {});
}

export async function toggleBookmark(id: string): Promise<{ bookmarked: boolean }> {
  return apiPost(`/api/v1/experiences/${id}/bookmark`, {});
}

export async function generateInterpretation(
  content: string,
  domain: string,
): Promise<{ interpretation: string }> {
  return apiPost('/api/v1/ai/experience/generate-interpretation', { content, domain });
}
