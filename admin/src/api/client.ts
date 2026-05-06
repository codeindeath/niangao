const API_BASE = window.location.hostname === 'localhost'
  ? 'http://115.190.177.146'
  : '';

export async function apiGet(path: string) {
  const token = localStorage.getItem('admin_token');
  const res = await fetch(API_BASE + path, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!res.ok) throw { status: res.status, message: await res.text() };
  return res.json();
}

export async function apiPost(path: string, body?: unknown) {
  const token = localStorage.getItem('admin_token');
  const res = await fetch(API_BASE + path, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw { status: res.status, message: await res.text() };
  return res.json();
}

export async function apiPut(path: string, body?: unknown) {
  const token = localStorage.getItem('admin_token');
  const res = await fetch(API_BASE + path, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw { status: res.status, message: await res.text() };
  return res.json();
}

export async function apiDel(path: string) {
  const token = localStorage.getItem('admin_token');
  const res = await fetch(API_BASE + path, {
    method: 'DELETE',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!res.ok) throw { status: res.status, message: await res.text() };
  return res.json();
}
