import React, {useState, useEffect, useCallback, useRef} from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Dimensions,
  Alert,
  Animated,
} from 'react-native';
import {useSafeAreaInsets} from 'react-native-safe-area-context';
import {
  fetchRecommendations,
  fetchExperiences,
  Experience,
  toggleLike,
  toggleBookmark,
  deleteExperience,
} from '../services/api';
import {getToken, getUserInfo} from '../services/config';

// Tab bar height (React Navigation bottom tab + safe area ≈ 80px)
const TAB_BAR_ESTIMATE = 80;
const SCREEN_HEIGHT = Dimensions.get('window').height;
const PAGE_SIZE = 20;

const DOMAIN_LABELS: Record<string, string> = {
  career: '职场成长', relationship: '人际关系', cognition: '认知升级',
  life: '生活智慧', emotion: '情感',
};
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

// ══════════════════════════════════════════
// FlipCard — 3D 翻转卡片组件
// ══════════════════════════════════════════
function FlipCard({item, currentUserId, cardHeight, onLike, onBookmark, onDelete}: {
  item: Experience;
  currentUserId: string | null;
  cardHeight: number;
  onLike: (id: string) => void;
  onBookmark: (id: string) => void;
  onDelete: (id: string) => void;
}) {
  const flipAnim = useRef(new Animated.Value(0)).current;
  const [isFlipped, setIsFlipped] = useState(false);

  const isPlatform = item.source_type === 'platform';
  const domainLabel = SUB_LABELS[item.sub_domain] || DOMAIN_LABELS[item.domain] || item.domain;
  const displayName = item.creator_name || item.author_name || '匿名';
  const showScore = item.quality_score != null && item.quality_score > 0;
  const stars = showScore ? Math.round(item.quality_score! / 2) : 0;

  const handleFlip = () => {
    if (!item.interpretation) return; // 没有解读不能翻
    const toValue = isFlipped ? 0 : 1;
    Animated.spring(flipAnim, {
      toValue,
      friction: 8,
      tension: 60,
      useNativeDriver: true,
    }).start();
    setIsFlipped(!isFlipped);
  };

  const frontInterpolate = flipAnim.interpolate({
    inputRange: [0, 1],
    outputRange: ['0deg', '180deg'],
  });
  const backInterpolate = flipAnim.interpolate({
    inputRange: [0, 1],
    outputRange: ['180deg', '360deg'],
  });

  const frontOpacity = flipAnim.interpolate({
    inputRange: [0, 0.5, 1],
    outputRange: [1, 0, 0],
  });
  const backOpacity = flipAnim.interpolate({
    inputRange: [0, 0.5, 1],
    outputRange: [0, 0, 1],
  });

  return (
    <TouchableOpacity
      activeOpacity={0.95}
      onPress={handleFlip}
      style={[s.cardPage, {height: cardHeight}]}
    >
      {/* ═══ 正面 — 经验内容 ═══ */}
      <Animated.View
        style={[
          s.face,
          {
            transform: [{perspective: 1000}, {rotateY: frontInterpolate}],
            opacity: frontOpacity,
          },
        ]}
        pointerEvents={isFlipped ? 'none' : 'auto'}
      >
        {/* Top tags */}
        <View style={s.topRow}>
          <View style={s.tag}><Text style={s.tagText}>{domainLabel}</Text></View>
          {isPlatform && <View style={s.platformTag}><Text style={s.platformTagText}>官</Text></View>}
        </View>

        {/* Flip hint */}
        {item.interpretation ? (
          <View style={s.flipHint}>
            <Text style={s.flipHintText}>点击翻转 ↻</Text>
          </View>
        ) : null}

        {/* Content */}
        <View style={s.frontContent}>
          <Text style={s.quoteMark}>"</Text>
          <Text style={s.content}>{item.content}</Text>
          <View style={s.divider} />

          {/* Creator info */}
          <View style={s.creatorRow}>
            <View style={s.avatar}><Text style={s.avatarText}>{displayName.charAt(0)}</Text></View>
            <View>
              <Text style={s.creatorName}>{displayName}</Text>
              {item.source_label ? <Text style={s.sourceLabel}>{item.source_label}</Text> : null}
            </View>
          </View>

          {/* Stars */}
          {showScore && (
            <View style={s.starRow}>
              <Text style={s.stars}>{'★'.repeat(stars)}{'☆'.repeat(5 - stars)}</Text>
              {item.score_reason ? <Text style={s.scoreText}>{item.score_reason}</Text> : null}
            </View>
          )}
        </View>

        {/* Bottom actions — always visible on front */}
        <View style={s.bottomActions}>
          <TouchableOpacity
            style={[s.actionBtn, item.is_liked && s.actionLiked]}
            onPress={(e) => { e.stopPropagation(); onLike(item.id); }}
          >
            <Text style={[s.actionText, item.is_liked && s.actionLikedText]}>
              ♥ {item.like_count > 0 ? item.like_count : '点赞'}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[s.actionBtn, item.is_bookmarked && s.actionSaved]}
            onPress={(e) => { e.stopPropagation(); onBookmark(item.id); }}
          >
            <Text style={[s.actionText, item.is_bookmarked && s.actionSavedText]}>
              ★ {item.is_bookmarked ? '已收藏' : '收藏'}
            </Text>
          </TouchableOpacity>
          {currentUserId && item.author_id === currentUserId && (
            <TouchableOpacity
              style={s.deleteBtn}
              onPress={(e) => { e.stopPropagation(); onDelete(item.id); }}
            >
              <Text style={s.deleteText}>删除</Text>
            </TouchableOpacity>
          )}
        </View>
      </Animated.View>

      {/* ═══ 背面 — 经验解读 ═══ */}
      <Animated.View
        style={[
          s.face,
          {
            transform: [{perspective: 1000}, {rotateY: backInterpolate}],
            opacity: backOpacity,
          },
        ]}
        pointerEvents={isFlipped ? 'auto' : 'none'}
      >
        {/* Back header */}
        <View style={s.backHeader}>
          <Text style={s.backHeaderTitle}>经验解读</Text>
          <View style={s.backDomainTag}>
            <Text style={s.backDomainText}>{domainLabel}</Text>
          </View>
        </View>

        {/* Interpretation content */}
        <View style={s.backContent}>
          <Text style={s.backQuote}>“{item.content}”</Text>
          <View style={s.backDivider} />
          <Text style={s.interpText}>
            {item.interpretation || '暂无解读'}
          </Text>
        </View>

        {/* Back tap hint */}
        <View style={s.backHint}>
          <Text style={s.backHintText}>点击翻回正面 ↻</Text>
        </View>
      </Animated.View>
    </TouchableOpacity>
  );
}

// ══════════════════════════════════════════
// HomeScreen
// ══════════════════════════════════════════
export default function HomeScreen() {
  const insets = useSafeAreaInsets();
  const CARD_HEIGHT = SCREEN_HEIGHT - insets.top - TAB_BAR_ESTIMATE;

  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isPersonalized, setIsPersonalized] = useState(false);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);

  const loadingMoreRef = useRef(false);
  const hasMoreRef = useRef(true);
  const tokenRef = useRef<string | null>(null);
  const offsetRef = useRef(0);

  useEffect(() => {
    getToken().then(t => { tokenRef.current = t; });
    getUserInfo().then(u => setCurrentUserId(u?.id || null));
  }, []);

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
      setHasMore(false);
    }
    if (append) {
      setExperiences(prev => {
        const ids = new Set(prev.map(e => e.id));
        return [...prev, ...data.filter((e: Experience) => !ids.has(e.id))];
      });
    } else {
      hasMoreRef.current = true;
      setHasMore(true);
      offsetRef.current = 0;
      setExperiences(data);
    }
    offsetRef.current += data.length;
    return data.length;
  }, []);

  const loadInitial = useCallback(async () => {
    try { await loadPage(0, false); }
    catch (e) { console.error(e); setError('加载失败'); }
    finally { setLoading(false); }
  }, [loadPage]);

  useEffect(() => { loadInitial(); }, []);

  const handleLoadMore = useCallback(async () => {
    if (loadingMoreRef.current || !hasMoreRef.current || offsetRef.current === 0) return;
    loadingMoreRef.current = true;
    setLoadingMore(true);
    try { await loadPage(offsetRef.current, true); }
    catch (e) { console.error(e); }
    finally { loadingMoreRef.current = false; setLoadingMore(false); }
  }, [loadPage]);

  const handleLike = async (id: string) => {
    setExperiences(prev => prev.map(e =>
      e.id === id ? {...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1} : e
    ));
    try { await toggleLike(id); } catch {
      setExperiences(prev => prev.map(e =>
        e.id === id ? {...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1} : e
      ));
    }
  };

  const handleBookmark = async (id: string) => {
    setExperiences(prev => prev.map(e => e.id === id ? {...e, is_bookmarked: !e.is_bookmarked} : e));
    try { await toggleBookmark(id); } catch {
      setExperiences(prev => prev.map(e => e.id === id ? {...e, is_bookmarked: !e.is_bookmarked} : e));
    }
  };

  const handleDelete = (id: string) => {
    Alert.alert('删除经验', '确定要删除这条经验吗？', [
      {text: '取消', style: 'cancel'},
      {text: '删除', style: 'destructive', onPress: async () => {
        try {
          await deleteExperience(id);
          setExperiences(prev => prev.filter(e => e.id !== id));
        } catch (e: any) {
          Alert.alert('删除失败', e?.message || '请稍后再试');
        }
      }},
    ]);
  };

  if (loading) {
    return <View style={s.container}><ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 200}} /></View>;
  }
  if (error && experiences.length === 0) {
    return (
      <View style={s.container}>
        <View style={{flex:1,justifyContent:'center',alignItems:'center',paddingBottom:80}}>
          <Text style={{fontSize:15,color:'#9a9a9a',marginBottom:16}}>{error}</Text>
          <TouchableOpacity style={{backgroundColor:'#4a7c59',borderRadius:20,paddingHorizontal:24,paddingVertical:10}} onPress={() => { setError(null); loadInitial(); }}>
            <Text style={{color:'#fff',fontSize:14,fontWeight:'600'}}>重试</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  return (
    <View style={s.container}>
      <FlatList
        data={experiences}
        keyExtractor={item => item.id}
        renderItem={({item}) => (
          <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5'}}>
            <FlipCard
              item={item}
              currentUserId={currentUserId}
              cardHeight={CARD_HEIGHT}
              onLike={handleLike}
              onBookmark={handleBookmark}
              onDelete={handleDelete}
            />
          </View>
        )}
        getItemLayout={(_, index) => ({
          length: CARD_HEIGHT,
          offset: CARD_HEIGHT * index,
          index,
        })}
        snapToInterval={CARD_HEIGHT}
        snapToAlignment="start"
        disableIntervalMomentum
        showsVerticalScrollIndicator={false}
        decelerationRate="fast"
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.5}
        ListEmptyComponent={
          <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5', justifyContent: 'center', alignItems: 'center'}}>
            <Text style={{fontSize:15,color:'#9a9a9a'}}>暂无推荐内容</Text>
            <Text style={{fontSize:12,color:'#b5b0a8',marginTop:6}}>发布经验后，推荐会更精准</Text>
          </View>
        }
        ListFooterComponent={
          loadingMore ? (
            <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5', justifyContent: 'center', alignItems: 'center'}}>
              <ActivityIndicator size="small" color="#4a7c59" />
            </View>
          ) : !hasMore && experiences.length > 0 ? (
            <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5', justifyContent: 'center', alignItems: 'center'}}>
              <Text style={{fontSize:14,color:'#b5b0a8'}}>— 已经到底了 —</Text>
            </View>
          ) : null
        }
      />
    </View>
  );
}

const s = StyleSheet.create({
  container: {flex: 1, backgroundColor: '#faf8f5'},

  // ═══ Card page (full screen) ═══
  cardPage: {
    backgroundColor: '#faf8f5',
  },
  face: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: '#faf8f5',
    backfaceVisibility: 'hidden' as const,
  },

  // ═══ Top row (fixed) ═══
  topRow: {
    position: 'absolute', top: 60, left: 24,
    flexDirection: 'row', gap: 6, zIndex: 2,
  },
  tag: {backgroundColor: '#eaf2e8', paddingHorizontal: 10, paddingVertical: 4, borderRadius: 12},
  tagText: {fontSize: 11, fontWeight: '600', color: '#4a7c59'},
  platformTag: {
    backgroundColor: '#4a7c59', width: 18, height: 18, borderRadius: 9,
    justifyContent: 'center', alignItems: 'center',
  },
  platformTagText: {fontSize: 9, fontWeight: '800', color: '#fff'},

  // ═══ Flip hint ═══
  flipHint: {
    position: 'absolute', top: 62, right: 70, zIndex: 2,
  },
  flipHintText: {fontSize: 11, color: '#c4c0b8', fontWeight: '500'},

  // ═══ Front — content area ═══
  frontContent: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
    paddingBottom: 120,
  },
  quoteMark: {
    fontSize: 72, color: '#4a7c59', opacity: 0.12,
    fontFamily: 'Georgia', lineHeight: 72, marginBottom: -20,
  },
  content: {
    fontSize: 26, fontWeight: '700', lineHeight: 40,
    color: '#1a1a1a', textAlign: 'center',
  },
  divider: {width: 40, height: 2, backgroundColor: '#d4e0d6', marginVertical: 24},

  // Creator
  creatorRow: {flexDirection: 'row', alignItems: 'center', gap: 8},
  avatar: {
    width: 32, height: 32, borderRadius: 16, backgroundColor: '#eaf2e8',
    justifyContent: 'center', alignItems: 'center',
  },
  avatarText: {fontSize: 13, fontWeight: '700', color: '#4a7c59'},
  creatorName: {fontSize: 14, fontWeight: '600', color: '#4a4a4a'},
  sourceLabel: {fontSize: 11, color: '#9a9a9a', marginTop: 1},

  // Stars
  starRow: {flexDirection: 'row', alignItems: 'center', gap: 4, marginTop: 8},
  stars: {fontSize: 16, color: '#e8a850', letterSpacing: 2},
  scoreText: {fontSize: 10, color: '#b5b0a8'},

  // ═══ Bottom actions (fixed) ═══
  bottomActions: {
    position: 'absolute', bottom: 50,
    left: 0, right: 0,
    flexDirection: 'row', justifyContent: 'center', gap: 12,
    zIndex: 2,
  },
  actionBtn: {
    paddingHorizontal: 16, paddingVertical: 8,
    borderRadius: 20, backgroundColor: '#fff',
    borderWidth: 0.5, borderColor: '#f0ece7',
  },
  actionText: {fontSize: 13, color: '#9a9a9a', fontWeight: '500'},
  actionLiked: {borderColor: '#fce8e8', backgroundColor: '#fff5f5'},
  actionLikedText: {color: '#e85d5d'},
  actionSaved: {borderColor: '#fdf3e4', backgroundColor: '#fffcf5'},
  actionSavedText: {color: '#e8a850'},
  deleteBtn: {
    paddingHorizontal: 16, paddingVertical: 8,
    borderRadius: 20, backgroundColor: '#fff5f5',
    borderWidth: 0.5, borderColor: '#fce8e8',
  },
  deleteText: {fontSize: 13, color: '#e85d5d', fontWeight: '500'},

  // ═══ Back face — interpretation ═══
  backHeader: {
    position: 'absolute', top: 60, left: 24, right: 24,
    flexDirection: 'row', alignItems: 'center', gap: 10,
  },
  backHeaderTitle: {
    fontSize: 15, fontWeight: '700', color: '#4a7c59',
  },
  backDomainTag: {
    backgroundColor: '#eaf2e8', paddingHorizontal: 10, paddingVertical: 3, borderRadius: 10,
  },
  backDomainText: {fontSize: 11, fontWeight: '600', color: '#4a7c59'},

  backContent: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: 32,
    paddingBottom: 80,
  },
  backQuote: {
    fontSize: 17, fontWeight: '600', color: '#4a4a4a',
    lineHeight: 28, textAlign: 'center', fontStyle: 'italic',
    marginBottom: 20,
  },
  backDivider: {
    width: 40, height: 2, backgroundColor: '#d4e0d6',
    alignSelf: 'center', marginBottom: 20,
  },
  interpText: {
    fontSize: 15, lineHeight: 26, color: '#3d3d3d',
    textAlign: 'left',
  },

  backHint: {
    position: 'absolute', bottom: 40, left: 0, right: 0,
    alignItems: 'center',
  },
  backHintText: {fontSize: 12, color: '#b5b0a8'},

});
