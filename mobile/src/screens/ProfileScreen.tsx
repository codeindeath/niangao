import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {getUserInfo, clearToken} from '../services/config';
import {fetchMyExperiences, fetchMyBookmarks, Experience} from '../services/api';

type TabType = 'my' | 'bookmarks';

export default function ProfileScreen({navigation}: any) {
  const [user, setUser] = useState<any>(null);
  const [tab, setTab] = useState<TabType>('my');
  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [bookmarks, setBookmarks] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadProfile();
  }, []);

  const loadProfile = async () => {
    try {
      const userInfo = await getUserInfo();
      setUser(userInfo);

      if (userInfo) {
        const [myResult, bmResult] = await Promise.all([
          fetchMyExperiences(1),
          fetchMyBookmarks(1),
        ]);
        setExperiences(myResult.data || []);
        setBookmarks(bmResult.data || []);
      }
    } catch (e) {
      console.error('Failed to load profile:', e);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    await clearToken();
    setUser(null);
    setExperiences([]);
    setBookmarks([]);
  };

  const domainLabels: Record<string, string> = {
    career: '职场成长',
    relationship: '人际关系',
    cognition: '认知升级',
    life: '生活智慧',
    emotion: '情感',
  };

  const currentList = tab === 'my' ? experiences : bookmarks;

  const renderItem = ({item}: {item: Experience}) => (
    <TouchableOpacity
      style={styles.card}
      onPress={() => navigation.navigate('detail', {id: item.id})}
      activeOpacity={0.8}>
      <View style={styles.cardAuthorRow}>
        <View style={styles.cardAvatar}>
          <Text style={styles.cardAvatarText}>
            {item.author_name?.charAt(0) || '?'}
          </Text>
        </View>
        <Text style={styles.cardAuthorName}>{item.author_name || '匿名'}</Text>
        <View style={styles.cardDomainTag}>
          <Text style={styles.cardDomainText}>
            {domainLabels[item.domain] || item.domain}
          </Text>
        </View>
      </View>
      <Text style={styles.cardContent}>{item.content}</Text>
      <View style={styles.cardActions}>
        <Text style={styles.cardActionText}>♥ {item.like_count}</Text>
        <Text style={styles.cardActionText}>
          ★ {item.bookmark_count > 0 ? item.bookmark_count : '收藏'}
        </Text>
      </View>
    </TouchableOpacity>
  );

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <FlatList
        data={currentList}
        keyExtractor={item => item.id}
        renderItem={renderItem}
        contentContainerStyle={{paddingBottom: 80}}
        ListHeaderComponent={
          <>
            {/* Profile header */}
            <View style={styles.profileHeader}>
              <View style={styles.avatarLarge}>
                <Text style={styles.avatarLargeText}>
                  {user?.nickname?.charAt(0) || '?'}
                </Text>
              </View>
              <Text style={styles.nickname}>{user?.nickname || '未登录'}</Text>

              {/* Tabs */}
              <View style={styles.tabRow}>
                <TouchableOpacity
                  style={[styles.tab, tab === 'my' && styles.tabActive]}
                  onPress={() => setTab('my')}>
                  <Text style={[styles.tabText, tab === 'my' && styles.tabTextActive]}>
                    我的经验 ({experiences.length})
                  </Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={[styles.tab, tab === 'bookmarks' && styles.tabActive]}
                  onPress={() => setTab('bookmarks')}>
                  <Text style={[styles.tabText, tab === 'bookmarks' && styles.tabTextActive]}>
                    我的收藏 ({bookmarks.length})
                  </Text>
                </TouchableOpacity>
              </View>
            </View>

            {loading && (
              <ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 40}} />
            )}
          </>
        }
        ListFooterComponent={
          user ? (
            <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
              <Text style={styles.logoutText}>退出登录</Text>
            </TouchableOpacity>
          ) : (
            <TouchableOpacity
              style={styles.loginButton}
              onPress={() => navigation.navigate('login')}>
              <Text style={styles.loginText}>去登录</Text>
            </TouchableOpacity>
          )
        }
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#faf8f5',
  },
  profileHeader: {
    alignItems: 'center',
    paddingTop: 30,
    paddingBottom: 16,
    backgroundColor: '#ffffff',
    borderBottomWidth: 0.5,
    borderBottomColor: '#e8e4df',
  },
  avatarLarge: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: '#eaf2e8',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 12,
  },
  avatarLargeText: {
    fontSize: 24,
    fontWeight: '700',
    color: '#4a7c59',
  },
  nickname: {
    fontSize: 18,
    fontWeight: '700',
    color: '#1a1a1a',
    marginBottom: 16,
  },
  tabRow: {
    flexDirection: 'row',
    gap: 0,
    width: '100%',
  },
  tab: {
    flex: 1,
    paddingVertical: 10,
    alignItems: 'center',
    borderBottomWidth: 2,
    borderBottomColor: 'transparent',
  },
  tabActive: {
    borderBottomColor: '#4a7c59',
  },
  tabText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#9a9a9a',
  },
  tabTextActive: {
    color: '#4a7c59',
    fontWeight: '700',
  },
  card: {
    backgroundColor: '#ffffff',
    borderRadius: 16,
    padding: 16,
    marginHorizontal: 14,
    marginTop: 8,
    borderWidth: 0.5,
    borderColor: '#f0ece7',
  },
  cardAuthorRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  cardAvatar: {
    width: 22,
    height: 22,
    borderRadius: 11,
    backgroundColor: '#eaf2e8',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 6,
  },
  cardAvatarText: {
    fontSize: 11,
    fontWeight: '700',
    color: '#4a7c59',
  },
  cardAuthorName: {
    fontSize: 12,
    fontWeight: '500',
    color: '#6e6e6e',
    flex: 1,
  },
  cardDomainTag: {
    backgroundColor: '#eaf2e8',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 8,
  },
  cardDomainText: {
    fontSize: 10,
    fontWeight: '600',
    color: '#4a7c59',
  },
  cardContent: {
    fontSize: 15,
    lineHeight: 23,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 10,
  },
  cardActions: {
    flexDirection: 'row',
    gap: 16,
    paddingTop: 8,
    borderTopWidth: 0.5,
    borderTopColor: '#f0ece7',
  },
  cardActionText: {
    fontSize: 11,
    color: '#9a9a9a',
  },
  logoutButton: {
    marginHorizontal: 18,
    marginTop: 30,
    backgroundColor: '#ffffff',
    borderRadius: 14,
    paddingVertical: 14,
    alignItems: 'center',
    borderWidth: 0.5,
    borderColor: '#e8e4df',
  },
  logoutText: {
    fontSize: 15,
    color: '#e85d5d',
    fontWeight: '500',
  },
  loginButton: {
    marginHorizontal: 18,
    marginTop: 30,
    backgroundColor: '#4a7c59',
    borderRadius: 14,
    paddingVertical: 14,
    alignItems: 'center',
  },
  loginText: {
    fontSize: 16,
    color: '#ffffff',
    fontWeight: '600',
  },
});
