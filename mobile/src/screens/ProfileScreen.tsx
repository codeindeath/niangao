import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {fetchMyExperiences, fetchMyBookmarks, fetchProfile, updateProfile, Experience, UserProfile} from '../services/api';
import {logout} from '../services/auth';

type TabType = 'my' | 'bookmarks';

export default function ProfileScreen({navigation}: any) {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [tab, setTab] = useState<TabType>('my');
  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [bookmarks, setBookmarks] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadProfile();
  }, []);

  // Re-fetch when tab is focused
  useEffect(() => {
    const unsubscribe = navigation.addListener('focus', () => {
      loadProfile();
    });
    return unsubscribe;
  }, [navigation]);

  const loadProfile = async () => {
    try {
      // Fetch fresh profile from API
      const [profileData, myResult, bmResult] = await Promise.all([
        fetchProfile(),
        fetchMyExperiences(1),
        fetchMyBookmarks(1),
      ]);
      setProfile(profileData);
      setExperiences(myResult.data || []);
      setBookmarks(bmResult.data || []);
    } catch (e: any) {
      console.error('Failed to load profile:', e);
      if (e?.status === 401) {
        // Token expired/invalid — navigate to login
        setProfile(null);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    setProfile(null);
    setExperiences([]);
    setBookmarks([]);
  };

  const handleEditNickname = () => {
    Alert.prompt(
      '修改昵称',
      '输入新昵称（1-30字）',
      [
        {text: '取消', style: 'cancel'},
        {
          text: '保存',
          onPress: async (text?: string) => {
            if (!text || !text.trim()) return;
            try {
              const updated = await updateProfile({nickname: text.trim()});
              setProfile(updated);
            } catch (e: any) {
              Alert.alert('修改失败', e?.message || '请稍后再试');
            }
          },
        },
      ],
      'plain-text',
      profile?.nickname || '',
    );
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
                  {profile?.nickname?.charAt(0) || '?'}
                </Text>
              </View>
              <Text style={styles.nickname}>{profile?.nickname || '未登录'}</Text>
              {profile && (
                <TouchableOpacity onPress={handleEditNickname} style={styles.editBtn}>
                  <Text style={styles.editBtnText}>✎ 修改</Text>
                </TouchableOpacity>
              )}

              {/* Stats */}
              {profile && (
                <View style={styles.statsRow}>
                  <View style={styles.stat}>
                    <Text style={styles.statNumber}>{profile.experience_count}</Text>
                    <Text style={styles.statLabel}>经验</Text>
                  </View>
                  <View style={styles.stat}>
                    <Text style={styles.statNumber}>{profile.bookmark_count}</Text>
                    <Text style={styles.statLabel}>收藏</Text>
                  </View>
                  <View style={styles.stat}>
                    <Text style={styles.statNumber}>{profile.practiced_count}</Text>
                    <Text style={styles.statLabel}>实践</Text>
                  </View>
                </View>
              )}

              {/* Bio */}
              {profile?.bio ? (
                <Text style={styles.bio}>{profile.bio}</Text>
              ) : null}

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
          profile ? (
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
    paddingHorizontal: 20,
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
    marginBottom: 4,
  },
  editBtn: {
    marginBottom: 8,
  },
  editBtnText: {
    fontSize: 13,
    color: '#4a7c59',
    fontWeight: '500',
  },
  bio: {
    fontSize: 14,
    color: '#6e6e6e',
    marginBottom: 12,
    textAlign: 'center',
    paddingHorizontal: 20,
  },
  statsRow: {
    flexDirection: 'row',
    gap: 24,
    marginBottom: 16,
  },
  stat: {
    alignItems: 'center',
  },
  statNumber: {
    fontSize: 18,
    fontWeight: '700',
    color: '#4a7c59',
  },
  statLabel: {
    fontSize: 12,
    color: '#8a8a8a',
    marginTop: 2,
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
    color: '#8a8a8a',
    fontWeight: '500',
  },
  tabTextActive: {
    color: '#4a7c59',
    fontWeight: '700',
  },
  card: {
    backgroundColor: '#fff',
    marginHorizontal: 12,
    marginTop: 8,
    borderRadius: 12,
    padding: 14,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 1,
  },
  cardAuthorRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  cardAvatar: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: '#eaf2e8',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 8,
  },
  cardAvatarText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#4a7c59',
  },
  cardAuthorName: {
    fontSize: 13,
    color: '#6e6e6e',
    flex: 1,
  },
  cardDomainTag: {
    backgroundColor: '#f0f4ef',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 10,
  },
  cardDomainText: {
    fontSize: 11,
    color: '#4a7c59',
  },
  cardContent: {
    fontSize: 15,
    color: '#1a1a1a',
    lineHeight: 22,
    marginBottom: 8,
  },
  cardActions: {
    flexDirection: 'row',
    gap: 16,
  },
  cardActionText: {
    fontSize: 13,
    color: '#8a8a8a',
  },
  logoutButton: {
    marginTop: 20,
    marginHorizontal: 12,
    backgroundColor: '#ffffff',
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#e8e4df',
  },
  logoutText: {
    fontSize: 15,
    color: '#c44',
    fontWeight: '600',
  },
  loginButton: {
    marginTop: 20,
    marginHorizontal: 12,
    backgroundColor: '#4a7c59',
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
  },
  loginText: {
    fontSize: 15,
    color: '#fff',
    fontWeight: '600',
  },
});
