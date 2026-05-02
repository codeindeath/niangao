import { supabase } from './config';

export interface AuthResponse {
  user: any;
  session: any;
  error: Error | null;
}

export async function signInWithPhone(phone: string): Promise<AuthResponse> {
  const { data, error } = await supabase.auth.signInWithOtp({
    phone,
  });
  return { user: data.user, session: data.session, error };
}

export async function verifyOTP(phone: string, token: string): Promise<AuthResponse> {
  const { data, error } = await supabase.auth.verifyOtp({
    phone,
    token,
    type: 'sms',
  });
  return { user: data.user, session: data.session, error };
}

export async function signOut(): Promise<void> {
  await supabase.auth.signOut();
}

export function getCurrentSession() {
  return supabase.auth.getSession();
}

export function onAuthStateChange(callback: (session: any) => void) {
  return supabase.auth.onAuthStateChange((_event, session) => {
    callback(session);
  });
}
