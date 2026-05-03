import React, { useState } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import * as AppleAuthentication from 'expo-apple-authentication';
import { appleLogin } from '../services/api';
import { setToken, setUserInfo } from '../services/config';

export default function LoginScreen({ onLoginSuccess }: { onLoginSuccess: () => void }) {
  const [loading, setLoading] = useState(false);

  const handleAppleLogin = async () => {
    try {
      const credential = await AppleAuthentication.signInAsync({
        requestedScopes: [
          AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
          AppleAuthentication.AppleAuthenticationScope.EMAIL,
        ],
      });

      setLoading(true);

      const fullName = credential.fullName
        ? [credential.fullName.givenName, credential.fullName.familyName]
            .filter(Boolean)
            .join(' ')
        : undefined;

      const result = await appleLogin(credential.identityToken!, fullName);

      await setToken(result.token);
      await setUserInfo(result.user);
      setLoading(false);
      onLoginSuccess();
    } catch (e: any) {
      setLoading(false);
      if (e.code === 'ERR_CANCELED') {
        // 用户取消，不提示
        return;
      }
      Alert.alert('登录失败', e.message || 'Apple 登录出错');
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.hero}>
        <Text style={styles.logo}>年糕</Text>
        <Text style={styles.tagline}>记录经验，年年成长</Text>
      </View>

      <View style={styles.bottom}>
        {/* Apple Sign In */}
        <AppleAuthentication.AppleAuthenticationButton
          buttonType={AppleAuthentication.AppleAuthenticationButtonType.SIGN_IN}
          buttonStyle={AppleAuthentication.AppleAuthenticationButtonStyle.WHITE_OUTLINE}
          cornerRadius={26}
          style={styles.appleButton}
          onPress={handleAppleLogin}
        />

        {/* WeChat 保留，标灰提示不可用 */}
        <TouchableOpacity
          style={styles.wechatButton}
          disabled
          activeOpacity={0.6}
        >
          <Text style={styles.wechatButtonText}>微信登录（暂不可用）</Text>
        </TouchableOpacity>

        {loading && (
          <ActivityIndicator style={{ marginTop: 12 }} color="#4a7c59" />
        )}

        <Text style={styles.agreement}>
          登录即表示同意《用户协议》和《隐私政策》
        </Text>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#faf8f5',
    justifyContent: 'space-between',
  },
  hero: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  logo: {
    fontSize: 48,
    fontWeight: '700',
    color: '#4a7c59',
    marginBottom: 12,
  },
  tagline: {
    fontSize: 16,
    color: '#6e6e6e',
    letterSpacing: 2,
  },
  bottom: {
    paddingHorizontal: 32,
    paddingBottom: 60,
    alignItems: 'center',
  },
  appleButton: {
    width: '100%',
    height: 52,
    marginBottom: 12,
  },
  wechatButton: {
    width: '100%',
    height: 44,
    backgroundColor: '#d4d4d4',
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  wechatButtonText: {
    color: '#999',
    fontSize: 15,
    fontWeight: '600',
  },
  agreement: {
    fontSize: 12,
    color: '#9a9a9a',
    textAlign: 'center',
  },
});
