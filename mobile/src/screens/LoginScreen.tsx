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
import { loginWithWechat } from '../services/auth';

export default function LoginScreen({ onLoginSuccess }: { onLoginSuccess: () => void }) {
  const [loading, setLoading] = useState(false);

  const handleWechatLogin = async () => {
    setLoading(true);
    const result = await loginWithWechat();
    setLoading(false);

    if (result.success) {
      onLoginSuccess();
    } else if (result.error) {
      Alert.alert('登录失败', result.error);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.hero}>
        <Text style={styles.logo}>年糕</Text>
        <Text style={styles.tagline}>记录经验，年年成长</Text>
      </View>

      <View style={styles.bottom}>
        <TouchableOpacity
          style={styles.wechatButton}
          onPress={handleWechatLogin}
          disabled={loading}
          activeOpacity={0.8}
        >
          {loading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.wechatButtonText}>微信一键登录</Text>
          )}
        </TouchableOpacity>
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
  wechatButton: {
    width: '100%',
    height: 52,
    backgroundColor: '#4a7c59',
    borderRadius: 26,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  wechatButtonText: {
    color: '#fff',
    fontSize: 17,
    fontWeight: '600',
  },
  agreement: {
    fontSize: 12,
    color: '#9a9a9a',
    textAlign: 'center',
  },
});
