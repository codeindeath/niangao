import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  StyleSheet,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {fetchUserStats, UserStats, DomainCount} from '../services/api';
import {fetchProfile, UserProfile} from '../services/api';
import {logout} from '../services/auth';

const SUB_LABELS: Record<string, string> = {
  'career-planning': '职业规划', 'skill-building': '技能提升',
  'side-hustle': '副业创业', 'workplace-comm': '职场沟通',
  'intimate': '亲密关系', 'family': '家庭关系',
  'social-skill': '社交技巧', 'communication': '沟通表达',
  'mental-model': '思维模型', 'learning': '学习方法',
  'decision': '决策判断', 'psychology': '心理认知',
  'finance': '理财规划', 'health': '健康养生',
  'time-mgmt': '时间管理', 'habits': '习惯养成',
  'digital-life': '数字生活',
  'regulation': '情绪调节', 'self-growth': '自我成长',
  'happiness': '幸福感', 'stress-mgmt': '压力管理',
};

type PubTab = 'published' | 'liked_by_others' | 'bookmarked_by_others';
type InterTab = 'viewed' | 'liked' | 'bookmarked';

const BAR_COLORS: Record<string, string> = {
  published: '#3d6a4b', liked_by_others: '#6fa87c', bookmarked_by_others: '#e8a850',
  viewed: '#5c7aa8', liked: '#e85d5d', bookmarked: '#e8a850',
};
const BAR_GRADIENTS: Record<string, [string, string]> = {
  published: ['#3d6a4b', '#4a7c59'], liked_by_others: ['#6fa87c', '#8bc49a'],
  bookmarked_by_others: ['#e8a850', '#f0be70'],
  viewed: ['#5c7aa8', '#80a0c8'], liked: ['#e85d5d', '#f08080'],
  bookmarked: ['#e8a850', '#f0be70'],
};

function top5(dist: DomainCount[] | null | undefined): DomainCount[] {
  if (!dist || dist.length === 0) return [];
  const filtered = dist.filter(d => d.count > 0 && d.domain !== '');
  return filtered.length > 5 ? filtered.slice(0, 5) : filtered;
}

function domainLabel(domain: string): string {
  return SUB_LABELS[domain] || domain || '未分类';
}

export default function ProfileScreen({navigation}: any) {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [pubTab, setPubTab] = useState<PubTab>('published');
  const [interTab, setInterTab] = useState<InterTab>('viewed');

  const load = async () => {
    try {
      const [p, s] = await Promise.all([fetchProfile(), fetchUserStats()]);
      setProfile(p);
      setStats(s);
    } catch (e: any) {
      console.error('Failed to load profile/stats:', e);
      if (e?.status === 401) setProfile(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);
  useEffect(() => {
    const unsub = navigation.addListener('focus', () => load());
    return unsub;
  }, [navigation]);

  const handleLogout = async () => {
    await logout();
    setProfile(null);
    setStats(null);
  };

  if (loading) {
    return (
      <SafeAreaView style={st.container} edges={['top']}>
        <ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 200}} />
      </SafeAreaView>
    );
  }

  if (!profile) {
    return (
      <SafeAreaView style={st.container} edges={['top']}>
        <View style={{flex: 1, justifyContent: 'center', alignItems: 'center'}}>
          <TouchableOpacity style={st.loginBtn} onPress={() => navigation.navigate('login')}>
            <Text style={st.loginText}>去登录</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  const pubDists = stats?.published_dist;
  const interDists = stats?.interactions_dist;
  const pubItems = top5(pubDists?.[pubTab]);
  const interItems = top5(interDists?.[interTab]);
  const maxPubCount = pubItems.length > 0 ? pubItems[0].count : 1;
  const maxInterCount = interItems.length > 0 ? interItems[0].count : 1;

  return (
    <SafeAreaView style={st.container} edges={['top']}>
      <ScrollView style={st.scroll} bounces={false}>
        {/* ═══ Profile header ═══ */}
        <TouchableOpacity
          style={st.profileRow}
          onPress={() => navigation.navigate('profileEdit')}
          activeOpacity={0.7}>
          <View style={st.avatar}><Text style={st.avatarText}>{(profile.nickname || '?').charAt(0)}</Text></View>
          <View style={st.profileInfo}>
            <Text style={st.nickname}>{profile.nickname || '年糕用户'}</Text>
            <Text style={st.titleText}>{profile.title || ''}</Text>
          </View>
          <Text style={st.arrow}>›</Text>
        </TouchableOpacity>

        <View style={st.blockDivider} />

        {/* ═══ Block 1: 我的内容表现 ═══ */}
        <View style={st.block}>
          <View style={st.statsRow}>
            <View style={st.statCell}><Text style={st.statLabel}>我发布的</Text><Text style={st.statNum}>{stats?.published.count ?? 0}</Text></View>
            <View style={st.statCell}><Text style={st.statLabel}>获点赞</Text><Text style={st.statNum}>{stats?.published.liked_by_others ?? 0}</Text></View>
            <View style={st.statCell}><Text style={st.statLabel}>获收藏</Text><Text style={st.statNum}>{stats?.published.bookmarked_by_others ?? 0}</Text></View>
          </View>

          <View style={st.domainHeader}>
            <Text style={st.domainTitle}>领域分布</Text>
            <View style={st.toggle}>
              {(['published','liked_by_others','bookmarked_by_others'] as PubTab[]).map(t => (
                <TouchableOpacity key={t} onPress={() => setPubTab(t)} style={[st.toggleBtn, pubTab === t && st.toggleActive]}>
                  <Text style={[st.toggleText, pubTab === t && st.toggleTextActive]}>
                    {t === 'published' ? '发布' : t === 'liked_by_others' ? '获赞' : '被收藏'}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>

          {pubItems.length === 0 ? (
            <Text style={st.emptyDist}>暂无数据</Text>
          ) : pubItems.map(d => (
            <View key={d.domain} style={st.barItem}>
              <Text style={st.barLabel} numberOfLines={1}>{domainLabel(d.domain)}</Text>
              <View style={st.barTrack}>
                <View style={[st.barFill, {width: `${Math.max(5, (d.count / maxPubCount) * 100)}%`, backgroundColor: BAR_GRADIENTS[pubTab]?.[0] || '#4a7c59'}]}>
                  <View style={[st.barDot, {backgroundColor: BAR_COLORS[pubTab] || '#4a7c59'}]}>
                    <Text style={st.barDotText}>{d.count}</Text>
                  </View>
                </View>
              </View>
            </View>
          ))}
        </View>

        <View style={st.blockDivider} />

        {/* ═══ Block 2: 我的互动足迹 ═══ */}
        <View style={st.block}>
          <View style={st.statsRow}>
            <View style={st.statCell}><Text style={st.statLabel}>我看过的</Text><Text style={st.statNum}>{stats?.interactions.viewed ?? 0}</Text></View>
            <View style={st.statCell}><Text style={st.statLabel}>我点赞</Text><Text style={st.statNum}>{stats?.interactions.liked ?? 0}</Text></View>
            <View style={st.statCell}><Text style={st.statLabel}>我收藏</Text><Text style={st.statNum}>{stats?.interactions.bookmarked ?? 0}</Text></View>
          </View>

          <View style={st.domainHeader}>
            <Text style={st.domainTitle}>领域分布</Text>
            <View style={st.toggle}>
              {(['viewed','liked','bookmarked'] as InterTab[]).map(t => (
                <TouchableOpacity key={t} onPress={() => setInterTab(t)} style={[st.toggleBtn, interTab === t && st.toggleActive]}>
                  <Text style={[st.toggleText, interTab === t && st.toggleTextActive]}>
                    {t === 'viewed' ? '看过' : t === 'liked' ? '点赞' : '收藏'}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>

          {interItems.length === 0 ? (
            <Text style={st.emptyDist}>暂无数据</Text>
          ) : interItems.map(d => (
            <View key={d.domain} style={st.barItem}>
              <Text style={st.barLabel} numberOfLines={1}>{domainLabel(d.domain)}</Text>
              <View style={st.barTrack}>
                <View style={[st.barFill, {width: `${Math.max(5, (d.count / maxInterCount) * 100)}%`, backgroundColor: BAR_GRADIENTS[interTab]?.[0] || '#5c7aa8'}]}>
                  <View style={[st.barDot, {backgroundColor: BAR_COLORS[interTab] || '#5c7aa8'}]}>
                    <Text style={st.barDotText}>{d.count}</Text>
                  </View>
                </View>
              </View>
            </View>
          ))}
        </View>

        <View style={st.blockDivider} />

        {/* ═══ Chat ═══ */}
        <View style={st.block}>
          <Text style={st.chatTitle}>对话总数</Text>
          <View style={st.chatCards}>
            <View style={[st.chatCard, st.chatCard1]}>
              <Text style={[st.chatNum, st.chatNum1]}>{stats?.chat.conversations ?? 0}</Text>
              <Text style={st.chatUnit}>次对话</Text>
            </View>
            <View style={[st.chatCard, st.chatCard2]}>
              <Text style={[st.chatNum, st.chatNum2]}>{stats?.chat.messages ?? 0}</Text>
              <Text style={st.chatUnit}>条消息</Text>
            </View>
          </View>
        </View>

        <View style={st.blockDivider} />

        {/* ═══ Service list ═══ */}
        <TouchableOpacity style={st.listItem} onPress={() => navigation.navigate('placeholder', {title: '经验包'})} activeOpacity={0.6}>
          <Text style={st.listIcon}>📦</Text><Text style={st.listLabel}>经验包</Text><Text style={st.listArrow}>›</Text>
        </TouchableOpacity>
        <View style={st.divider} />
        <TouchableOpacity style={st.listItem} onPress={() => navigation.navigate('placeholder', {title: '对话人格'})} activeOpacity={0.6}>
          <Text style={st.listIcon}>💬</Text><Text style={st.listLabel}>对话人格</Text><Text style={st.listArrow}>›</Text>
        </TouchableOpacity>
        <View style={st.divider} />
        <TouchableOpacity style={st.listItem} onPress={() => navigation.navigate('placeholder', {title: '设置'})} activeOpacity={0.6}>
          <Text style={st.listIcon}>⚙</Text><Text style={st.listLabel}>设置</Text><Text style={st.listArrow}>›</Text>
        </TouchableOpacity>
        <View style={st.divider} />

        <TouchableOpacity style={st.logoutBtn} onPress={handleLogout}>
          <Text style={st.logoutText}>退出登录</Text>
        </TouchableOpacity>
      </ScrollView>
    </SafeAreaView>
  );
}

const st = StyleSheet.create({
  container: {flex: 1, backgroundColor: '#faf8f5'},
  scroll: {flex: 1},
  profileRow: {flexDirection: 'row', alignItems: 'center', paddingTop: 36, paddingBottom: 24, paddingHorizontal: 24},
  avatar: {width: 52, height: 52, borderRadius: 26, backgroundColor: '#4a7c59', justifyContent: 'center', alignItems: 'center', marginRight: 14},
  avatarText: {fontSize: 22, fontWeight: '700', color: '#fff'},
  profileInfo: {flex: 1},
  nickname: {fontSize: 18, fontWeight: '700', color: '#1a1a1a'},
  titleText: {fontSize: 13, color: '#9b9487', marginTop: 2},
  arrow: {fontSize: 22, color: '#c5bfb3', marginLeft: 8},
  block: {paddingHorizontal: 24, paddingBottom: 14},
  blockDivider: {height: 8, backgroundColor: '#f5f5f5'},
  statsRow: {flexDirection: 'row', marginBottom: 14},
  statCell: {flex: 1, alignItems: 'center', paddingVertical: 8},
  statLabel: {fontSize: 11, color: '#9b9487', marginBottom: 4},
  statNum: {fontSize: 22, fontWeight: '800', color: '#4a7c59'},
  domainHeader: {flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10},
  domainTitle: {fontSize: 11, color: '#b5b0a8', fontWeight: '500'},
  toggle: {flexDirection: 'row', backgroundColor: '#ece8df', borderRadius: 7, padding: 2},
  toggleBtn: {paddingHorizontal: 9, paddingVertical: 3, borderRadius: 6},
  toggleActive: {backgroundColor: '#fff', shadowColor: '#000', shadowOffset: {width: 0, height: 1}, shadowOpacity: 0.08, shadowRadius: 2, elevation: 1},
  toggleText: {fontSize: 10, fontWeight: '500', color: '#8b8274'},
  toggleTextActive: {color: '#4a7c59', fontWeight: '600'},
  barItem: {flexDirection: 'row', alignItems: 'center', marginBottom: 7},
  barLabel: {fontSize: 12, color: '#5c5548', width: 56, textAlign: 'right', marginRight: 10, fontWeight: '500'},
  barTrack: {flex: 1, height: 18, backgroundColor: '#ece8df', borderRadius: 9, overflow: 'visible'},
  barFill: {height: 18, borderRadius: 9, justifyContent: 'center'},
  barDot: {position: 'absolute', right: -2, width: 22, height: 22, borderRadius: 11, justifyContent: 'center', alignItems: 'center', shadowColor: '#000', shadowOffset: {width: 0, height: 2}, shadowOpacity: 0.15, shadowRadius: 6, elevation: 3},
  barDotText: {fontSize: 10, fontWeight: '800', color: '#fff'},
  emptyDist: {fontSize: 13, color: '#b5b0a8', textAlign: 'center', paddingVertical: 16},
  chatTitle: {fontSize: 11, color: '#b5b0a8', fontWeight: '500', marginBottom: 10},
  chatCards: {flexDirection: 'row', gap: 8},
  chatCard: {flex: 1, borderRadius: 12, paddingVertical: 14, alignItems: 'center'},
  chatCard1: {backgroundColor: '#eaf2e8'},
  chatCard2: {backgroundColor: '#e8f0fc'},
  chatNum: {fontSize: 24, fontWeight: '800'},
  chatNum1: {color: '#3d6a4b'},
  chatNum2: {color: '#4a6fa5'},
  chatUnit: {fontSize: 11, color: '#8b8274', marginTop: 2},
  divider: {height: 8, backgroundColor: '#f5f5f5'},
  listItem: {flexDirection: 'row', alignItems: 'center', paddingVertical: 14, paddingHorizontal: 24},
  listIcon: {fontSize: 16, marginRight: 12},
  listLabel: {flex: 1, fontSize: 15, color: '#2a2722'},
  listArrow: {fontSize: 16, color: '#c5bfb3'},
  logoutBtn: {marginHorizontal: 24, marginTop: 28, marginBottom: 20, paddingVertical: 14, alignItems: 'center'},
  logoutText: {fontSize: 16, color: '#c44', fontWeight: '500'},
  loginBtn: {backgroundColor: '#4a7c59', borderRadius: 20, paddingHorizontal: 36, paddingVertical: 14},
  loginText: {color: '#fff', fontSize: 16, fontWeight: '600'},
});
