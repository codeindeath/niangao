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
import PlaceholderScreen from './src/screens/PlaceholderScreen';
import {getToken, clearToken, API_BASE} from './src/services/config';
import {isLoggedIn} from './src/services/auth';

const Stack = createNativeStackNavigator();

async function checkAndValidateToken(): Promise<boolean> {
  try {
    const token = await getToken();
    if (!token) return false;

    // Quick server-side validation
    const res = await fetch(`${API_BASE}/api/v1/user/profile`, {
      headers: {Authorization: `Bearer ${token}`},
    });
    if (res.status === 401) {
      await clearToken();
      return false;
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

  useEffect(() => {
    async function init() {
      try {
        const loggedIn = await checkAndValidateToken();
        setAuthenticated(loggedIn);
      } catch (e) {
        console.error('Failed to check login state:', e);
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
        {authenticated ? (
          <Stack.Navigator screenOptions={{headerShown: false}}>
            <Stack.Screen name="main" component={BottomTabNavigator} />
            <Stack.Screen name="detail" component={DetailScreen} />
            <Stack.Screen name="searchPage" component={SearchPage} options={{animation: 'slide_from_right'}} />
            <Stack.Screen name="searchCard" component={SearchCardScreen} options={{animation: 'slide_from_right'}} />
            <Stack.Screen name="profileEdit" component={ProfileEditScreen} options={{animation: 'slide_from_right'}} />
            <Stack.Screen name="placeholder" component={PlaceholderScreen} options={{animation: 'slide_from_right'}} />
            <Stack.Screen name="login" component={LoginScreen} />
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
