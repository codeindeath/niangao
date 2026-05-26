import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import Ionicons from '@expo/vector-icons/Ionicons';
import HomeScreen from '../screens/HomeScreen';
import ChatScreen from '../screens/ChatScreen';
import CreateScreen from '../screens/CreateScreen';
import ProfileScreen from '../screens/ProfileScreen';
import {openProtectedMainTab} from '../utils/protectedTab';

const Tab = createBottomTabNavigator();
type TabRouteName = 'home' | 'chat' | 'create' | 'profile';

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
  const icons: Record<TabRouteName, keyof typeof Ionicons.glyphMap> = {
    home: focused ? 'sparkles' : 'sparkles-outline',
    chat: focused ? 'chatbubble-ellipses' : 'chatbubble-ellipses-outline',
    create: focused ? 'add-circle' : 'add-circle-outline',
    profile: focused ? 'person-circle' : 'person-circle-outline',
  };

  const routeName = name as TabRouteName;
  const iconName = icons[routeName] ?? 'ellipse-outline';

  return (
    <Ionicons
      name={iconName}
      size={focused ? 23 : 21}
      color={focused ? '#4a7c59' : '#9a9a9a'}
    />
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
        options={{ tabBarLabel: '看看' }}
      />
      <Tab.Screen
        name="chat"
        component={ChatScreen}
        options={{ tabBarLabel: '聊聊' }}
        listeners={({navigation}) => ({
          tabPress: e => {
            void openProtectedMainTab(e, navigation, 'chat');
          },
        })}
      />
      <Tab.Screen
        name="create"
        component={CreateScreen}
        options={{ tabBarLabel: '记下' }}
        listeners={({navigation}) => ({
          tabPress: e => {
            void openProtectedMainTab(e, navigation, 'create');
          },
        })}
      />
      <Tab.Screen
        name="profile"
        component={ProfileScreen}
        options={{ tabBarLabel: '我的' }}
      />
    </Tab.Navigator>
  );
}
