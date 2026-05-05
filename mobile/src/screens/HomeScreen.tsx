import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  ActivityIndicator,
  Modal,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  fetchRecommendations,
  fetchExperiences,
  Experience,
  toggleLike,
  toggleBookmark,
} from '../services/api';
import { getToken } from '../services/config';

function StarRating({ score, reason }: { score: number; reason?: string }) {
  const [showReason, setShowReason] = useState(false);
  const stars = Math.round(score / 2);
  return (
    <>
      <TouchableOpacity
        style={starStyles.row}
        onPress={() => reason && setShowReason(true)}
        activeOpacity={reason ? 0.6 : 1}
      >
        {[1, 2, 3, 4, 5].map(i => (
          <Text key={i} style={starStyles.star}>
            {i <= stars ? '★' : '☆'}
          </Text>
        ))}
        <Text style={starStyles.label}>价值度</Text>
      </TouchableOpacity>
      {reason && (
        <Modal visible={showReason} transparent animationType="fade">
          <TouchableOpacity
            style={starStyles.overlay}
            activeOpacity={1}
            onPress={() => setShowReason(false)}
          >
            <View style={starStyles.popup}>
              <Text style={starStyles.popupTitle}>价值度 · {stars}星</Text>
              <Text style={starStyles.popupText}>{reason}</Text>
            </View>
          </TouchableOpacity>
        </Modal>
      )}
    </>
  );
}

const starStyles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', gap: 1 },
  star: { fontSize: 12, color: '#e8a850' },
  label: { fontSize: 10, color: '#9a9a9a', marginLeft: 4 },
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.35)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  popup: {
    backgroundColor: '#ffffff',
    borderRadius: 16,
    padding: 24,
    marginHorizontal: 40,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.12,
    shadowRadius: 12,
    elevation: 6,
  },
  popupTitle: { fontSize: 16, fontWeight: '700', color: '#1a1a1a', marginBottom: 10 },
  popupText: { fontSize: 14, color: '#4a4a4a', lineHeight: 22, textAlign: 'center' },
});

export default function HomeScreen({ navigation }: any) {
  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isPersonalized, setIsPersonalized] = useState(false);

  const PAGE_SIZE = 20;
  const loadingMoreRef = useRef(false);
  const hasMoreRef = useRef(true);
  const tokenRef = useRef<string | null>(null);
  const offsetRef = useRef(0);

  // Init token on mount
  useEffect(() => { getToken().then(t => { tokenRef.current = t; }); }, []);

  const loadPage = useCallback(async (offset: number, append: boolean) => {
    let result;
    if (tokenRef.current) {
      result = await fetchRecommendations(PAGE_SIZE, offset);
      setIsPersonalized(true);
    } else {
      result = await fetchExperiences(Math.floor(offset / PAGE_SIZE) + 1);
      setIsPersonalized(false);
    }
    const data = Array.isArray(result?.data) ? result.data : [];
    if (data.length < PAGE_SIZE) {
      hasMoreRef.current = false;
    }
    if (append) {
      setExperiences(prev => [...prev, ...data]);
    } else {
      hasMoreRef.current = true;
      offsetRef.current = 0;
      setExperiences(data);
    }
    setError(null);
    offsetRef.current += data.length;
    return data.length;
  }, []);

  const loadInitial = useCallback(async () => {
    try {
      await loadPage(0, false);
    } catch (e) {
      console.error('Failed to load:', e);
      setError('加载失败，请检查网络连接');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [loadPage]);

  useEffect(() => { loadInitial(); }, []);

  const handleRefresh = () => {
    setRefreshing(true);
    loadInitial();
  };

  const handleLoadMore = useCallback(async () => {
    if (loadingMoreRef.current || !hasMoreRef.current) return;
    loadingMoreRef.current = true;
    setLoadingMore(true);
    try {
      await loadPage(offsetRef.current, true);
    } catch (e) {
      console.error('Failed to load more:', e);
    } finally {
      loadingMoreRef.current = false;
      setLoadingMore(false);
    }
  }, [loadPage]);

  const handleLike = async (id: string) => {
    setExperiences(prev => prev.map(e =>
      e.id === id ? { ...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1 } : e
    ));
    try { await toggleLike(id); } catch (e) {
      setExperiences(prev => prev.map(e =>
        e.id === id ? { ...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1 } : e
      ));
    }
  };

  const handleBookmark = async (id: string) => {
    setExperiences(prev => prev.map(e => (e.id === id ? { ...e, is_bookmarked: !e.is_bookmarked } : e)));
    try { await toggleBookmark(id); } catch (e) {
      setExperiences(prev => prev.map(e => (e.id === id ? { ...e, is_bookmarked: !e.is_bookmarked } : e)));
    }
  };

  const domainLabels: Record<string, string> = {
    career: '职场成长', relationship: '人际关系', cognition: '认知升级',
    life: '生活智慧', emotion: '情感',
  };
  const subDomainLabels: Record<string, string> = {
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

  if (loading) {
    return <SafeAreaView style={s.container}><ActivityIndicator size="large" color="#4a7c59" style={{ marginTop: 200 }} /></SafeAreaView>;
  }
  if (error && experiences.length === 0) {
    return (
      <SafeAreaView style={s.container}>
        <View style={s.errorContainer}>
          <Text style={s.errorText}>{error}</Text>
          <TouchableOpacity style={s.retryButton} onPress={() => { setError(null); handleRefresh(); }}>
            <Text style={s.retryButtonText}>重试</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  const renderItem = ({ item }: { item: Experience }) => {
    const isPlatform = item.source_type === 'platform';
    const isRejected = item.review_status === 'rejected';
    const showScore = item.quality_score != null && item.quality_score > 0;
    const displayName = item.creator_name || item.author_name || '匿名';

    return (
      <TouchableOpacity
        style={s.card}
        onPress={() => navigation.navigate('detail', { id: item.id })}
        activeOpacity={0.8}
      >
        {/* Author row */}
        <View style={s.authorRow}>
          <View style={s.avatar}>
            <Text style={s.avatarText}>{displayName.charAt(0)}</Text>
          </View>
          <Text style={s.authorName} numberOfLines={1}>{displayName}</Text>

          {/* Platform badge */}
          {isPlatform && (
            <View style={s.platformBadge}>
              <Text style={s.platformBadgeText}>官</Text>
            </View>
          )}

          {/* Rejected indicator */}
          {isRejected && (
            <View style={s.rejectedBadge}>
              <Text style={s.rejectedIcon}>❕</Text>
            </View>
          )}

          <View style={s.domainTag}>
            <Text style={s.domainText}>
              {subDomainLabels[item.sub_domain] || domainLabels[item.domain] || item.domain}
            </Text>
          </View>
        </View>

        {/* Content */}
        <Text style={s.content}>{item.content}</Text>

        {/* Bottom row: score + actions */}
        <View style={s.bottomRow}>
          {showScore && <StarRating score={item.quality_score!} reason={item.score_reason} />}
          <View style={{ flex: 1 }} />
          <View style={s.actions}>
            <TouchableOpacity onPress={() => handleLike(item.id)} style={s.actionButton}>
              <Text style={[s.actionText, item.is_liked && s.actionLiked]}>
                ♥ {item.like_count}
              </Text>
            </TouchableOpacity>
            <TouchableOpacity onPress={() => handleBookmark(item.id)} style={s.actionButton}>
              <Text style={[s.actionText, item.is_bookmarked && s.actionSaved]}>
                ★ {item.is_bookmarked ? '已收藏' : '收藏'}
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView style={s.container}>
      <View style={s.header}>
        <Text style={s.headerTitle}>为你推荐</Text>
        <Text style={s.headerSub}>
          {isPersonalized ? '基于你的偏好 · 个性化推荐' : '热门经验精选'}
        </Text>
      </View>
      <FlatList
        data={experiences}
        keyExtractor={item => item.id}
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={handleRefresh} tintColor="#4a7c59" />}
        contentContainerStyle={s.list}
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.3}
        ListFooterComponent={
          loadingMore ? (
            <View style={{ paddingVertical: 20 }}>
              <ActivityIndicator size="small" color="#4a7c59" />
            </View>
          ) : null
        }
        ListEmptyComponent={
          <View style={s.emptyContainer}>
            <Text style={s.emptyText}>暂无推荐内容</Text>
            <Text style={s.emptyHint}>发布经验后，推荐会更精准</Text>
          </View>
        }
        renderItem={renderItem}
      />
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#faf8f5' },
  header: { paddingHorizontal: 18, paddingTop: 4, paddingBottom: 8 },
  headerTitle: { fontSize: 13, fontWeight: '700', color: '#9a9a9a', letterSpacing: 1, textTransform: 'uppercase' },
  headerSub: { fontSize: 11, color: '#9a9a9a', marginTop: 1 },
  list: { paddingHorizontal: 14, paddingBottom: 20 },
  card: {
    backgroundColor: '#ffffff', borderRadius: 16, padding: 16, marginBottom: 10,
    borderWidth: 0.5, borderColor: '#f0ece7', shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 }, shadowOpacity: 0.04, shadowRadius: 6, elevation: 1,
  },
  authorRow: { flexDirection: 'row', alignItems: 'center', marginBottom: 8 },
  avatar: {
    width: 24, height: 24, borderRadius: 12, backgroundColor: '#eaf2e8',
    justifyContent: 'center', alignItems: 'center', marginRight: 6,
  },
  avatarText: { fontSize: 11, fontWeight: '700', color: '#4a7c59' },
  authorName: { fontSize: 12, fontWeight: '600', color: '#4a4a4a', maxWidth: 100 },
  platformBadge: {
    marginLeft: 4, backgroundColor: '#4a7c59', width: 18, height: 18, borderRadius: 9,
    justifyContent: 'center', alignItems: 'center',
  },
  platformBadgeText: { fontSize: 10, fontWeight: '800', color: '#ffffff' },
  rejectedBadge: { marginLeft: 4 },
  rejectedIcon: { fontSize: 14, color: '#9a9a9a' },
  domainTag: { backgroundColor: '#eaf2e8', paddingHorizontal: 8, paddingVertical: 2, borderRadius: 8, marginLeft: 'auto' },
  domainText: { fontSize: 10, fontWeight: '600', color: '#4a7c59' },
  content: { fontSize: 15, lineHeight: 23, fontWeight: '600', color: '#1a1a1a', marginBottom: 10 },
  bottomRow: { flexDirection: 'row', alignItems: 'center', paddingTop: 8, borderTopWidth: 0.5, borderTopColor: '#f0ece7' },
  actions: { flexDirection: 'row', gap: 10 },
  actionButton: { paddingVertical: 2 },
  actionText: { fontSize: 11, color: '#9a9a9a' },
  actionLiked: { color: '#e85d5d' },
  actionSaved: { color: '#e8a850' },
  errorContainer: { flex: 1, justifyContent: 'center', alignItems: 'center', paddingBottom: 80 },
  errorText: { fontSize: 15, color: '#9a9a9a', marginBottom: 16 },
  retryButton: { backgroundColor: '#4a7c59', borderRadius: 20, paddingHorizontal: 24, paddingVertical: 10 },
  retryButtonText: { color: '#ffffff', fontSize: 14, fontWeight: '600' },
  emptyContainer: { paddingTop: 100, alignItems: 'center' },
  emptyText: { fontSize: 15, color: '#9a9a9a' },
  emptyHint: { fontSize: 12, color: '#b5b0a8', marginTop: 6 },
});
