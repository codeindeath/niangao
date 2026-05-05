import React, {useState, useEffect, useCallback, useRef} from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  ScrollView,
  StyleSheet,
  ActivityIndicator,
  Dimensions,
  Alert,
} from 'react-native';
import {
  fetchRecommendations,
  fetchExperiences,
  Experience,
  toggleLike,
  toggleBookmark,
  deleteExperience,
} from '../services/api';
import {getToken, getUserInfo} from '../services/config';

const {height: SCREEN_HEIGHT} = Dimensions.get('window');
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

export default function HomeScreen({navigation: _navigation}: any) {
  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isPersonalized, setIsPersonalized] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [openInterps, setOpenInterps] = useState<Set<string>>(new Set());

  const loadingMoreRef = useRef(false);
  const hasMoreRef = useRef(true);
  const tokenRef = useRef<string | null>(null);
  const offsetRef = useRef(0);
  const listRef = useRef<FlatList>(null);

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
      setActiveIndex(0);
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

  const onViewableItemsChanged = useRef(({viewableItems}: any) => {
    if (viewableItems.length > 0) {
      setActiveIndex(viewableItems[0].index || 0);
    }
  }).current;

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

  const renderCard = ({item, index}: {item: Experience; index: number}) => {
    const isPlatform = item.source_type === 'platform';
    const domainLabel = SUB_LABELS[item.sub_domain] || DOMAIN_LABELS[item.domain] || item.domain;
    const displayName = item.creator_name || item.author_name || '匿名';
    const showScore = item.quality_score != null && item.quality_score > 0;
    const stars = showScore ? Math.round(item.quality_score! / 2) : 0;

    return (
      <View style={s.cardPage}>
        {/* Top tags */}
        <View style={s.topRow}>
          <View style={s.tag}><Text style={s.tagText}>{domainLabel}</Text></View>
          {isPlatform && <View style={s.platformTag}><Text style={s.platformTagText}>官</Text></View>}
        </View>

        {/* Page counter */}
        <Text style={s.counter}>{index + 1}/{experiences.length}{loadingMore ? '…' : ''}</Text>

        {/* Scrollable content area */}
        <ScrollView
          style={s.scrollArea}
          contentContainerStyle={s.scrollContent}
          showsVerticalScrollIndicator={false}
          bounces={false}
        >
          <Text style={s.quoteMark}>{'"'}</Text>
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

          {/* Interpretation toggle */}
          {item.interpretation ? (
            <>
              <TouchableOpacity onPress={() => {
                setOpenInterps(prev => {
                  const next = new Set(prev);
                  next.has(item.id) ? next.delete(item.id) : next.add(item.id);
                  return next;
                });
              }}>
                <Text style={s.interpToggle}>
                  {openInterps.has(item.id) ? '▾' : '▸'} 查看解读
                </Text>
              </TouchableOpacity>
              {openInterps.has(item.id) && (
                <Text style={s.interpBody}>{item.interpretation}</Text>
              )}
            </>
          ) : null}

          {/* Bottom spacer so content doesn't hide behind action bar */}
          <View style={{height: 80}} />
        </ScrollView>

        {/* Bottom actions */}
        <View style={s.bottomActions}>
          <TouchableOpacity
            style={[s.actionBtn, item.is_liked && s.actionLiked]}
            onPress={() => handleLike(item.id)}
          >
            <Text style={[s.actionText, item.is_liked && s.actionLikedText]}>
              ♥ {item.like_count > 0 ? item.like_count : '点赞'}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[s.actionBtn, item.is_bookmarked && s.actionSaved]}
            onPress={() => handleBookmark(item.id)}
          >
            <Text style={[s.actionText, item.is_bookmarked && s.actionSavedText]}>
              ★ {item.is_bookmarked ? '已收藏' : '收藏'}
            </Text>
          </TouchableOpacity>
          {currentUserId && item.author_id === currentUserId && (
            <TouchableOpacity style={s.deleteBtn} onPress={() => handleDelete(item.id)}>
              <Text style={s.deleteText}>删除</Text>
            </TouchableOpacity>
          )}
        </View>

        {/* Page counter */}
        <Text style={s.counter}>{index + 1}/{experiences.length}{loadingMore ? '…' : ''}</Text>
      </View>
    );
  };

  return (
    <View style={s.container}>
      <FlatList
        data={experiences}
        keyExtractor={item => item.id}
        renderItem={renderCard}
        pagingEnabled
        showsVerticalScrollIndicator={false}
        decelerationRate="fast"
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.5}
        onViewableItemsChanged={onViewableItemsChanged}
        viewabilityConfig={{itemVisiblePercentThreshold: 50}}
        ListEmptyComponent={
          <View style={s.cardPage}>
            <View style={{flex:1,justifyContent:'center',alignItems:'center'}}>
              <Text style={{fontSize:15,color:'#9a9a9a'}}>暂无推荐内容</Text>
              <Text style={{fontSize:12,color:'#b5b0a8',marginTop:6}}>发布经验后，推荐会更精准</Text>
            </View>
          </View>
        }
        ListFooterComponent={
          loadingMore ? (
            <View style={[s.cardPage, {justifyContent:'center',alignItems:'center'}]}>
              <ActivityIndicator size="small" color="#4a7c59" />
            </View>
          ) : !hasMore && experiences.length > 0 ? (
            <View style={[s.cardPage, {justifyContent:'center',alignItems:'center'}]}>
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
  cardPage: {
    height: SCREEN_HEIGHT,
    backgroundColor: '#faf8f5',
  },

  // Scrollable area
  scrollArea: {
    flex: 1,
    marginTop: 130,
    marginBottom: 100,
  },
  scrollContent: {
    paddingHorizontal: 32,
    alignItems: 'center',
  },

  // Top
  topRow: {
    position: 'absolute', top: 60, left: 24,
    flexDirection: 'row', gap: 6,
    zIndex: 1,
  },
  tag: {backgroundColor: '#eaf2e8', paddingHorizontal: 10, paddingVertical: 4, borderRadius: 12},
  tagText: {fontSize: 11, fontWeight: '600', color: '#4a7c59'},
  platformTag: {
    backgroundColor: '#4a7c59', width: 18, height: 18, borderRadius: 9,
    justifyContent: 'center', alignItems: 'center',
  },
  platformTagText: {fontSize: 9, fontWeight: '800', color: '#fff'},

  // Content
  quoteMark: {
    fontSize: 72, color: '#4a7c59', opacity: 0.12,
    fontFamily: 'Georgia', lineHeight: 72, marginBottom: -20,
  },
  content: {
    fontSize: 26, fontWeight: '700', lineHeight: 40,
    color: '#1a1a1a', textAlign: 'center',
  },
  divider: {width: 40, height: 2, backgroundColor: '#d4e0d6', marginVertical: 28},

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
  starRow: {flexDirection: 'row', alignItems: 'center', gap: 4, marginTop: 10},
  stars: {fontSize: 16, color: '#e8a850', letterSpacing: 2},
  scoreText: {fontSize: 10, color: '#b5b0a8'},

  // Interpretation
  interpToggle: {
    fontSize: 13, color: '#b5b0a8', marginTop: 14, textAlign: 'center',
  },
  interpBody: {
    fontSize: 13, lineHeight: 22, color: '#6e6e6e',
    textAlign: 'center', marginTop: 8, paddingHorizontal: 10,
  },

  // Bottom actions
  bottomActions: {
    position: 'absolute', bottom: 50,
    left: 0, right: 0,
    flexDirection: 'row', justifyContent: 'center', gap: 12,
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

  // Counter
  counter: {
    position: 'absolute', top: 60, right: 22,
    fontSize: 13, fontWeight: '600', color: '#d4d0c8',
  },
});
