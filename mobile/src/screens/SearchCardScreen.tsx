import React, {useState, useEffect, useRef, useCallback} from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  Dimensions,
  Alert,
  StyleSheet,
} from 'react-native';
import {useSafeAreaInsets} from 'react-native-safe-area-context';
import {Experience, markInspired, setCollected, updateExperience, deleteExperience, recordView, recordExperienceEvent} from '../services/api';
import {getUserInfo} from '../services/config';
import FlipCard from '../components/ExperienceCard';
import {handleAuthExpired, requireLogin} from '../utils/authGate';
import Ionicons from '@expo/vector-icons/Ionicons';

const SCREEN_HEIGHT = Dimensions.get('window').height;
const HEADER_HEIGHT = 44;
const CARD_PADDING = 16;

function truncateKeyword(kw: string, max = 10): string {
  const runes = [...kw];
  return runes.length > max ? runes.slice(0, max).join('') + '...' : kw;
}

export default function SearchCardScreen({route, navigation}: any) {
  const {results, initialIndex, keyword} = route.params as {
    results: Experience[];
    initialIndex: number;
    keyword: string;
  };
  const insets = useSafeAreaInsets();
  const CARD_HEIGHT = SCREEN_HEIGHT - insets.top - HEADER_HEIGHT - CARD_PADDING * 2;
  const CONTENT_TOP = insets.top + HEADER_HEIGHT + 12;
  const flatListRef = useRef<FlatList>(null);

  const [cards, setCards] = useState<Experience[]>(results);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [flippedCards, setFlippedCards] = useState<Set<string>>(new Set());
  const [visibleCardId, setVisibleCardId] = useState<string | null>(null);

  useEffect(() => {
    getUserInfo().then(u => setCurrentUserId(u?.id || null));
  }, []);

  const updateCard = (id: string, patch: Partial<Experience>) => {
    setCards(prev => prev.map(e => e.id === id ? {...e, ...patch} : e));
  };

  const handleLike = async (id: string) => {
    if (!(await requireLogin(navigation, '登录后可以标记有启发，年糕也会更懂你的偏好。'))) return;
    const card = cards.find(c => c.id === id);
    if (!card || card.is_inspired) return;
    updateCard(id, {is_inspired: true, inspiration_count: card.inspiration_count + 1});
    try { await markInspired(id); } catch (e) {
      updateCard(id, {is_inspired: card.is_inspired, inspiration_count: card.inspiration_count});
      await handleAuthExpired(navigation, e);
    }
  };

  const handleBookmark = async (id: string) => {
    if (!(await requireLogin(navigation, '登录后可以收藏经验，之后在看看里随时翻回来。'))) return;
    const card = cards.find(c => c.id === id);
    if (!card) return;
    const nextCollected = !card.is_collected;
    updateCard(id, {
      is_collected: nextCollected,
      collection_count: Math.max(card.collection_count + (nextCollected ? 1 : -1), 0),
    });
    try { await setCollected(id, nextCollected); } catch (e) {
      updateCard(id, {is_collected: card.is_collected, collection_count: card.collection_count});
      await handleAuthExpired(navigation, e);
    }
  };

  const handleDelete = (id: string) => {
    const card = cards.find(e => e.id === id);
    const isPublic = card && card.visibility !== 'private';
    const turnPrivate = async () => {
      if (!card) return;
      try {
        await updateExperience(
          card.id,
          card.content,
          card.domain,
          card.sub_domain,
          true,
          card.interpretation,
          card.topics,
        );
        updateCard(card.id, {
          visibility: 'private',
          review_status: 'private',
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
            setCards(prev => prev.filter(e => e.id !== id));
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
      recordExperienceEvent(id, 'flip', 'search', {
        keyword,
        side: 'back',
      });
    }
  }, [keyword]);

  const viewabilityConfig = useRef({itemVisiblePercentThreshold: 50}).current;
  const onViewableItemsChanged = useRef(({viewableItems}: any) => {
    if (viewableItems.length > 0) {
      setVisibleCardId(viewableItems[0].item.id);
      recordView(viewableItems[0].item.id);
    }
  }).current;

  const displayKeyword = truncateKeyword(keyword);

  return (
    <View style={s.container}>
      {/* Full-screen FlatList */}
      <FlatList
        ref={flatListRef}
        data={cards}
        keyExtractor={item => item.id}
        initialScrollIndex={cards.length > initialIndex ? initialIndex : 0}
        getItemLayout={(_, index) => ({
          length: CARD_HEIGHT + CARD_PADDING * 2,
          offset: (CARD_HEIGHT + CARD_PADDING * 2) * index,
          index,
        })}
        snapToInterval={CARD_HEIGHT + CARD_PADDING * 2}
        snapToAlignment="start"
        disableIntervalMomentum
        showsVerticalScrollIndicator={false}
        decelerationRate="fast"
        onScrollToIndexFailed={(info) => {
          setTimeout(() => {
            flatListRef.current?.scrollToIndex({
              index: info.index, animated: false, viewPosition: 0,
            });
          }, 200);
        }}
        onViewableItemsChanged={onViewableItemsChanged}
        viewabilityConfig={viewabilityConfig}
        renderItem={({item}: {item: Experience}) => (
          <View style={{height: CARD_HEIGHT + CARD_PADDING * 2, padding: CARD_PADDING}}>
            <FlipCard
              item={item}
              cardHeight={CARD_HEIGHT}
              contentTop={CONTENT_TOP}
              currentUserId={currentUserId}
              onLike={handleLike}
              onBookmark={handleBookmark}
              onDelete={handleDelete}
              onFlipChange={handleFlipChange}
              showActions
            />
          </View>
        )}
        ListFooterComponent={
          <View style={{paddingVertical: 12, justifyContent: 'center', alignItems: 'center'}}>
            <Text style={s.footerHint}>— 共 {cards.length} 条 —</Text>
          </View>
        }
        ListEmptyComponent={
          <View style={{height: CARD_HEIGHT, justifyContent: 'center', alignItems: 'center'}}>
            <Text style={{fontSize: 15, color: '#9a9a9a'}}>没有相关经验</Text>
          </View>
        }
      />

      {/* Absolute-positioned transparent top bar */}
      <View style={[s.topBar, {top: insets.top}]}>
        <TouchableOpacity style={s.backBtn} onPress={() => navigation.goBack()} accessibilityRole="button" accessibilityLabel="返回">
          <Ionicons name="chevron-back" size={21} color="#5c5548" />
        </TouchableOpacity>
        <Text style={s.title} numberOfLines={1}>
          和'{displayKeyword}'相关的经验
        </Text>
        <Text style={s.count}>· {cards.length}条</Text>
      </View>

      {/* Flip hint overlay */}
      {(() => {
        const vExp = visibleCardId ? cards.find(e => e.id === visibleCardId) : null;
        if (!vExp?.interpretation) return null;
        return (
          <View style={s.flipHintOverlay} pointerEvents="none">
            <Text style={s.flipHintText}>
              {flippedCards.has(visibleCardId!) ? '点击回正面' : '点击卡片看解读'}
            </Text>
          </View>
        );
      })()}
    </View>
  );
}

const s = StyleSheet.create({
  container: {flex: 1, backgroundColor: '#faf8f5'},
  topBar: {
    position: 'absolute',
    left: 0, right: 0,
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingTop: 8,
    paddingBottom: 6,
    gap: 8,
    backgroundColor: 'transparent',
    zIndex: 10,
  },
  backBtn: {
    width: 36, height: 36,
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  title: {fontSize: 15, fontWeight: '600', color: '#2a2722', flexShrink: 1},
  count: {fontSize: 12, color: '#9b9487'},
  footerHint: {fontSize: 11, color: '#c5bfb3'},
  flipHintOverlay: {
    position: 'absolute',
    bottom: 40, right: 16,
    zIndex: 20,
  },
  flipHintText: {fontSize: 10, color: '#c4c0b8', fontWeight: '400'},
});
