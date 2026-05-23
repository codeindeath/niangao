import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Text } from 'react-native';
import HomeScreen from '../screens/HomeScreen';
import ChatScreen from '../screens/ChatScreen';
import CreateScreen from '../screens/CreateScreen';
import ProfileScreen from '../screens/ProfileScreen';

const Tab = createBottomTabNavigator();

const TAB_BAR_VISIBLE = {
  backgroundColor: 'rgba(250,248,245,0.93)',
  borderTopColor: '#e8e4df',
  borderTopWidth: 0.5,
  height: 80,
  paddingBottom: 10,
  paddingTop: 6,
};

const TAB_BAR_HIDDEN = { display: 'none' as const };

function TabIcon({ name, focused }: { name: string; focused: boolean }) {
  const icons: Record<string, string> = {
    home: '⌂',
    chat: '💬',
    create: '📝',
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
        tabBarStyle:
          route.name === 'chat' || route.name === 'create'
            ? TAB_BAR_HIDDEN
            : TAB_BAR_VISIBLE,
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
        options={{ tabBarLabel: '记录' }}
      />
      <Tab.Screen
        name="profile"
        component={ProfileScreen}
        options={{ tabBarLabel: '我的' }}
      />
    </Tab.Navigator>
  );
}
