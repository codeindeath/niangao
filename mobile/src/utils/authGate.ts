import {Alert} from 'react-native';
import {getToken} from '../services/config';

export async function requireLogin(navigation: any, message: string): Promise<boolean> {
  const token = await getToken();
  if (token) return true;

  Alert.alert('先登录一下', message, [
    {text: '先看看', style: 'cancel'},
    {
      text: 'Apple登录',
      onPress: () => {
        const parent = navigation.getParent?.();
        if (parent?.navigate) {
          parent.navigate('login');
          return;
        }
        navigation.navigate('login');
      },
    },
  ]);
  return false;
}
