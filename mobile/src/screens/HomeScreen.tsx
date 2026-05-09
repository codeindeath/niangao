import React, {useState, useEffect, useCallback, useRef} from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  Dimensions,
  Alert,
  StyleSheet,
} from 'react-native';
import {useSafeAreaInsets} from 'react-native-safe-area-context';
import {useFocusEffect, useNavigation} from '@react-navigation/native';
import {
  fetchRecommendations,
  fetchExperiences,
  fetchMyExperiences,
  fetchMyBookmarks,
  Experience,
  toggleLike,
  toggleBookmark,
  deleteExperience,
} from '../services/api';
import {getToken, getUserInfo} from '../services/config';
import FlipCard from '../components/ExperienceCard';
import {recordView} from '../services/api';

// Module-level tab refresh trigger (called from CreateScreen after publish)
let _pendingTabRefresh: string | null = null;
export function triggerTabRefresh(tab: string) { _pendingTabRefresh = tab; }

// Tab bar height (React Navigation bottom tab + safe area ≈ 80px)
const TAB_BAR_ESTIMATE = 80;
const TOP_TABS_HEIGHT = 44; // 方案A 顶部三标签栏
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
// HomeScreen
// ══════════════════════════════════════════
export default function HomeScreen() {
  const insets = useSafeAreaInsets();
  const navigation = useNavigation<any>();
  // card = screen - status bar - top tabs(44) - bottom tab(80) - margin(8)
  const CARD_HEIGHT = SCREEN_HEIGHT - insets.top - TOP_TABS_HEIGHT - TAB_BAR_ESTIMATE - 8;
  // 卡片内容区从 tab bar 下方 12px 开始
  const CONTENT_TOP = insets.top + TOP_TABS_HEIGHT + 12;

  type TabName = 'recommend' | 'my' | 'bookmarks';
  const tabOrder: TabName[] = ['recommend', 'my', 'bookmarks'];
  const [activeTab, setActiveTab] = useState<TabName>('recommend');

  type TabCache = { items: Experience[]; offset: number; hasMore: boolean; loaded: boolean };
  const defaultCache = (): TabCache => ({ items: [], offset: 0, hasMore: true, loaded: false });
  const [tabCaches, setTabCaches] = useState<Record<TabName, TabCache>>({
    recommend: defaultCache(), my: defaultCache(), bookmarks: defaultCache(),
  });

  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isPersonalized, setIsPersonalized] = useState(false);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [flippedCards, setFlippedCards] = useState<Set<string>>(new Set());
  const [visibleCardId, setVisibleCardId] = useState<string | null>(null);

  const loadingMoreRef = useRef(false);
  const tabScrollIndex = useRef<Record<TabName, number>>({ recommend: 0, my: 0, bookmarks: 0 });
  const flatListRef = useRef<FlatList>(null);
  const tokenRef = useRef<string | null>(null);

  // Derived
  const experiences = tabCaches[activeTab].items;
  const hasMore = tabCaches[activeTab].hasMore;

  useEffect(() => {
    getToken().then(t => { tokenRef.current = t; loadInitial('recommend'); });
    getUserInfo().then(u => setCurrentUserId(u?.id || null));
  }, []);

  // When screen gains focus, check if a tab needs refresh (e.g. after publishing)
  useFocusEffect(
    useCallback(() => {
      if (_pendingTabRefresh) {
        const tab = _pendingTabRefresh as TabName;
        _pendingTabRefresh = null;
        setTabCaches(prev => ({
          ...prev,
          [tab]: { items: [], offset: 0, hasMore: true, loaded: false },
        }));
      }
    }, [])
  );

  const loadPage = useCallback(async (offset: number, append: boolean, tab: TabName) => {
    let result;
    if (!tokenRef.current || tab !== 'recommend') {
      if (tab === 'bookmarks') {
        result = await fetchMyBookmarks(Math.floor(offset / PAGE_SIZE) + 1);
      } else if (tab === 'my') {
        result = await fetchMyExperiences(Math.floor(offset / PAGE_SIZE) + 1);
      } else {
        result = await fetchExperiences(Math.floor(offset / PAGE_SIZE) + 1);
      }
      setIsPersonalized(false);
    } else {
      result = await fetchRecommendations(PAGE_SIZE, offset);
      setIsPersonalized(true);
    }
    const data = Array.isArray(result?.data) ? result.data : [];
    const noMore = data.length < PAGE_SIZE;

    setTabCaches(prev => {
      const cache = prev[tab];
      let items: Experience[];
      if (append) {
        const ids = new Set(cache.items.map(e => e.id));
        items = [...cache.items, ...data.filter((e: Experience) => !ids.has(e.id))];
      } else {
        items = data;
      }
      return {
        ...prev,
        [tab]: { items, offset: offset + data.length, hasMore: !noMore, loaded: true },
      };
    });
    return data.length;
  }, []);

  const loadInitial = useCallback(async (tab: TabName) => {
    try { await loadPage(0, false, tab); }
    catch (e) { console.error(e); setError('加载失败'); }
    finally { setLoading(false); }
  }, [loadPage]);

  const handleLoadMore = useCallback(async () => {
    const cache = tabCaches[activeTab];
    if (loadingMoreRef.current || !cache.hasMore) return;
    loadingMoreRef.current = true;
    setLoadingMore(true);
    try { await loadPage(cache.offset, true, activeTab); }
    catch (e) { console.error(e); }
    finally { loadingMoreRef.current = false; setLoadingMore(false); }
  }, [loadPage, activeTab, tabCaches]);

  const handleTabChange = (tab: TabName) => {
    if (tab === activeTab) return;
    // Save current scroll index
    tabScrollIndex.current[activeTab] = tabScrollIndex.current[activeTab] || 0;
    setActiveTab(tab);
    setError(null);
    const cache = tabCaches[tab];
    if (cache.loaded) {
      // Restore cached data + scroll position
      setTimeout(() => {
        const idx = tabScrollIndex.current[tab] || 0;
        if (idx > 0 && flatListRef.current) {
          flatListRef.current.scrollToIndex({ index: idx, animated: false, viewPosition: 0 });
        }
      }, 50);
    } else {
      setLoading(true);
      loadPage(0, false, tab).catch(e => { console.error(e); setError('加载失败'); }).finally(() => setLoading(false));
    }
  };

  const handleLike = async (id: string) => {
    setTabCaches(prev => ({
      ...prev,
      [activeTab]: {
        ...prev[activeTab],
        items: prev[activeTab].items.map(e =>
          e.id === id ? {...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1} : e
        ),
      },
    }));
    try { await toggleLike(id); } catch {
      setTabCaches(prev => ({
        ...prev,
        [activeTab]: {
          ...prev[activeTab],
          items: prev[activeTab].items.map(e =>
            e.id === id ? {...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1} : e
          ),
        },
      }));
    }
  };

  const handleBookmark = async (id: string) => {
    setTabCaches(prev => ({
      ...prev,
      [activeTab]: {
        ...prev[activeTab],
        items: prev[activeTab].items.map(e => e.id === id ? {...e, is_bookmarked: !e.is_bookmarked} : e),
      },
    }));
    try { await toggleBookmark(id); } catch {
      setTabCaches(prev => ({
        ...prev,
        [activeTab]: {
          ...prev[activeTab],
          items: prev[activeTab].items.map(e => e.id === id ? {...e, is_bookmarked: !e.is_bookmarked} : e),
        },
      }));
    }
  };

  const handleDelete = (id: string) => {
    Alert.alert('删除经验', '确定要删除这条经验吗？', [
      {text: '取消', style: 'cancel'},
      {text: '删除', style: 'destructive', onPress: async () => {
        try {
          await deleteExperience(id);
          setTabCaches(prev => ({
            ...prev,
            [activeTab]: {
              ...prev[activeTab],
              items: prev[activeTab].items.filter(e => e.id !== id),
            },
          }));
        } catch (e: any) {
          Alert.alert('删除失败', e?.message || '请稍后再试');
        }
      }},
    ]);
  };

  const handleFlipChange = useCallback((id: string, isFlipped: boolean) => {
    setFlippedCards(prev => {
      const next = new Set(prev);
      if (isFlipped) next.add(id); else next.delete(id);
      return next;
    });
  }, []);

  const viewabilityConfig = useRef({itemVisiblePercentThreshold: 50}).current;

  const onViewableItemsChanged = useRef(({viewableItems}: any) => {
    if (viewableItems.length > 0) {
      const item = viewableItems[0];
      setVisibleCardId(item.item.id);
      tabScrollIndex.current[activeTab] = item.index;
      recordView(item.item.id);
    }
  }).current;

  // ═══ 左右滑动切标签（Touch 事件，不干扰 FlatList/卡片） ═══
  // 左滑(dx<0): 推荐→我的→收藏→推荐（循环）
  // 右滑(dx>0): 推荐→收藏→我的→推荐（循环）
  const swipeStart = useRef<{x: number; y: number} | null>(null);

  const handleTouchStart = (e: any) => {
    const {pageX, pageY} = e.nativeEvent;
    swipeStart.current = {x: pageX, y: pageY};
  };

  const handleTouchEnd = (e: any) => {
    const s = swipeStart.current;
    swipeStart.current = null;
    if (!s) return;
    const dx = e.nativeEvent.pageX - s.x;
    const dy = e.nativeEvent.pageY - s.y;
    if (Math.abs(dx) > 60 && Math.abs(dx) > Math.abs(dy) * 2) {
      const idx = tabOrder.indexOf(activeTab);
      const n = tabOrder.length;
      if (dx < 0) {
        handleTabChange(tabOrder[(idx + 1) % n]);
      } else {
        handleTabChange(tabOrder[(idx - 1 + n) % n]);
      }
    }
  };

  if (loading) {
    return <View style={s.container}><ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 200}} /></View>;
  }
  if (error && experiences.length === 0) {
    return (
      <View style={s.container}>
        <View style={{flex:1,justifyContent:'center',alignItems:'center',paddingBottom:80}}>
          <Text style={{fontSize:15,color:'#9a9a9a',marginBottom:16}}>{error}</Text>
          <TouchableOpacity style={{backgroundColor:'#4a7c59',borderRadius:20,paddingHorizontal:24,paddingVertical:10}} onPress={() => { setError(null); loadInitial(activeTab); }}>
            <Text style={{color:'#fff',fontSize:14,fontWeight:'600'}}>重试</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  return (
    <View
      style={s.container}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      {/* ═══ 方案A 顶部三标签 + 搜索图标 ═══ */}
      <View style={[s.tabBar, {top: insets.top}]}>
        <View style={s.tabBarInner}>
          {tabOrder.map(tab => (
            <TouchableOpacity key={tab} onPress={() => handleTabChange(tab)} style={s.tabItem}>
              <Text style={[s.tabLabel, activeTab === tab && s.tabLabelActive]}>
                {tab === 'recommend' ? '推荐' : tab === 'my' ? '我的' : '收藏'}
              </Text>
              {activeTab === tab && <View style={s.tabUnderline} />}
            </TouchableOpacity>
          ))}
        </View>
        <TouchableOpacity onPress={() => navigation.navigate('searchPage')} style={s.searchIconBtn}>
          <Text style={s.searchIconText}>🔍</Text>
        </TouchableOpacity>
      </View>

      <FlatList
        ref={flatListRef}
        data={experiences}
        keyExtractor={item => item.id}
        renderItem={({item}) => (
          <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5', overflow: 'visible'}}>
            <FlipCard
              item={item}
              currentUserId={currentUserId}
              cardHeight={CARD_HEIGHT}
              contentTop={CONTENT_TOP}
              onLike={handleLike}
              onBookmark={handleBookmark}
              onDelete={handleDelete}
              onFlipChange={handleFlipChange}
              showActions
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
        onViewableItemsChanged={onViewableItemsChanged}
        viewabilityConfig={viewabilityConfig}
        onScrollToIndexFailed={(info) => {
          // Retry after layout settles
          setTimeout(() => {
            if (flatListRef.current && info.index < experiences.length) {
              flatListRef.current.scrollToIndex({ index: info.index, animated: false, viewPosition: 0 });
            }
          }, 200);
        }}
        ListEmptyComponent={
          <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5', justifyContent: 'center', alignItems: 'center'}}>
            <Text style={{fontSize:15,color:'#9a9a9a'}}>
              {activeTab === 'recommend' ? '暂无推荐内容' : activeTab === 'my' ? '你还没有发布经验' : '你还没有收藏经验'}
            </Text>
            <Text style={{fontSize:12,color:'#b5b0a8',marginTop:6}}>
              {activeTab === 'recommend' ? '发布经验后，推荐会更精准' : activeTab === 'my' ? '分享一条你的经验吧' : '去发现页面收藏喜欢的经验'}
            </Text>
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
      {/* ═══ Flip hint: screen-level overlay, bottom:0 touches tab bar ═══ */}
      {(() => {
        const vExp = visibleCardId ? experiences.find(e => e.id === visibleCardId) : null;
        if (!vExp?.interpretation) return null;
        return (
          <View style={{position: 'absolute', bottom: 4, right: 16, zIndex: 20}} pointerEvents="none">
            <Text style={s.flipHintText}>{flippedCards.has(visibleCardId!) ? '点击回正面' : '点击卡片看解读'}</Text>
          </View>
        );
      })()}
    </View>
  );
}

const s = StyleSheet.create({
  container: {flex: 1, backgroundColor: '#faf8f5'},

  // ═══ 方案A 顶部三标签 + 搜索 ═══
  tabBar: {
    position: 'absolute', left: 0, right: 0, zIndex: 10,
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    paddingHorizontal: 24, paddingTop: 8, paddingBottom: 6,
    backgroundColor: 'transparent',
  },
  tabBarInner: {
    flexDirection: 'row', justifyContent: 'space-around', flex: 1,
  },
  searchIconBtn: {
    width: 36, height: 36,
    justifyContent: 'center', alignItems: 'center',
    marginLeft: 4,
  },
  searchIconText: {fontSize: 18},
  tabItem: { alignItems: 'center', paddingHorizontal: 8 },
  tabLabel: { fontSize: 15, fontWeight: '500', color: '#b5b0a8' },
  tabLabelActive: { color: '#1a1a1a', fontWeight: '700' },
  tabUnderline: { width: 20, height: 3, backgroundColor: '#4a7c59', borderRadius: 2, marginTop: 4 },

  // ═══ Card page (full screen) ═══
  cardPage: {
    backgroundColor: '#faf8f5',
    overflow: 'visible',
  },
  face: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: '#faf8f5',
    backfaceVisibility: 'hidden' as const,
  },

  // ═══ Floating capsule — domain + sub-domain + flip hint ═══
  capsuleWrapper: {
    position: 'absolute', left: 0, right: 0,
    alignItems: 'center', zIndex: 3,
  },
  floatingCapsule: {
    flexDirection: 'row', alignItems: 'center', gap: 5,
    paddingHorizontal: 12, paddingVertical: 5,
    borderRadius: 14,
    backgroundColor: 'rgba(255,255,255,0.88)',
    shadowColor: '#000', shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.06, shadowRadius: 6, elevation: 2,
  },
  capsuleDot: {
    width: 5, height: 5, borderRadius: 3, backgroundColor: '#4a7c59',
    marginRight: 2,
  },
  capsuleDomain: {fontSize: 11, fontWeight: '700', color: '#4a7c59'},
  capsuleSep: {fontSize: 11, color: '#b5b0a8', marginHorizontal: 1},
  capsuleSub: {fontSize: 11, fontWeight: '500', color: '#6b8a72'},
  capsuleBadge: {
    backgroundColor: '#4a7c59', width: 15, height: 15, borderRadius: 8,
    justifyContent: 'center', alignItems: 'center', marginLeft: 2,
  },
  capsuleBadgeText: {fontSize: 8, fontWeight: '800', color: '#fff'},
  capsuleDivider: {width: 1, height: 12, backgroundColor: '#e8e4db'},
  capsuleSource: {fontSize: 10, color: '#8a8a7a', fontWeight: '500'},

  // ═══ Front — content area ═══
  frontContent: {
    flex: 1,
    alignItems: 'center',
    paddingHorizontal: 32,
    paddingBottom: 100,
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
  titleText: {fontSize: 11, color: '#4a7c59', fontWeight: '500', marginTop: 2},

  // Stars
  starRow: {flexDirection: 'row', alignItems: 'center', gap: 4, marginTop: 8},
  stars: {fontSize: 16, color: '#e8a850', letterSpacing: 2},
  scoreText: {fontSize: 10, color: '#b5b0a8'},

  // ═══ Bottom actions (fixed) ═══
  bottomActions: {
    position: 'absolute', bottom: 28,
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

  // ═══ Flip hint (bottom right) ═══
  flipHintText: {fontSize: 10, color: '#c4c0b8', fontWeight: '400'},

  // ═══ Back face — interpretation ═══
  capsuleFlowWrapper: {
    alignItems: 'center',
    marginBottom: 24,
  },
  backQuoteArea: {
    paddingHorizontal: 32,
    alignItems: 'center',
  },
  backQuoteSmall: {
    fontSize: 14, fontWeight: '500', color: '#8a8a8a',
    lineHeight: 22, textAlign: 'center', fontStyle: 'italic',
  },

  backDivider: {
    width: 40, height: 2, backgroundColor: '#d4e0d6',
    alignSelf: 'center', marginVertical: 16,
  },

  backInterpArea: {
    flex: 1,
    paddingHorizontal: 32,
    paddingBottom: 24,
  },
  backInterpTitle: {
    fontSize: 17, fontWeight: '700', color: '#4a7c59',
    textAlign: 'center', marginBottom: 16,
  },

  interpText: {
    fontSize: 15, lineHeight: 26, color: '#3d3d3d',
    textAlign: 'left',
  },

});
