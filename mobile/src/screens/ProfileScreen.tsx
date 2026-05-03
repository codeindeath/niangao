import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {getUserInfo, clearToken} from '../services/config';
import {fetchExperiences, Experience} from '../services/api';

export default function ProfileScreen({navigation}: any) {
  const [user, setUser] = useState<any>(null);
  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadProfile();
  }, []);

  const loadProfile = async () => {
    try {
      const userInfo = await getUserInfo();
      setUser(userInfo);

      // 加载经验（未登录时加载官方经验作为示例）
      const result = await fetchExperiences(1, undefined, 'latest');
      setExperiences(result.data || []);
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
  };

  const domainLabels: Record<string, string> = {
    career: '职场成长',
    relationship: '人际关系',
    cognition: '认知升级',
    life: '生活智慧',
    emotion: '情感',
  };

  const stats = {
    experiences: user?.experience_count || 5,
    bookmarks: user?.bookmark_count || 0,
    practiced: user?.practiced_count || 0,
  };

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <ScrollView contentContainerStyle={{paddingBottom: 40}}>
        {/* Profile header */}
        <View style={styles.profileHeader}>
          <View style={styles.avatarLarge}>
            <Text style={styles.avatarLargeText}>
              {user?.nickname?.charAt(0) || '?'}
            </Text>
          </View>
          <Text style={styles.nickname}>{user?.nickname || '未登录'}</Text>
          <Text style={styles.bio}>{user?.bio || '登录后查看个人主页'}</Text>

          {/* Stats */}
          <View style={styles.statsRow}>
            <View style={styles.statItem}>
              <Text style={styles.statNumber}>{stats.experiences}</Text>
              <Text style={styles.statLabel}>经验</Text>
            </View>
            <View style={styles.statDivider} />
            <View style={styles.statItem}>
              <Text style={styles.statNumber}>{stats.bookmarks}</Text>
              <Text style={styles.statLabel}>收藏</Text>
            </View>
            <View style={styles.statDivider} />
            <View style={styles.statItem}>
              <Text style={styles.statNumber}>{stats.practiced}</Text>
              <Text style={styles.statLabel}>实践</Text>
            </View>
          </View>
        </View>

        {/* Section: 经验列表 */}
        <View style={styles.sectionHeader}>
          <Text style={styles.sectionTitle}>全部经验</Text>
        </View>

        {loading ? (
          <ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 40}} />
        ) : (
          experiences.map(item => (
            <TouchableOpacity
              key={item.id}
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
          ))
        )}

        {/* Logout */}
        {user && (
          <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
            <Text style={styles.logoutText}>退出登录</Text>
          </TouchableOpacity>
        )}
        {!user && (
          <TouchableOpacity
            style={styles.loginButton}
            onPress={() => navigation.navigate('login')}>
            <Text style={styles.loginText}>登录</Text>
          </TouchableOpacity>
        )}
      </ScrollView>
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
    paddingBottom: 24,
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
  bio: {
    fontSize: 13,
    color: '#9a9a9a',
    marginBottom: 20,
  },
  statsRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statItem: {
    alignItems: 'center',
    paddingHorizontal: 28,
  },
  statNumber: {
    fontSize: 18,
    fontWeight: '700',
    color: '#1a1a1a',
  },
  statLabel: {
    fontSize: 11,
    color: '#9a9a9a',
    marginTop: 2,
  },
  statDivider: {
    width: 0.5,
    height: 24,
    backgroundColor: '#e8e4df',
  },
  sectionHeader: {
    paddingHorizontal: 18,
    paddingTop: 24,
    paddingBottom: 10,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: '700',
    color: '#6e6e6e',
    letterSpacing: 0.5,
  },
  card: {
    backgroundColor: '#ffffff',
    borderRadius: 16,
    padding: 16,
    marginHorizontal: 14,
    marginBottom: 8,
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
