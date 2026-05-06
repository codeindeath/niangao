import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import HomeScreen from '../screens/HomeScreen';
import ChatScreen from '../screens/ChatScreen';
import CreateScreen from '../screens/CreateScreen';
import FriendsScreen from '../screens/FriendsScreen';
import ProfileScreen from '../screens/ProfileScreen';

const Tab = createBottomTabNavigator();

// 简化的 Tab 图标 — 生产环境替换为 SVG 图标库
function TabIcon({ name, focused }: { name: string; focused: boolean }) {
  const icons: Record<string, string> = {
    home: '⌂',
    chat: '💬',
    create: '+',
    search: '👥',
    profile: '👤',
  };
  return (
    <Text style={{ fontSize: focused ? 22 : 20, opacity: focused ? 1 : 0.5 }}>
      {icons[name] || '?'}
    </Text>
  );
}

export default function BottomTabNavigator() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarIcon: ({ focused }) => (
          <TabIcon name={route.name} focused={focused} />
        ),
        tabBarActiveTintColor: '#4a7c59',
        tabBarInactiveTintColor: '#9a9a9a',
        tabBarStyle: {
          backgroundColor: 'rgba(250,248,245,0.93)',
          borderTopColor: '#e8e4df',
          borderTopWidth: 0.5,
          height: 80,
          paddingBottom: 10,
          paddingTop: 6,
        },
        tabBarLabelStyle: {
          fontSize: 10,
          fontWeight: '600',
        },
      })}
    >
      <Tab.Screen
        name="home"
        component={HomeScreen}
        options={{ tabBarLabel: '首页' }}
      />
      <Tab.Screen
        name="chat"
        component={ChatScreen}
        options={{ tabBarLabel: '对话' }}
      />
      <Tab.Screen
        name="create"
        component={CreateScreen}
        options={{
          tabBarLabel: '',
          tabBarIcon: () => (
            <View style={styles.createButton}>
              <Text style={styles.createButtonText}>+</Text>
            </View>
          ),
        }}
      />
      <Tab.Screen
        name="search"
        component={FriendsScreen}
        options={{ tabBarLabel: '朋友' }}
      />
      <Tab.Screen
        name="profile"
        component={ProfileScreen}
        options={{ tabBarLabel: '我的' }}
      />
    </Tab.Navigator>
  );
}

const styles = StyleSheet.create({
  createButton: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: '#4a7c59',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 4,
    shadowColor: '#4a7c59',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.35,
    shadowRadius: 8,
    elevation: 6,
  },
  createButtonText: {
    color: 'white',
    fontSize: 24,
    fontWeight: '600',
    lineHeight: 28,
  },
});
