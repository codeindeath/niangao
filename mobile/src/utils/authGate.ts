import {Alert} from 'react-native';
import {clearToken, getToken} from '../services/config';

function navigateToLogin(navigation: any): void {
  const parent = navigation.getParent?.();
  if (parent?.navigate) {
    parent.navigate('login');
    return;
  }
  navigation.navigate('login');
}

export async function requireLogin(navigation: any, message: string): Promise<boolean> {
  const token = await getToken();
  if (token) return true;

  Alert.alert('先登录一下', message, [
    {text: '先看看', style: 'cancel'},
    {
      text: 'Apple登录',
      onPress: () => navigateToLogin(navigation),
    },
  ]);
  return false;
}

export function isAuthExpiredError(err: any): boolean {
  return err?.status === 401;
}

export async function handleAuthExpired(
  navigation: any,
  err: any,
  message: string = '重新登录后可以继续。',
): Promise<boolean> {
  if (!isAuthExpiredError(err)) return false;

  await clearToken();
  Alert.alert('登录状态过期', message, [
    {text: '先看看', style: 'cancel'},
    {text: 'Apple登录', onPress: () => navigateToLogin(navigation)},
  ]);
  return true;
}
