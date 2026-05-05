import { clearToken } from './config';

export async function logout(): Promise<void> {
  await clearToken();
}

export async function isLoggedIn(): Promise<boolean> {
  const { getToken } = require('./config');
  const token = await getToken();
  return !!token;
}
