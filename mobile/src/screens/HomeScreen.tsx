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
import Ionicons from '@expo/vector-icons/Ionicons';
import {
  fetchRecommendations,
  fetchMyExperiences,
  fetchMyBookmarks,
  Experience,
  markInspired,
  setCollected,
  updateExperience,
  deleteExperience,
  recordView,
  recordExperienceEvent,
} from '../services/api';
import {getToken, getUserInfo} from '../services/config';
import FlipCard from '../components/ExperienceCard';
import {handleAuthExpired, requireLogin} from '../utils/authGate';
import {reportHandledError} from '../utils/logging';

// Module-level tab refresh trigger (called from CreateScreen after publish)
let _pendingTabRefresh: string | null = null;
export function triggerTabRefresh(tab: string) { _pendingTabRefresh = tab; }

// Tab bar height (React Navigation bottom tab + safe area ≈ 80px)
const TAB_BAR_ESTIMATE = 80;
const TOP_TABS_HEIGHT = 44; // 方案A 顶部三标签栏
const TOP_HIT_HEIGHT = 58;
const TOP_HIT_HORIZONTAL_PADDING = 24;
const SEARCH_HIT_WIDTH = 64;
const SCREEN_HEIGHT = Dimensions.get('window').height;
const SCREEN_WIDTH = Dimensions.get('window').width;
const PAGE_SIZE = 20;

const DOMAIN_LABELS: Record<string, string> = {
  vitality: '生命', living: '生活', work: '工作',
  relationship: '关系', cognition: '认知', meaning: '意义',
};
const SUB_LABELS: Record<string, string> = {
  'health': '健康', 'housing': '居住', 'transit': '出行',
  'diet': '饮食', 'exercise': '运动',
  'pets': '宠物', 'travel': '旅行', 'fashion': '衣着',
  'selfcare': '养护', 'shopping': '购物', 'fun': '娱乐',
  'jobhunt': '求职', 'promotion': '升职', 'startup': '创业',
  'work-comm': '沟通', 'management': '管理', 'productivity': '效率',
  'marriage': '夫妻', 'romance': '恋人', 'friendship': '朋友',
  'parenting': '亲子', 'parents': '父母', 'siblings': '兄妹',
  'cog-learning': '学习', 'thinking': '思维', 'info': '信息',
  'tools': '工具', 'creativity': '创造', 'expression': '表达',
  'self': '自我', 'happiness': '幸福', 'emotion': '情绪', 'faith': '信仰',
  'mission': '使命', 'belonging': '归属',
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
  const tabOrder: TabName[] = ['recommend', 'bookmarks', 'my'];
  const [activeTab, setActiveTab] = useState<TabName>('recommend');

  type TabCache = { items: Experience[]; offset: number; nextCursor?: string; hasMore: boolean; loaded: boolean };
  const defaultCache = (): TabCache => ({ items: [], offset: 0, nextCursor: undefined, hasMore: true, loaded: false });
  const [tabCaches, setTabCaches] = useState<Record<TabName, TabCache>>({
    recommend: defaultCache(), my: defaultCache(), bookmarks: defaultCache(),
  });

  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loadMoreError, setLoadMoreError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPersonalized, setIsPersonalized] = useState(false);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [flippedCards, setFlippedCards] = useState<Set<string>>(new Set());
  const [visibleCardId, setVisibleCardId] = useState<string | null>(null);

  const loadingMoreRef = useRef(false);
  const tabScrollIndex = useRef<Record<TabName, number>>({ recommend: 0, my: 0, bookmarks: 0 });
  const flatListRef = useRef<FlatList>(null);
  const tokenRef = useRef<string | null>(null);
  const inspiringIdsRef = useRef<Set<string>>(new Set());
  const collectingIdsRef = useRef<Set<string>>(new Set());

  // Derived
  const experiences = tabCaches[activeTab].items;
  const hasMore = tabCaches[activeTab].hasMore;

  useEffect(() => {
    getToken().then(t => {
      tokenRef.current = t;
      loadInitial('recommend');
    });
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
          [tab]: defaultCache(),
        }));
      }
    }, [])
  );

  const loadPage = useCallback(async (cursorOrOffset: number | string, append: boolean, tab: TabName) => {
    const offset = typeof cursorOrOffset === 'number' ? cursorOrOffset : 0;
    let result;
    if (tab === 'bookmarks') {
      result = await fetchMyBookmarks(Math.floor(offset / PAGE_SIZE) + 1);
      setIsPersonalized(false);
    } else if (tab === 'my') {
      result = await fetchMyExperiences(Math.floor(offset / PAGE_SIZE) + 1);
      setIsPersonalized(false);
    } else {
      result = await fetchRecommendations(PAGE_SIZE, cursorOrOffset);
      setIsPersonalized(Boolean(tokenRef.current));
    }
    const data = Array.isArray(result?.data) ? result.data : [];
    const noMore = typeof result?.has_more === 'boolean' ? !result.has_more : data.length < PAGE_SIZE;

    setTabCaches(prev => {
      const cache = prev[tab];
      let items: Experience[];
      const nextCursor = tab === 'recommend' ? (result as {next_cursor?: string})?.next_cursor || undefined : undefined;
      if (append) {
        const ids = new Set(cache.items.map(e => e.id));
        items = [...cache.items, ...data.filter((e: Experience) => !ids.has(e.id))];
      } else {
        items = data;
      }
      return {
        ...prev,
        [tab]: {
          items,
          offset: offset + data.length,
          nextCursor,
          hasMore: !noMore,
          loaded: true,
        },
      };
    });
    setLoadMoreError(null);
    return data.length;
  }, []);

  const handleFeedAuthExpired = useCallback(async (e: any, fallbackTab?: TabName) => {
    if (!(await handleAuthExpired(navigation, e))) return false;
    tokenRef.current = null;
    setCurrentUserId(null);
    setError(null);
    setLoadMoreError(null);
    if (fallbackTab) setActiveTab(fallbackTab);
    return true;
  }, [navigation]);

  const loadInitial = useCallback(async (tab: TabName) => {
    try { await loadPage(0, false, tab); }
    catch (e) {
      if (await handleFeedAuthExpired(e)) return;
      reportHandledError('HomeScreen.loadInitial', e);
      setError('加载失败');
    }
    finally { setLoading(false); }
  }, [handleFeedAuthExpired, loadPage]);

  const handleLoadMore = useCallback(async () => {
    const cache = tabCaches[activeTab];
    if (loadingMoreRef.current || !cache.hasMore) return;
    loadingMoreRef.current = true;
    setLoadingMore(true);
    const cursorOrOffset = activeTab === 'recommend'
      ? (cache.nextCursor || cache.offset)
      : cache.offset;
    try { await loadPage(cursorOrOffset, true, activeTab); }
    catch (e) {
      if (await handleFeedAuthExpired(e, activeTab === 'recommend' ? undefined : 'recommend')) return;
      reportHandledError('HomeScreen.handleLoadMore', e);
      setLoadMoreError('网络不稳，点一下重试');
    }
    finally { loadingMoreRef.current = false; setLoadingMore(false); }
  }, [handleFeedAuthExpired, loadPage, activeTab, tabCaches]);

  const handleRefreshCurrentTab = () => {
    setError(null);
    setLoadMoreError(null);
    setTabCaches(prev => ({
      ...prev,
      [activeTab]: defaultCache(),
    }));
    setLoading(true);
    loadPage(0, false, activeTab)
      .catch(async e => {
        if (await handleFeedAuthExpired(e, activeTab === 'recommend' ? undefined : 'recommend')) return;
        reportHandledError('HomeScreen.handleRefreshCurrentTab', e);
        setError('加载失败');
      })
      .finally(() => setLoading(false));
  };

  const handleTabChange = async (tab: TabName) => {
    if (tab === activeTab) return;
    if ((tab === 'bookmarks' || tab === 'my') && !(await requireLogin(navigation, '登录后可以查看收藏和自己记下的经验。'))) {
      return;
    }
    // Save current scroll index
    const previousTab = activeTab;
    tabScrollIndex.current[activeTab] = tabScrollIndex.current[activeTab] || 0;
    setActiveTab(tab);
    setError(null);
    setLoadMoreError(null);
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
      loadPage(0, false, tab)
        .catch(async e => {
          if (await handleFeedAuthExpired(e, previousTab)) return;
          reportHandledError('HomeScreen.handleTabChange', e);
          setError('加载失败');
        })
        .finally(() => setLoading(false));
    }
  };

  const handleLike = async (id: string) => {
    if (inspiringIdsRef.current.has(id)) return;
    inspiringIdsRef.current.add(id);
    let optimisticApplied = false;
    try {
      if (!(await requireLogin(navigation, '登录后可以标记有启发，年糕也会更懂你的偏好。'))) return;
      const current = tabCaches[activeTab].items.find(e => e.id === id);
      if (!current || current.is_inspired) return;
      setTabCaches(prev => ({
        ...prev,
        [activeTab]: {
          ...prev[activeTab],
          items: prev[activeTab].items.map(e =>
            e.id === id ? {...e, is_inspired: true, inspiration_count: e.inspiration_count + 1} : e
          ),
        },
      }));
      optimisticApplied = true;
      await markInspired(id);
    } catch (e: any) {
      if (optimisticApplied) {
        setTabCaches(prev => ({
          ...prev,
          [activeTab]: {
            ...prev[activeTab],
            items: prev[activeTab].items.map(item =>
              item.id === id ? {...item, is_inspired: false, inspiration_count: Math.max(item.inspiration_count - 1, 0)} : item
            ),
          },
        }));
      }
      if (await handleAuthExpired(navigation, e)) return;
      Alert.alert('操作失败', e?.message || '请稍后再试');
    } finally {
      inspiringIdsRef.current.delete(id);
    }
  };

  const handleBookmark = async (id: string) => {
    if (collectingIdsRef.current.has(id)) return;
    collectingIdsRef.current.add(id);
    let optimisticApplied = false;
    let nextCollected = false;
    try {
      if (!(await requireLogin(navigation, '登录后可以收藏经验，之后在看看里随时翻回来。'))) return;
      const current = tabCaches[activeTab].items.find(e => e.id === id);
      if (!current) return;
      nextCollected = !current?.is_collected;
      setTabCaches(prev => ({
        ...prev,
        [activeTab]: {
          ...prev[activeTab],
          items: prev[activeTab].items.map(e => e.id === id ? {
            ...e,
            is_collected: nextCollected,
            collection_count: nextCollected ? e.collection_count + 1 : Math.max(e.collection_count - 1, 0),
          } : e),
        },
      }));
      optimisticApplied = true;
      await setCollected(id, nextCollected);
    } catch (e: any) {
      if (optimisticApplied) {
        setTabCaches(prev => ({
          ...prev,
          [activeTab]: {
            ...prev[activeTab],
            items: prev[activeTab].items.map(item => item.id === id ? {
              ...item,
              is_collected: !nextCollected,
              collection_count: nextCollected ? Math.max(item.collection_count - 1, 0) : item.collection_count + 1,
            } : item),
          },
        }));
      }
      if (await handleAuthExpired(navigation, e)) return;
      Alert.alert('操作失败', e?.message || '请稍后再试');
    } finally {
      collectingIdsRef.current.delete(id);
    }
  };

  const patchExperienceInCurrentTab = (id: string, patch: Partial<Experience>) => {
    setTabCaches(prev => ({
      ...prev,
      [activeTab]: {
        ...prev[activeTab],
        items: prev[activeTab].items.map(e => e.id === id ? {...e, ...patch} : e),
      },
    }));
  };

  const removeExperienceFromCurrentTab = (id: string) => {
    setTabCaches(prev => ({
      ...prev,
      [activeTab]: {
        ...prev[activeTab],
        items: prev[activeTab].items.filter(e => e.id !== id),
      },
    }));
  };

  const handleDelete = (id: string) => {
    const item = tabCaches[activeTab].items.find(e => e.id === id);
    const isPublic = item && item.visibility !== 'private';
    const turnPrivate = async () => {
      if (!item) return;
      try {
        await updateExperience(
          item.id,
          item.content,
          item.domain,
          item.sub_domain,
          true,
          item.interpretation,
          item.topic,
        );
        patchExperienceInCurrentTab(item.id, {
          visibility: 'private',
        });
      } catch (e: any) {
        if (await handleAuthExpired(navigation, e)) return;
        Alert.alert('操作失败', e?.message || '请稍后再试');
      }
    };
    Alert.alert(
      '删除经验',
      isPublic
        ? '删除后，他人收藏和历史引用会显示不可见。你也可以先转为私密，只停止公开展示、推荐和 AI 引用。'
        : '确定要删除这条经验吗？',
      [
        {text: '取消', style: 'cancel'},
        ...(isPublic ? [{text: '转为私密', onPress: turnPrivate}] : []),
        {text: '删除', style: 'destructive', onPress: async () => {
          try {
            await deleteExperience(id);
            removeExperienceFromCurrentTab(id);
          } catch (e: any) {
            if (await handleAuthExpired(navigation, e)) return;
            Alert.alert('删除失败', e?.message || '请稍后再试');
          }
        }},
      ],
    );
  };

  const handleFlipChange = useCallback((id: string, isFlipped: boolean) => {
    setFlippedCards(prev => {
      const next = new Set(prev);
      if (isFlipped) next.add(id); else next.delete(id);
      return next;
    });
    if (isFlipped) {
      recordExperienceEvent(id, 'flip', 'feed', {
        tab: activeTab,
        side: 'back',
      });
    }
  }, [activeTab]);

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
  // 左滑(dx<0): 推荐→收藏→我的→推荐（循环）
  // 右滑(dx>0): 推荐→我的→收藏→推荐（循环）
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
    if (Math.abs(dx) < 12 && Math.abs(dy) < 12) {
      const tapX = e.nativeEvent.pageX;
      const tapY = e.nativeEvent.pageY;
      const topStart = insets.top;
      const topEnd = topStart + TOP_HIT_HEIGHT;
      if (tapY >= topStart && tapY <= topEnd) {
        if (tapX >= SCREEN_WIDTH - SEARCH_HIT_WIDTH) {
          navigation.navigate('searchPage');
          return;
        }
        const tabsWidth = SCREEN_WIDTH - TOP_HIT_HORIZONTAL_PADDING * 2 - SEARCH_HIT_WIDTH;
        const localX = tapX - TOP_HIT_HORIZONTAL_PADDING;
        if (localX >= 0 && localX <= tabsWidth) {
          const idx = Math.min(tabOrder.length - 1, Math.max(0, Math.floor(localX / (tabsWidth / tabOrder.length))));
          handleTabChange(tabOrder[idx]);
          return;
        }
      }
    }
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
          <TouchableOpacity style={{backgroundColor:'#4a7c59',borderRadius:20,paddingHorizontal:24,paddingVertical:10}} onPress={handleRefreshCurrentTab}>
            <Text style={{color:'#fff',fontSize:14,fontWeight:'600'}}>重试</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  return (
    <View
      testID="home-screen"
      style={s.container}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      <FlatList
        testID="home-feed-list"
        style={s.feedList}
        onTouchStart={handleTouchStart}
        onTouchEnd={handleTouchEnd}
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
              {activeTab === 'recommend' ? '暂无推荐内容' : activeTab === 'bookmarks' ? '你还没有收藏经验' : '你还没有记下经验'}
            </Text>
            <Text style={{fontSize:12,color:'#b5b0a8',marginTop:6}}>
              {activeTab === 'recommend' ? '先随便看看也可以' : activeTab === 'bookmarks' ? '看到有用的经验可以先收藏' : '把最近想明白的一点记下来'}
            </Text>
          </View>
        }
        ListFooterComponent={
          loadingMore ? (
            <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5', justifyContent: 'center', alignItems: 'center'}}>
              <ActivityIndicator size="small" color="#4a7c59" />
            </View>
          ) : loadMoreError ? (
            <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5', justifyContent: 'center', alignItems: 'center'}}>
              <Text style={{fontSize:14,color:'#9a9a9a',marginBottom:12}}>{loadMoreError}</Text>
              <TouchableOpacity style={s.footerRetryBtn} onPress={handleLoadMore}>
                <Text style={s.footerRetryText}>重试</Text>
              </TouchableOpacity>
            </View>
          ) : !hasMore && experiences.length > 0 ? (
            <View style={{height: CARD_HEIGHT, backgroundColor: '#faf8f5', justifyContent: 'center', alignItems: 'center'}}>
              <Text style={{fontSize:14,color:'#b5b0a8',marginBottom:12}}>
                {activeTab === 'recommend' ? '这轮先看到这里' : '先看到这里'}
              </Text>
              {activeTab === 'recommend' ? (
                <TouchableOpacity style={s.footerRetryBtn} onPress={handleRefreshCurrentTab}>
                  <Text style={s.footerRetryText}>刷新</Text>
                </TouchableOpacity>
              ) : null}
            </View>
          ) : null
        }
      />
      {/* ═══ 方案A 顶部三标签 + 搜索图标 ═══ */}
      <View style={[s.tabBar, {top: insets.top}]}>
        <View style={s.tabBarInner}>
          {tabOrder.map(tab => (
            <TouchableOpacity
              key={tab}
              onPress={() => handleTabChange(tab)}
              style={s.tabItem}
              hitSlop={{top: 12, right: 10, bottom: 12, left: 10}}
              accessibilityRole="tab"
              accessibilityLabel={`${tab === 'recommend' ? '推荐' : tab === 'bookmarks' ? '收藏' : '我的'}分页`}
            >
              <Text style={[s.tabLabel, activeTab === tab && s.tabLabelActive]}>
                {tab === 'recommend' ? '推荐' : tab === 'bookmarks' ? '收藏' : '我的'}
              </Text>
              {activeTab === tab && <View style={s.tabUnderline} />}
            </TouchableOpacity>
          ))}
        </View>
        <TouchableOpacity
          onPress={() => navigation.navigate('searchPage')}
          style={s.searchIconBtn}
          accessibilityRole="button"
          accessibilityLabel="搜索经验"
          hitSlop={{top: 14, right: 14, bottom: 14, left: 14}}
        >
          <Ionicons name="search-outline" size={22} color="#1a1a1a" />
        </TouchableOpacity>
      </View>
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
  feedList: {zIndex: 0},

  // ═══ 方案A 顶部三标签 + 搜索 ═══
  tabBar: {
    position: 'absolute', left: 0, right: 0, zIndex: 100, elevation: 100,
    height: 58,
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    paddingHorizontal: 24, paddingTop: 8, paddingBottom: 6,
    backgroundColor: 'rgba(250,248,245,0.01)',
  },
  tabBarInner: {
    flexDirection: 'row', justifyContent: 'space-around', flex: 1,
  },
  searchIconBtn: {
    width: 36, height: 36,
    justifyContent: 'center', alignItems: 'center',
    marginLeft: 4,
  },
  tabItem: { minHeight: 36, justifyContent: 'center', alignItems: 'center', paddingHorizontal: 8 },
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
  footerRetryBtn: {
    borderRadius: 14,
    borderWidth: 1,
    borderColor: '#d9e4d5',
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: '#fff',
  },
  footerRetryText: {fontSize: 13, color: '#4a7c59', fontWeight: '700'},

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
