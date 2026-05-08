import React, {useState, useRef} from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  Animated,
  StyleSheet,
} from 'react-native';
import {Experience} from '../services/api';

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

interface FlipCardProps {
  item: Experience;
  cardHeight: number;
  contentTop?: number;
  currentUserId?: string | null;
  onLike?: (id: string) => void;
  onBookmark?: (id: string) => void;
  onDelete?: (id: string) => void;
  onFlipChange?: (id: string, isFlipped: boolean) => void;
  showActions?: boolean;
}

export default function FlipCard({
  item, cardHeight, contentTop = 0,
  currentUserId, onLike, onBookmark, onDelete, onFlipChange,
  showActions = false,
}: FlipCardProps) {
  const flipAnim = useRef(new Animated.Value(0)).current;
  const [isFlipped, setIsFlipped] = useState(false);
  const MAX_INTERP_LINES = 15;

  const isPlatform = item.source_type === 'platform';
  const primaryDomain = DOMAIN_LABELS[item.domain] || item.domain;
  const subDomain = item.sub_domain ? SUB_LABELS[item.sub_domain] : null;
  const displayName = item.creator_name || item.author_name || '匿名';
  const showScore = item.quality_score != null && item.quality_score > 0;
  const stars = showScore ? Math.round(item.quality_score! / 2) : 0;

  const handleFlip = () => {
    if (!item.interpretation) return;
    const toValue = isFlipped ? 0 : 1;
    Animated.spring(flipAnim, {
      toValue,
      friction: 8,
      tension: 60,
      useNativeDriver: true,
    }).start();
    const newFlipped = !isFlipped;
    setIsFlipped(newFlipped);
    onFlipChange?.(item.id, newFlipped);
  };

  const frontInterpolate = flipAnim.interpolate({
    inputRange: [0, 1], outputRange: ['0deg', '180deg'],
  });
  const backInterpolate = flipAnim.interpolate({
    inputRange: [0, 1], outputRange: ['180deg', '360deg'],
  });
  const frontOpacity = flipAnim.interpolate({
    inputRange: [0, 0.5, 1], outputRange: [1, 0, 0],
  });
  const backOpacity = flipAnim.interpolate({
    inputRange: [0, 0.5, 1], outputRange: [0, 0, 1],
  });

  return (
    <View style={[styles.cardPage, {height: cardHeight}]}>
      <TouchableOpacity activeOpacity={0.95} onPress={handleFlip} style={styles.face}>
        {/* ═══ 正面 ═══ */}
        <Animated.View
          style={[
            styles.face,
            {transform: [{perspective: 1000}, {rotateY: frontInterpolate}], opacity: frontOpacity},
          ]}
          pointerEvents={isFlipped ? 'none' : 'auto'}
        >
          <View style={[styles.capsuleWrapper, {top: contentTop}]}>
            <View style={styles.floatingCapsule}>
              <View style={styles.capsuleDot} />
              <Text style={styles.capsuleDomain}>{primaryDomain}</Text>
              {subDomain ? <><Text style={styles.capsuleSep}>·</Text><Text style={styles.capsuleSub}>{subDomain}</Text></> : null}
              <View style={styles.capsuleDivider} />
              <Text style={styles.capsuleSource}>{isPlatform ? '来自年糕' : '用户原创'}</Text>
              {isPlatform && <View style={styles.capsuleBadge}><Text style={styles.capsuleBadgeText}>官</Text></View>}
            </View>
          </View>
          <View style={[styles.frontContent, {paddingTop: contentTop + 68}]}>
            <Text style={styles.quoteMark}>"</Text>
            <Text style={styles.content}>{item.content}</Text>
            <View style={styles.divider} />
            <View style={styles.creatorRow}>
              <View style={styles.avatar}><Text style={styles.avatarText}>{displayName.charAt(0)}</Text></View>
              <View>
                <Text style={styles.creatorName}>{displayName}</Text>
                {item.source_label ? <Text style={styles.sourceLabel}>{item.source_label}</Text> : null}
                {item.author_title ? <Text style={styles.titleText}>{item.author_title}</Text> : null}
              </View>
            </View>
            {showScore && (
              <View style={styles.starRow}>
                <Text style={styles.stars}>{'★'.repeat(stars)}{'☆'.repeat(5 - stars)}</Text>
                {item.score_reason ? <Text style={styles.scoreText}>{item.score_reason}</Text> : null}
              </View>
            )}
            {item.original_text ? (
              <View style={styles.originalCard}>
                <Text style={styles.originalLabel}>📖 原文</Text>
                <Text style={styles.originalText}>{item.original_text}</Text>
              </View>
            ) : null}
          </View>
          {showActions && (
            <View style={styles.bottomActions}>
              <TouchableOpacity
                style={[styles.actionBtn, item.is_liked && styles.actionLiked]}
                onPress={(e) => { e.stopPropagation(); onLike?.(item.id); }}
              >
                <Text style={[styles.actionText, item.is_liked && styles.actionLikedText]}>
                  ♥ {item.like_count > 0 ? item.like_count : '点赞'}
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.actionBtn, item.is_bookmarked && styles.actionSaved]}
                onPress={(e) => { e.stopPropagation(); onBookmark?.(item.id); }}
              >
                <Text style={[styles.actionText, item.is_bookmarked && styles.actionSavedText]}>
                  ★ {item.is_bookmarked ? '已收藏' : '收藏'}
                </Text>
              </TouchableOpacity>
              {currentUserId && item.author_id === currentUserId && (
                <TouchableOpacity
                  style={styles.deleteBtn}
                  onPress={(e) => { e.stopPropagation(); onDelete?.(item.id); }}
                >
                  <Text style={styles.deleteText}>删除</Text>
                </TouchableOpacity>
              )}
            </View>
          )}
        </Animated.View>

        {/* ═══ 背面 ═══ */}
        <Animated.View
          style={[
            styles.face,
            {transform: [{perspective: 1000}, {rotateY: backInterpolate}], opacity: backOpacity},
          ]}
          pointerEvents={isFlipped ? 'auto' : 'none'}
        >
          <TouchableOpacity onPress={handleFlip} activeOpacity={0.8}>
            <View style={[styles.capsuleFlowWrapper, {paddingTop: contentTop}]}>
              <View style={styles.floatingCapsule}>
                <View style={styles.capsuleDot} />
                <Text style={styles.capsuleDomain}>{primaryDomain}</Text>
                {subDomain ? <><Text style={styles.capsuleSep}>·</Text><Text style={styles.capsuleSub}>{subDomain}</Text></> : null}
                <View style={styles.capsuleDivider} />
                <Text style={styles.capsuleSource}>{isPlatform ? '来自年糕' : '用户原创'}</Text>
              </View>
            </View>
            <View style={styles.backQuoteArea}>
              <Text style={styles.backQuoteSmall}>"{item.content}"</Text>
            </View>
            <View style={styles.backDivider} />
          </TouchableOpacity>
          {item.interpretation ? (
            <View style={styles.backInterpArea}>
              <Text style={styles.backInterpTitle}>经验解读</Text>
              <Text style={styles.interpText} numberOfLines={MAX_INTERP_LINES}>{item.interpretation}</Text>
            </View>
          ) : (
            <View style={styles.backInterpArea}>
              <Text style={styles.backInterpTitle}>经验解读</Text>
              <Text style={[styles.interpText, {color: '#b5b0a8', textAlign: 'center'}]}>暂无解读</Text>
            </View>
          )}
        </Animated.View>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  cardPage: {backgroundColor: '#faf8f5', overflow: 'visible'},
  face: {...StyleSheet.absoluteFillObject, backgroundColor: '#faf8f5', backfaceVisibility: 'hidden' as const},
  capsuleWrapper: {position: 'absolute', left: 0, right: 0, alignItems: 'center', zIndex: 3},
  floatingCapsule: {
    flexDirection: 'row', alignItems: 'center', gap: 5,
    paddingHorizontal: 12, paddingVertical: 5, borderRadius: 14,
    backgroundColor: 'rgba(255,255,255,0.88)',
    shadowColor: '#000', shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.06, shadowRadius: 6, elevation: 2,
  },
  capsuleDot: {width: 5, height: 5, borderRadius: 3, backgroundColor: '#4a7c59', marginRight: 2},
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
  frontContent: {flex: 1, alignItems: 'center', paddingHorizontal: 32, paddingBottom: 100},
  quoteMark: {fontSize: 72, color: '#4a7c59', opacity: 0.12, fontFamily: 'Georgia', lineHeight: 72, marginBottom: -20},
  content: {fontSize: 26, fontWeight: '700', lineHeight: 40, color: '#1a1a1a', textAlign: 'center'},
  divider: {width: 40, height: 2, backgroundColor: '#d4e0d6', marginVertical: 24},
  creatorRow: {flexDirection: 'row', alignItems: 'center', gap: 8},
  avatar: {
    width: 32, height: 32, borderRadius: 16, backgroundColor: '#eaf2e8',
    justifyContent: 'center', alignItems: 'center',
  },
  avatarText: {fontSize: 13, fontWeight: '700', color: '#4a7c59'},
  creatorName: {fontSize: 14, fontWeight: '600', color: '#4a4a4a'},
  sourceLabel: {fontSize: 11, color: '#9a9a9a', marginTop: 1},
  titleText: {fontSize: 11, color: '#4a7c59', fontWeight: '500', marginTop: 2},
  starRow: {flexDirection: 'row', alignItems: 'center', gap: 4, marginTop: 8},
  stars: {fontSize: 16, color: '#e8a850', letterSpacing: 2},
  scoreText: {fontSize: 10, color: '#b5b0a8'},
  originalCard: {
    marginTop: 16, marginHorizontal: 16,
    padding: 14, backgroundColor: '#f7f4ee',
    borderRadius: 10, borderWidth: 1, borderColor: '#e0dbd0', borderStyle: 'dashed',
    alignSelf: 'stretch',
  },
  originalLabel: {fontSize: 10, fontWeight: '600', color: '#b5b0a8', letterSpacing: 1, marginBottom: 6},
  originalText: {
    fontSize: 13, lineHeight: 20, color: '#6a6a5a',
    fontStyle: 'italic', textAlign: 'center',
    fontFamily: 'Georgia',
  },
  bottomActions: {position: 'absolute', bottom: 28, left: 0, right: 0, flexDirection: 'row', justifyContent: 'center', gap: 12, zIndex: 2},
  actionBtn: {paddingHorizontal: 16, paddingVertical: 8, borderRadius: 20, backgroundColor: '#fff', borderWidth: 0.5, borderColor: '#f0ece7'},
  actionText: {fontSize: 13, color: '#9a9a9a', fontWeight: '500'},
  actionLiked: {borderColor: '#fce8e8', backgroundColor: '#fff5f5'},
  actionLikedText: {color: '#e85d5d'},
  actionSaved: {borderColor: '#fdf3e4', backgroundColor: '#fffcf5'},
  actionSavedText: {color: '#e8a850'},
  deleteBtn: {paddingHorizontal: 16, paddingVertical: 8, borderRadius: 20, backgroundColor: '#fff5f5', borderWidth: 0.5, borderColor: '#fce8e8'},
  deleteText: {fontSize: 13, color: '#e85d5d', fontWeight: '500'},
  capsuleFlowWrapper: {alignItems: 'center', marginBottom: 24},
  backQuoteArea: {paddingHorizontal: 32, alignItems: 'center'},
  backQuoteSmall: {fontSize: 14, fontWeight: '500', color: '#8a8a8a', lineHeight: 22, textAlign: 'center', fontStyle: 'italic'},
  backDivider: {width: 40, height: 2, backgroundColor: '#d4e0d6', alignSelf: 'center', marginVertical: 16},
  backInterpArea: {flex: 1, paddingHorizontal: 32, paddingBottom: 24},
  backInterpTitle: {fontSize: 17, fontWeight: '700', color: '#4a7c59', textAlign: 'center', marginBottom: 16},
  interpText: {fontSize: 15, lineHeight: 26, color: '#3d3d3d', textAlign: 'left'},
});
