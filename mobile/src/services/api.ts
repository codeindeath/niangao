import { API_BASE } from './config';
import { supabase } from './config';

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
  sort: string = 'latest'
): Promise<{ data: Experience[]; total: number }> {
  const session = await supabase.auth.getSession();
  const token = session.data.session?.access_token;

  const params = new URLSearchParams({
    page: page.toString(),
    page_size: '20',
    sort,
  });
  if (domain) params.set('domain', domain);

  const res = await fetch(`${API_BASE}/api/v1/experiences?${params}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  return res.json();
}

export async function fetchExperience(id: string): Promise<Experience> {
  const session = await supabase.auth.getSession();
  const token = session.data.session?.access_token;

  const res = await fetch(`${API_BASE}/api/v1/experiences/${id}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  return res.json();
}

export async function createExperience(
  content: string,
  domain: string,
  interpretation?: string
): Promise<Experience> {
  const session = await supabase.auth.getSession();
  const token = session.data.session?.access_token;

  const res = await fetch(`${API_BASE}/api/v1/experiences`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ content, domain, interpretation }),
  });
  return res.json();
}

export async function toggleLike(id: string): Promise<{ liked: boolean }> {
  const session = await supabase.auth.getSession();
  const token = session.data.session?.access_token;

  const res = await fetch(`${API_BASE}/api/v1/experiences/${id}/like`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.json();
}

export async function toggleBookmark(id: string): Promise<{ bookmarked: boolean }> {
  const session = await supabase.auth.getSession();
  const token = session.data.session?.access_token;

  const res = await fetch(`${API_BASE}/api/v1/experiences/${id}/bookmark`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.json();
}

export async function generateInterpretation(
  content: string,
  domain: string
): Promise<{ interpretation: string }> {
  const res = await fetch(`${API_BASE}/api/v1/ai/experience/generate-interpretation`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content, domain }),
  });
  return res.json();
}
