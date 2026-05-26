import {getToken} from '../services/config';

export type ProtectedMainTab = 'chat' | 'create';

function openLogin(navigation: any) {
  const parent = navigation.getParent?.();
  if (parent?.navigate) {
    parent.navigate('login');
    return;
  }
  navigation.navigate?.('login');
}

export async function openProtectedMainTab(
  event: {preventDefault?: () => void},
  navigation: any,
  routeName: ProtectedMainTab,
): Promise<void> {
  event.preventDefault?.();

  try {
    const token = await getToken();
    if (token) {
      navigation.navigate(routeName);
      return;
    }
  } catch {}

  openLogin(navigation);
}
