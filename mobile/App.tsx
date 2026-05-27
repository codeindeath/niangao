import React, {useEffect, useState} from 'react';
import {ActivityIndicator, View, StyleSheet} from 'react-native';
import {NavigationContainer} from '@react-navigation/native';
import {createNativeStackNavigator} from '@react-navigation/native-stack';
import {SafeAreaProvider} from 'react-native-safe-area-context';
import {StatusBar} from 'react-native';
import BottomTabNavigator from './src/navigation/BottomTabNavigator';
import DetailScreen from './src/screens/DetailScreen';
import LoginScreen from './src/screens/LoginScreen';
import SearchPage from './src/screens/SearchPage';
import SearchCardScreen from './src/screens/SearchCardScreen';
import ProfileEditScreen from './src/screens/ProfileEditScreen';
import CreateScreen from './src/screens/CreateScreen';
import {getToken, clearToken, apiFetchWithTimeout} from './src/services/config';
import {refreshToken as refreshAuthToken} from './src/services/auth';
import {reportHandledError} from './src/utils/logging';

const Stack = createNativeStackNavigator();

export async function checkAndValidateToken(): Promise<boolean> {
  try {
    const token = await getToken();
    if (!token) return false;

    // Quick server-side validation
    const res = await apiFetchWithTimeout('/api/v1/me/profile', {
      headers: {Authorization: `Bearer ${token}`},
    });
    if (res.status === 401) {
      if (await refreshAuthToken()) {
        return true;
      }
      await clearToken();
      return false;
    }
    if (!res.ok) {
      return true;
    }
    return res.ok;
  } catch {
    // Network error — trust local token for now
    const token = await getToken();
    return !!token;
  }
}

export default function App() {
  const [loading, setLoading] = useState(true);
  const [authenticated, setAuthenticated] = useState(false);
  const [guestMode, setGuestMode] = useState(false);

  useEffect(() => {
    async function init() {
      try {
        const loggedIn = await checkAndValidateToken();
        setAuthenticated(loggedIn);
      } catch (e) {
        reportHandledError('App.initAuth', e);
        setAuthenticated(false);
      }
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
        {authenticated || guestMode ? (
          <Stack.Navigator screenOptions={{headerShown: false}}>
            <Stack.Screen name="main" component={BottomTabNavigator} />
            <Stack.Screen name="detail" component={DetailScreen} />
            <Stack.Screen name="searchPage" component={SearchPage} options={{animation: 'slide_from_right'}} />
            <Stack.Screen name="searchCard" component={SearchCardScreen} options={{animation: 'slide_from_right'}} />
            <Stack.Screen name="createEdit" component={CreateScreen} options={{animation: 'slide_from_right'}} />
            <Stack.Screen name="profileEdit" component={ProfileEditScreen} options={{animation: 'slide_from_right'}} />
            <Stack.Screen name="login">
              {() => (
                <LoginScreen
                  onLoginSuccess={() => {
                    setAuthenticated(true);
                    setGuestMode(false);
                  }}
                />
              )}
            </Stack.Screen>
          </Stack.Navigator>
        ) : (
          <LoginScreen
            onLoginSuccess={() => {
              setAuthenticated(true);
              setGuestMode(false);
            }}
            onSkip={() => setGuestMode(true)}
          />
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
