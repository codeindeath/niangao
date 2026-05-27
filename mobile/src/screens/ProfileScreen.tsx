import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  StyleSheet,
  Alert,
  Modal,
  TextInput,
  Platform,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {
  fetchProfile,
  fetchAssetStats,
  fetchContributionStats,
  fetchChangeStats,
  fetchRecentHarvestStats,
  fetchRecentRespondedExperiences,
  UserProfile,
  AssetStats,
  ContributionStats,
  ChangeStats,
  RecentHarvestRange,
  RecentHarvestStats,
  RespondedExperienceCard,
  submitFeedback,
  deleteAccount,
} from '../services/api';
import {logout} from '../services/auth';
import AsyncStorage from '@react-native-async-storage/async-storage';
import {reportHandledError} from '../utils/logging';
import {handleAuthExpired} from '../utils/authGate';
import {userFacingErrorMessage} from '../utils/errors';
import Ionicons from '@expo/vector-icons/Ionicons';

const STATS_CACHE_KEY = 'niangao:v4:me-stats-cache';

const DOMAIN_LABELS: Record<string, string> = {
  vitality: '生命',
  living: '生活',
  work: '工作',
  relationship: '关系',
  cognition: '认知',
  meaning: '意义',
};
const SUB_LABELS: Record<string, string> = {
  health: '健康', housing: '居住', transit: '出行',
  diet: '饮食', exercise: '运动',
  self: '自我', happiness: '幸福', emotion: '情绪', faith: '信仰',
  mission: '使命', belonging: '归属',
  marriage: '夫妻', romance: '恋人', friendship: '朋友',
  parenting: '亲子', parents: '父母', siblings: '兄妹',
  jobhunt: '求职', promotion: '升职', startup: '创业',
  'work-comm': '沟通', management: '管理', productivity: '效率',
  'cog-learning': '学习', thinking: '思维', info: '信息',
  tools: '工具', creativity: '创造', expression: '表达',
};
const RANGE_LABELS: Array<{key: RecentHarvestRange; label: string}> = [
  {key: '7d', label: '近7天'},
  {key: '30d', label: '近30天'},
  {key: 'all', label: '全部'},
];

function clampPercent(value: number, total: number): `${number}%` {
  if (total <= 0 || value <= 0) return '0%';
  return `${Math.max(8, Math.round((value / total) * 100))}%` as `${number}%`;
}

function taxonomyLabel(card: RespondedExperienceCard): string {
  const domain = DOMAIN_LABELS[card.domain] || card.domain || '未分类';
  const sub = card.sub_domain ? SUB_LABELS[card.sub_domain] || card.sub_domain : '';
  return sub ? `${domain} · ${sub}` : domain;
}

function statText(value?: number | null): string {
  return value == null ? '—' : String(value);
}

type ProfileStatsCache = {
  assets?: AssetStats;
  contribution?: ContributionStats;
  change?: ChangeStats;
  recentHarvestByRange?: Partial<Record<RecentHarvestRange, RecentHarvestStats>>;
  respondedCards?: RespondedExperienceCard[];
};

async function readStatsCache(): Promise<ProfileStatsCache> {
  try {
    const raw = await AsyncStorage.getItem(STATS_CACHE_KEY);
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
}

function writeStatsCache(cache: ProfileStatsCache): void {
  AsyncStorage.setItem(STATS_CACHE_KEY, JSON.stringify(cache)).catch(() => {});
}

export default function ProfileScreen({navigation}: any) {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [assets, setAssets] = useState<AssetStats | null>(null);
  const [contribution, setContribution] = useState<ContributionStats | null>(null);
  const [change, setChange] = useState<ChangeStats | null>(null);
  const [recentRange, setRecentRange] = useState<RecentHarvestRange>('30d');
  const [recentHarvest, setRecentHarvest] = useState<RecentHarvestStats | null>(null);
  const [respondedCards, setRespondedCards] = useState<RespondedExperienceCard[]>([]);
  const [loading, setLoading] = useState(true);
  const [statsError, setStatsError] = useState(false);
  const [feedbackVisible, setFeedbackVisible] = useState(false);
  const [feedbackContent, setFeedbackContent] = useState('');
  const [feedbackSending, setFeedbackSending] = useState(false);

  const load = async () => {
    try {
      const p = await fetchProfile();
      setProfile(p);
      const cached = await readStatsCache();
      const nextCache: ProfileStatsCache = {
        ...cached,
        recentHarvestByRange: {...(cached.recentHarvestByRange || {})},
      };
      let shouldWriteCache = false;

      const [a, c, ch, rh, responded] = await Promise.allSettled([
        fetchAssetStats(),
        fetchContributionStats(),
        fetchChangeStats(),
        fetchRecentHarvestStats(recentRange),
        fetchRecentRespondedExperiences(3),
      ]);

      let hasStatsError = false;
      if (a.status === 'fulfilled') {
        setAssets(a.value);
        nextCache.assets = a.value;
        shouldWriteCache = true;
      } else {
        hasStatsError = true;
        if (cached.assets) setAssets(cached.assets);
      }
      if (c.status === 'fulfilled') {
        setContribution(c.value);
        nextCache.contribution = c.value;
        shouldWriteCache = true;
      } else {
        hasStatsError = true;
        if (cached.contribution) setContribution(cached.contribution);
      }
      if (ch.status === 'fulfilled') {
        setChange(ch.value);
        nextCache.change = ch.value;
        shouldWriteCache = true;
      } else {
        hasStatsError = true;
        if (cached.change) setChange(cached.change);
      }
      if (rh.status === 'fulfilled') {
        setRecentHarvest(rh.value);
        nextCache.recentHarvestByRange = {
          ...(nextCache.recentHarvestByRange || {}),
          [recentRange]: rh.value,
        };
        shouldWriteCache = true;
      } else {
        hasStatsError = true;
        const cachedRecent = cached.recentHarvestByRange?.[recentRange];
        if (cachedRecent) setRecentHarvest(cachedRecent);
      }
      if (responded.status === 'fulfilled') {
        const cards = responded.value.data || [];
        setRespondedCards(cards);
        nextCache.respondedCards = cards;
        shouldWriteCache = true;
      } else {
        hasStatsError = true;
        if (cached.respondedCards) setRespondedCards(cached.respondedCards);
      }
      setStatsError(hasStatsError);
      if (shouldWriteCache) writeStatsCache(nextCache);
    } catch (e: any) {
      if (e?.status === 401) {
        await logout();
        setProfile(null);
        setAssets(null);
        setContribution(null);
        setChange(null);
        setRecentHarvest(null);
        setRespondedCards([]);
        setStatsError(false);
      } else {
        reportHandledError('ProfileScreen.load', e);
        setStatsError(true);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [recentRange]);
  useEffect(() => {
    const unsub = navigation.addListener('focus', () => load());
    return unsub;
  }, [navigation, recentRange]);

  const clearLocalProfileState = () => {
    setProfile(null);
    setAssets(null);
    setContribution(null);
    setChange(null);
    setRecentHarvest(null);
    setRespondedCards([]);
    setStatsError(false);
    AsyncStorage.removeItem(STATS_CACHE_KEY).catch(() => {});
  };

  const performLogout = async () => {
    await logout();
    clearLocalProfileState();
  };

  const confirmLogout = () => {
    Alert.alert('退出登录？', '退出后，本机不再保留登录状态。你仍然可以继续浏览公开经验。', [
      {text: '取消', style: 'cancel'},
      {text: '退出登录', style: 'destructive', onPress: performLogout},
    ]);
  };

  const confirmDeleteAccount = () => {
    Alert.alert('注销账号？', '注销后会清除账号登录状态，这个操作需要再次确认。', [
      {text: '取消', style: 'cancel'},
      {
        text: '继续',
        style: 'destructive',
        onPress: () => {
          Alert.alert('确认注销账号', '注销账号后，账号数据将按服务端规则删除或匿名化。确定继续吗？', [
            {text: '取消', style: 'cancel'},
            {
              text: '确认注销',
              style: 'destructive',
              onPress: async () => {
                try {
                  await deleteAccount();
                  await logout();
                  clearLocalProfileState();
                } catch (e: any) {
                  if (await handleAuthExpired(navigation, e)) {
                    clearLocalProfileState();
                    return;
                  }
                  Alert.alert('注销失败', userFacingErrorMessage(e));
                }
              },
            },
          ]);
        },
      },
    ]);
  };

  const handleSubmitFeedback = async () => {
    const content = feedbackContent.trim();
    if (!content) {
      Alert.alert('提示', '请先写下反馈内容');
      return;
    }
    try {
      setFeedbackSending(true);
      await submitFeedback({
        type: 'general',
        content,
        app_version: '0.1.0',
        device: Platform.OS,
        os_version: String(Platform.Version),
      });
      setFeedbackContent('');
      setFeedbackVisible(false);
      Alert.alert('已收到', '谢谢你的反馈。');
    } catch (e: any) {
      if (await handleAuthExpired(navigation, e)) {
        clearLocalProfileState();
        setFeedbackVisible(false);
        return;
      }
      Alert.alert('提交失败', userFacingErrorMessage(e));
    } finally {
      setFeedbackSending(false);
    }
  };

  if (loading) {
    return (
      <SafeAreaView style={st.container} edges={['top']}>
        <ActivityIndicator size="large" color="#375943" style={{marginTop: 200}} />
      </SafeAreaView>
    );
  }

  if (!profile) {
    return (
      <SafeAreaView style={st.container} edges={['top']}>
        <View style={st.center}>
          <TouchableOpacity style={st.loginBtn} onPress={() => navigation.navigate('login')}>
            <Text style={st.loginText}>Apple登录</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  const displayName = profile.display_name || profile.nickname || '年糕用户';
  const initial = displayName.trim().charAt(0) || '年';
  const totalOwned = assets?.my_experiences ?? 0;
  const publicCount = assets?.public_experiences ?? 0;
  const privateCount = assets?.private_experiences ?? 0;
  const noteCount = assets?.from_note ?? 0;
  const chatCount = assets?.from_chat ?? 0;
  const helpedCount = contribution?.inspired_users ?? 0;
  const collectedCount = contribution?.collected_count ?? 0;

  return (
    <SafeAreaView style={st.container} edges={['top']}>
      <ScrollView style={st.scroll} contentContainerStyle={st.scrollContent} bounces={false}>
        <TouchableOpacity
          style={st.profileRow}
          onPress={() => navigation.navigate('profileEdit')}
          activeOpacity={0.72}>
          <View style={st.avatar}><Text style={st.avatarText}>{initial}</Text></View>
          <View style={st.profileInfo}>
            <Text style={st.name}>{displayName}</Text>
            <Text style={st.subtitle} numberOfLines={1}>
              {profile.free_description || '生活有态度'}
            </Text>
          </View>
          <Ionicons name="chevron-forward" size={18} color="#b9ac9c" />
        </TouchableOpacity>

        {statsError && (
          <View style={st.inlineNotice}>
            <Text style={st.inlineNoticeText}>部分信息暂时没取到</Text>
            <TouchableOpacity onPress={load} activeOpacity={0.72} style={st.inlineRetry}>
              <Text style={st.inlineRetryText}>重试</Text>
            </TouchableOpacity>
          </View>
        )}

        <View style={st.recentPanel}>
          <View style={st.sectionHead}>
            <Text style={st.sectionTitle}>最近收获</Text>
            <View style={st.rangeTabs}>
              {RANGE_LABELS.map(item => (
                <TouchableOpacity
                  key={item.key}
                  style={[st.rangeTab, recentRange === item.key && st.rangeTabActive]}
                  onPress={() => setRecentRange(item.key)}
                  activeOpacity={0.72}>
                  <Text style={[st.rangeText, recentRange === item.key && st.rangeTextActive]}>
                    {item.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>
          <View style={st.orbit}>
            <View style={[st.orbitNode, st.orbitPrimary]}>
              <Text style={st.orbitValue}>{statText(recentHarvest?.note_added)}</Text>
              <Text style={st.orbitLabel}>新记下</Text>
            </View>
            <View style={st.orbitSide}>
              <View style={st.orbitMini}>
                <Text style={st.orbitMiniValue}>{statText(recentHarvest?.chat_experiences)}</Text>
                <Text style={st.orbitMiniLabel}>聊出经验</Text>
              </View>
              <View style={st.orbitMini}>
                <Text style={st.orbitMiniValue}>{statText(recentHarvest?.inspired_users)}</Text>
                <Text style={st.orbitMiniLabel}>给人启发</Text>
              </View>
              <View style={st.orbitMini}>
                <Text style={st.orbitMiniValue}>{statText(recentHarvest?.collected_count)}</Text>
                <Text style={st.orbitMiniLabel}>被收藏</Text>
              </View>
            </View>
          </View>
        </View>

        <View style={st.section}>
          <View style={st.sectionHead}>
            <Text style={st.sectionTitle}>最近有回应的经验</Text>
            <Text style={st.sectionHint}>原创经验</Text>
          </View>
          {respondedCards.length > 0 ? (
            respondedCards.map(card => (
              <View key={card.id} style={st.respondedItem}>
                <Text style={st.respondedContent} numberOfLines={2}>{card.content}</Text>
                <View style={st.respondedMeta}>
                  <Text style={st.starLine}>{'★'.repeat(card.star_rating || 0)}</Text>
                  <Text style={st.taxonomy}>{taxonomyLabel(card)}</Text>
                </View>
                <Text style={st.responseNote}>
                  {card.inspiration_count} 人觉得有启发 · {card.collection_count} 人收藏
                </Text>
              </View>
            ))
          ) : (
            <View style={st.responseEmpty}>
              <Text style={st.responseValue}>
                {statText(contribution ? helpedCount + collectedCount : null)}
              </Text>
              <Text style={st.responseLabel}>公开原创有回应后，会在这里出现具体经验</Text>
            </View>
          )}
        </View>

        <View style={st.section}>
          <View style={st.sectionHead}>
            <Text style={st.sectionTitle}>我的积累</Text>
            <Text style={st.sectionHint}>{statText(assets?.collections)} 条收藏</Text>
          </View>
          <View style={st.totalLine}>
            <Text style={st.totalNum}>{statText(assets?.my_experiences)}</Text>
            <Text style={st.totalText}>条自己创建的经验</Text>
          </View>
          <View style={st.stackBar}>
            <View style={[st.stackFill, st.stackPublic, {width: clampPercent(publicCount, totalOwned)}]} />
            <View style={[st.stackFill, st.stackPrivate, {width: clampPercent(privateCount, totalOwned)}]} />
          </View>
          <View style={st.legendRow}>
            <Text style={st.legendText}>公开 {statText(assets?.public_experiences)}</Text>
            <Text style={st.legendText}>私密 {statText(assets?.private_experiences)}</Text>
          </View>
          <View style={st.sourceRow}>
            <View style={st.sourcePill}>
              <Text style={st.sourceValue}>{statText(assets?.from_note)}</Text>
              <Text style={st.sourceLabel}>记下</Text>
            </View>
            <View style={st.sourcePill}>
              <Text style={st.sourceValue}>{statText(assets?.from_chat)}</Text>
              <Text style={st.sourceLabel}>聊聊</Text>
            </View>
            <View style={st.sourcePill}>
              <Text style={st.sourceValue}>{statText(change?.chat_topics)}</Text>
              <Text style={st.sourceLabel}>议题</Text>
            </View>
          </View>
        </View>

        <View style={st.menu}>
          <TouchableOpacity
            style={st.menuItem}
            onPress={() => setFeedbackVisible(true)}
            activeOpacity={0.65}
            accessibilityRole="button"
            accessibilityLabel="意见反馈">
            <View style={st.menuIconWrap}>
              <Ionicons name="chatbox-ellipses-outline" size={18} color="#375943" />
            </View>
            <Text style={st.menuLabel}>意见反馈</Text>
            <Ionicons name="chevron-forward" size={17} color="#b9ac9c" />
          </TouchableOpacity>
          <TouchableOpacity
            style={st.menuItem}
            onPress={confirmLogout}
            activeOpacity={0.65}
            accessibilityRole="button"
            accessibilityLabel="退出登录">
            <View style={st.menuIconWrap}>
              <Ionicons name="log-out-outline" size={18} color="#375943" />
            </View>
            <Text style={st.menuLabel}>退出登录</Text>
            <Ionicons name="chevron-forward" size={17} color="#b9ac9c" />
          </TouchableOpacity>
          <TouchableOpacity
            style={st.menuItem}
            onPress={confirmDeleteAccount}
            activeOpacity={0.65}
            accessibilityRole="button"
            accessibilityLabel="注销账号">
            <View style={st.menuIconWrap}>
              <Ionicons name="trash-outline" size={18} color="#a5483d" />
            </View>
            <Text style={[st.menuLabel, st.dangerText]}>注销账号</Text>
            <Ionicons name="chevron-forward" size={17} color="#b9ac9c" />
          </TouchableOpacity>
        </View>
      </ScrollView>

      <Modal visible={feedbackVisible} transparent animationType="fade" onRequestClose={() => setFeedbackVisible(false)}>
        <View style={st.modalScrim}>
          <View style={st.feedbackModal}>
            <Text style={st.modalTitle}>意见反馈</Text>
            <Text style={st.modalHint}>哪里用起来不顺，直接写给我们。</Text>
            <TextInput
              style={st.feedbackInput}
              value={feedbackContent}
              onChangeText={setFeedbackContent}
              placeholder="写下你的反馈"
              placeholderTextColor="#b5b0a8"
              multiline
              maxLength={1000}
              textAlignVertical="top"
            />
            <View style={st.modalActions}>
              <TouchableOpacity style={st.modalSecondary} onPress={() => setFeedbackVisible(false)} disabled={feedbackSending}>
                <Text style={st.modalSecondaryText}>取消</Text>
              </TouchableOpacity>
              <TouchableOpacity style={st.modalPrimary} onPress={handleSubmitFeedback} disabled={feedbackSending}>
                <Text style={st.modalPrimaryText}>{feedbackSending ? '提交中' : '提交'}</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </SafeAreaView>
  );
}

const st = StyleSheet.create({
  container: {flex: 1, backgroundColor: '#f8f4ed'},
  scroll: {flex: 1},
  scrollContent: {paddingBottom: 28},
  center: {flex: 1, justifyContent: 'center', alignItems: 'center'},
  profileRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingTop: 30,
    paddingBottom: 22,
    paddingHorizontal: 22,
  },
  avatar: {
    width: 58,
    height: 58,
    borderRadius: 18,
    backgroundColor: '#375943',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 14,
  },
  avatarText: {fontSize: 23, fontWeight: '800', color: '#fffaf0'},
  profileInfo: {flex: 1, minWidth: 0},
  name: {fontSize: 22, fontWeight: '800', color: '#211f1a'},
  subtitle: {fontSize: 13, color: '#8c8172', marginTop: 4},
  inlineNotice: {
    marginHorizontal: 18,
    marginBottom: 12,
    paddingVertical: 10,
    paddingHorizontal: 14,
    borderRadius: 12,
    backgroundColor: '#efe4d3',
    borderWidth: 1,
    borderColor: '#e0cfb7',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 12,
  },
  inlineNoticeText: {fontSize: 13, color: '#76624c', fontWeight: '700', flex: 1},
  inlineRetry: {minHeight: 32, justifyContent: 'center', paddingHorizontal: 10},
  inlineRetryText: {fontSize: 13, color: '#375943', fontWeight: '800'},
  recentPanel: {
    marginHorizontal: 18,
    backgroundColor: '#fffaf0',
    borderRadius: 16,
    padding: 18,
    borderWidth: 1,
    borderColor: '#ece2d3',
  },
  section: {
    marginHorizontal: 18,
    marginTop: 14,
    backgroundColor: '#fffaf0',
    borderRadius: 16,
    padding: 18,
    borderWidth: 1,
    borderColor: '#ece2d3',
  },
  sectionHead: {flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14},
  sectionTitle: {fontSize: 17, fontWeight: '800', color: '#211f1a'},
  sectionHint: {fontSize: 12, fontWeight: '600', color: '#8c8172'},
  rangeTabs: {flexDirection: 'row', backgroundColor: '#f3eadc', borderRadius: 9, padding: 2},
  rangeTab: {paddingHorizontal: 8, paddingVertical: 5, borderRadius: 7},
  rangeTabActive: {backgroundColor: '#375943'},
  rangeText: {fontSize: 11, color: '#766c60', fontWeight: '700'},
  rangeTextActive: {color: '#fffaf0'},
  orbit: {flexDirection: 'row', alignItems: 'stretch', gap: 14},
  orbitNode: {justifyContent: 'center', alignItems: 'center'},
  orbitPrimary: {
    width: 126,
    height: 126,
    borderRadius: 38,
    backgroundColor: '#375943',
  },
  orbitValue: {fontSize: 42, fontWeight: '900', color: '#fffaf0'},
  orbitLabel: {fontSize: 12, color: '#e6ddcf', marginTop: 3, fontWeight: '700'},
  orbitSide: {flex: 1, justifyContent: 'space-between'},
  orbitMini: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#f3eadc',
    borderRadius: 12,
    paddingHorizontal: 12,
    paddingVertical: 9,
  },
  orbitMiniValue: {fontSize: 19, fontWeight: '900', color: '#3f5f4b'},
  orbitMiniLabel: {fontSize: 12, color: '#6f665a', fontWeight: '700'},
  totalLine: {flexDirection: 'row', alignItems: 'baseline', marginBottom: 12},
  totalNum: {fontSize: 36, fontWeight: '900', color: '#211f1a', marginRight: 8},
  totalText: {fontSize: 13, color: '#766c60', fontWeight: '600'},
  stackBar: {
    height: 18,
    flexDirection: 'row',
    overflow: 'hidden',
    borderRadius: 9,
    backgroundColor: '#eee5d8',
  },
  stackFill: {height: 18},
  stackPublic: {backgroundColor: '#4e775d'},
  stackPrivate: {backgroundColor: '#d1a756'},
  legendRow: {flexDirection: 'row', justifyContent: 'space-between', marginTop: 8},
  legendText: {fontSize: 12, color: '#8c8172', fontWeight: '600'},
  sourceRow: {flexDirection: 'row', gap: 10, marginTop: 16},
  sourcePill: {
    flex: 1,
    backgroundColor: '#f3eadc',
    borderRadius: 12,
    paddingVertical: 12,
    alignItems: 'center',
  },
  sourceValue: {fontSize: 21, fontWeight: '900', color: '#375943'},
  sourceLabel: {fontSize: 12, color: '#766c60', marginTop: 2, fontWeight: '700'},
  respondedItem: {
    backgroundColor: '#f3eadc',
    borderRadius: 14,
    padding: 14,
    marginBottom: 10,
  },
  respondedContent: {fontSize: 16, lineHeight: 23, fontWeight: '800', color: '#211f1a'},
  respondedMeta: {flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginTop: 10},
  starLine: {fontSize: 13, color: '#d1a756', letterSpacing: 0},
  taxonomy: {fontSize: 12, color: '#766c60', fontWeight: '700'},
  responseEmpty: {backgroundColor: '#f3eadc', borderRadius: 14, paddingVertical: 18, alignItems: 'center'},
  responseValue: {fontSize: 30, fontWeight: '900', color: '#3f5f4b'},
  responseLabel: {fontSize: 12, color: '#766c60', marginTop: 3, fontWeight: '700'},
  responseNote: {fontSize: 12, lineHeight: 18, color: '#8c8172', marginTop: 12},
  menu: {
    marginHorizontal: 18,
    marginTop: 14,
    backgroundColor: '#fffaf0',
    borderRadius: 16,
    borderWidth: 1,
    borderColor: '#ece2d3',
    overflow: 'hidden',
  },
  menuItem: {flexDirection: 'row', alignItems: 'center', paddingVertical: 15, paddingHorizontal: 16},
  menuIconWrap: {width: 28, justifyContent: 'center', alignItems: 'flex-start'},
  menuLabel: {flex: 1, fontSize: 15, color: '#211f1a', fontWeight: '600'},
  dangerText: {color: '#a5483d'},
  modalScrim: {
    flex: 1,
    backgroundColor: 'rgba(33,31,26,0.34)',
    justifyContent: 'center',
    paddingHorizontal: 24,
  },
  feedbackModal: {
    backgroundColor: '#fffaf0',
    borderRadius: 16,
    padding: 18,
    borderWidth: 1,
    borderColor: '#e3d7c5',
  },
  modalTitle: {fontSize: 20, fontWeight: '900', color: '#211f1a'},
  modalHint: {fontSize: 13, lineHeight: 20, color: '#766c60', marginTop: 6, marginBottom: 14},
  feedbackInput: {
    minHeight: 132,
    borderRadius: 12,
    backgroundColor: '#f3eadc',
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 15,
    lineHeight: 22,
    color: '#211f1a',
  },
  modalActions: {flexDirection: 'row', justifyContent: 'flex-end', gap: 10, marginTop: 14},
  modalSecondary: {paddingHorizontal: 16, paddingVertical: 11, borderRadius: 10, backgroundColor: '#eee5d8'},
  modalSecondaryText: {fontSize: 14, fontWeight: '800', color: '#6f665a'},
  modalPrimary: {paddingHorizontal: 18, paddingVertical: 11, borderRadius: 10, backgroundColor: '#375943'},
  modalPrimaryText: {fontSize: 14, fontWeight: '800', color: '#fffaf0'},
  loginBtn: {backgroundColor: '#375943', borderRadius: 12, paddingHorizontal: 36, paddingVertical: 14},
  loginText: {color: '#fffaf0', fontSize: 16, fontWeight: '700'},
});
