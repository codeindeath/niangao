import React, {useEffect, useState} from 'react';
import {ActivityIndicator, View, StyleSheet} from 'react-native';
import {NavigationContainer} from '@react-navigation/native';
import {createNativeStackNavigator} from '@react-navigation/native-stack';
import {SafeAreaProvider} from 'react-native-safe-area-context';
import {StatusBar} from 'react-native';
import BottomTabNavigator from './navigation/BottomTabNavigator';
import DetailScreen from './screens/DetailScreen';
import LoginScreen from './screens/LoginScreen';
import {registerWechat, isLoggedIn} from './services/auth';

const Stack = createNativeStackNavigator();

export default function App() {
  const [loading, setLoading] = useState(true);
  const [authenticated, setAuthenticated] = useState(false);

  useEffect(() => {
    async function init() {
      // 微信 SDK 注册 — 开发阶段跳过可能的原生模块错误
      try {
        registerWechat();
      } catch (e) {
        console.log('WeChat SDK not available in dev:', e);
      }
      // 临时：跳过登录验证，直接进入主界面
      setAuthenticated(true);
      setLoading(false);
    }
    init();
  }, []);

  if (loading) {
    return (
      <View style={styles.loading}>
        <ActivityIndicator size="large" color="#4a7c59" />
      </View>
    );
  }

  return (
    <SafeAreaProvider>
      <StatusBar barStyle="dark-content" backgroundColor="#faf8f5" />
      <NavigationContainer>
        {authenticated ? (
          <Stack.Navigator screenOptions={{headerShown: false}}>
            <Stack.Screen name="main" component={BottomTabNavigator} />
            <Stack.Screen name="detail" component={DetailScreen} />
          </Stack.Navigator>
        ) : (
          <LoginScreen onLoginSuccess={() => setAuthenticated(true)} />
        )}
      </NavigationContainer>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  loading: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#faf8f5',
  },
});
