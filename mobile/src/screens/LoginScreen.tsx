import React, { useRef, useState } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Alert,
  ImageBackground,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import * as AppleAuthentication from 'expo-apple-authentication';
import { appleLogin, devLogin } from '../services/api';
import { setToken, setRefreshToken, setUserInfo } from '../services/config';
import {userFacingErrorMessage} from '../utils/errors';

declare const __DEV__: boolean;
declare const process: {env?: Record<string, string | undefined>};

export default function LoginScreen({
  onLoginSuccess,
  onSkip,
}: {
  onLoginSuccess?: () => void;
  onSkip?: () => void;
}) {
  const [loading, setLoading] = useState(false);
  const loginInFlightRef = useRef(false);
  const navigation = useNavigation<any>();
  const showDevLogin = __DEV__ && process.env?.EXPO_PUBLIC_ENABLE_DEV_LOGIN === '1';

  const beginLogin = () => {
    if (loginInFlightRef.current) {
      return false;
    }
    loginInFlightRef.current = true;
    setLoading(true);
    return true;
  };

  const finishLogin = () => {
    loginInFlightRef.current = false;
    setLoading(false);
  };

  const handleSuccess = () => {
    if (onLoginSuccess) {
      onLoginSuccess();
    } else {
      // Inside authenticated stack — navigate back to main
      navigation.navigate('main');
    }
  };

  const handleAppleLogin = async () => {
    if (!beginLogin()) {
      return;
    }

    let shouldEnterApp = false;
    let failureAlert: {title: string; message: string} | undefined;

    try {
      const credential = await AppleAuthentication.signInAsync({
        requestedScopes: [
          AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
          AppleAuthentication.AppleAuthenticationScope.EMAIL,
        ],
      });

      if (!credential.identityToken) {
        failureAlert = {title: '登录失败', message: 'Apple登录凭证无效，请重试'};
      } else {
        const fullName = credential.fullName
          ? [credential.fullName.givenName, credential.fullName.familyName]
              .filter(Boolean)
              .join(' ')
          : undefined;

        const result = await appleLogin(credential.identityToken, fullName);

        await setToken(result.token);
        if (result.refresh_token) {
          await setRefreshToken(result.refresh_token);
        }
        await setUserInfo(result.user);
        shouldEnterApp = true;
      }
    } catch (e: any) {
      if (e.code !== 'ERR_CANCELED') {
        failureAlert = {title: '登录失败', message: userFacingErrorMessage(e)};
      }
    } finally {
      finishLogin();
    }

    if (failureAlert) {
      Alert.alert(failureAlert.title, failureAlert.message);
      return;
    }

    if (shouldEnterApp) {
      handleSuccess();
    }
  };

  const handleDevLogin = async () => {
    if (!beginLogin()) {
      return;
    }

    let shouldEnterApp = false;
    let failureAlert: {title: string; message: string} | undefined;

    try {
      const result = await devLogin();
      await setToken(result.token);
      if (result.refresh_token) {
        await setRefreshToken(result.refresh_token);
      }
      await setUserInfo(result.user);
      shouldEnterApp = true;
    } catch (e: any) {
      if (e?.status === 404) {
        failureAlert = {
          title: '开发登录不可用',
          message: '当前后端还没有启用开发登录接口，请切换到 V4 测试后端再试。',
        };
      } else {
        failureAlert = {title: '登录失败', message: userFacingErrorMessage(e)};
      }
    } finally {
      finishLogin();
    }

    if (failureAlert) {
      Alert.alert(failureAlert.title, failureAlert.message);
      return;
    }

    if (shouldEnterApp) {
      handleSuccess();
    }
  };

  return (
    <ImageBackground
      source={require('../../assets/niangao-login-bg.png')}
      style={styles.background}
      resizeMode="cover">
      <View style={styles.shade} />
      <SafeAreaView style={styles.container}>
        <View style={styles.brand}>
          <Text style={styles.logo}>年糕</Text>
          <Text style={styles.tagline}>生活有态度</Text>
        </View>

        <View style={styles.bottom}>
          <TouchableOpacity
            style={[styles.primaryButton, loading && styles.buttonDisabled]}
            onPress={handleAppleLogin}
            activeOpacity={0.82}
            disabled={loading}>
            <Text style={styles.primaryButtonText}>Apple登录</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.secondaryButton}
            onPress={onSkip || (() => navigation.navigate('main'))}
            activeOpacity={0.78}
            disabled={loading}>
            <Text style={styles.secondaryButtonText}>先看看</Text>
          </TouchableOpacity>

          {showDevLogin && (
            <TouchableOpacity
              style={styles.devButton}
              onPress={handleDevLogin}
              activeOpacity={0.6}
              disabled={loading}>
              <Text style={styles.devButtonText}>开发模拟登录</Text>
            </TouchableOpacity>
          )}

          {loading && (
            <ActivityIndicator style={{ marginTop: 12 }} color="#f7f0e6" />
          )}

          <Text style={styles.agreement}>
            登录即表示同意《用户协议》和《隐私政策》
          </Text>
        </View>
      </SafeAreaView>
    </ImageBackground>
  );
}

const styles = StyleSheet.create({
  background: {
    flex: 1,
    backgroundColor: '#182119',
  },
  shade: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(13,22,16,0.22)',
  },
  container: {
    flex: 1,
    justifyContent: 'space-between',
  },
  brand: {
    flex: 1,
    justifyContent: 'flex-start',
    alignItems: 'flex-end',
    paddingTop: 86,
    paddingRight: 28,
  },
  logo: {
    fontSize: 44,
    fontWeight: '900',
    color: '#f7f0e6',
    letterSpacing: 0,
    textShadowColor: 'rgba(0,0,0,0.24)',
    textShadowOffset: {width: 0, height: 2},
    textShadowRadius: 10,
  },
  tagline: {
    fontSize: 16,
    fontWeight: '700',
    color: '#e4d8c4',
    marginTop: 8,
    letterSpacing: 0,
  },
  bottom: {
    paddingHorizontal: 24,
    paddingBottom: 46,
    alignItems: 'center',
  },
  primaryButton: {
    width: '100%',
    minHeight: 52,
    borderRadius: 14,
    backgroundColor: '#f7f0e6',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 10,
  },
  primaryButtonText: {
    color: '#151914',
    fontSize: 16,
    fontWeight: '800',
  },
  secondaryButton: {
    width: '100%',
    minHeight: 50,
    borderRadius: 14,
    borderWidth: 1,
    borderColor: 'rgba(247,240,230,0.62)',
    backgroundColor: 'rgba(24,35,27,0.28)',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 12,
  },
  secondaryButtonText: {
    color: '#f7f0e6',
    fontSize: 15,
    fontWeight: '800',
  },
  buttonDisabled: {
    opacity: 0.68,
  },
  devButton: {
    width: '100%',
    height: 44,
    backgroundColor: 'rgba(247,240,230,0.13)',
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 10,
  },
  devButtonText: {
    color: '#e4d8c4',
    fontSize: 13,
    fontWeight: '700',
  },
  agreement: {
    fontSize: 12,
    color: 'rgba(247,240,230,0.62)',
    textAlign: 'center',
  },
});
