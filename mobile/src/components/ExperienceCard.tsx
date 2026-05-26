import React, {useState, useRef} from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  Animated,
  StyleSheet,
} from 'react-native';
import Ionicons from '@expo/vector-icons/Ionicons';
import {Experience} from '../services/api';

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

  const isPlatform = item.experience_type === 'platform_selected';
  const primaryDomain = DOMAIN_LABELS[item.domain] || item.domain;
  const subDomain = item.sub_domain ? SUB_LABELS[item.sub_domain] : null;
  const displayName = item.creator_display_name || '匿名';
  const showScore = item.quality_score != null && item.quality_score > 0;
  const stars = showScore ? Math.round(item.quality_score! / 2) : 0;
  const ownerUserId = item.owner_user_id;
  const isOwner = Boolean(currentUserId && ownerUserId === currentUserId);
  const isUnavailable = Boolean(item.unavailable_reason);

  if (isUnavailable) {
    return (
      <View style={[styles.cardPage, {height: cardHeight}]}>
        <TouchableOpacity
          activeOpacity={1}
          onPress={() => {}}
          style={styles.face}
          testID={`experience-card-${item.id}`}
          accessibilityRole="button"
          accessibilityLabel="不可见经验卡片"
        >
          <View style={styles.unavailableContent}>
            <View style={styles.unavailableIconWrap}>
              <Ionicons name="alert-circle-outline" size={28} color="#7a806f" />
            </View>
            <Text style={styles.unavailableTitle}>该经验已不可见</Text>
            <Text style={styles.unavailableBody}>
              它可能已经被删除、转为私密，或正在重新处理。
            </Text>
          </View>
        </TouchableOpacity>
        {showActions && item.is_collected && (
          <View style={styles.bottomActions}>
            <TouchableOpacity
              style={[styles.actionBtn, styles.unavailableActionBtn, styles.actionSaved]}
              onPress={(e) => { e.stopPropagation(); onBookmark?.(item.id); }}
              accessibilityRole="button"
              accessibilityLabel="从收藏移除"
            >
              <View style={styles.actionContent}>
                <Ionicons name="star" size={15} color="#e8a850" />
                <Text style={[styles.actionText, styles.actionSavedText]}>从收藏移除</Text>
              </View>
            </TouchableOpacity>
          </View>
        )}
      </View>
    );
  }

  // 估算原文是否会导致底部按钮被挤出可视区
  const shouldShowOriginal = (() => {
    if (!item.original_text) return false;
    const contentLen = item.content?.length || 0;
    const originalLen = item.original_text.length;
    // 判断原文是否为中文（CJK字符占比 > 0.3）
    const cjkCount = [...item.original_text].filter(c => {
      const code = c.codePointAt(0)!;
      return (code >= 0x4E00 && code <= 0x9FFF) ||
             (code >= 0x3400 && code <= 0x4DBF) ||
             (code >= 0xF900 && code <= 0xFAFF);
    }).length;
    const isChinese = cjkCount / originalLen > 0.3;
    // 正文：26px字号 40px行高，~12中文字符/行
    const contentLines = Math.max(1, Math.ceil(contentLen / 12));
    const contentHeight = contentLines * 40;
    // 原文：13px字号 20px行高，Georgia斜体中文字符更宽~16字符/行，英文~35字符/行
    const charsPerLine = isChinese ? 16 : 35;
    const originalLines = Math.max(1, Math.ceil(originalLen / charsPerLine));
    const originalHeight = originalLines * 20 + 30; // +label padding
    // 固定占用：引号(72-20=52) + 胶囊区(contentTop+68) + 分隔线(26) + 作者(36) + 星级(~28) + 底部(60) + 安全余量(10)
    const fixedHeight = 52 + contentTop + 68 + 26 + 36 + (showScore ? 28 : 0) + 60 + 10;
    const available = cardHeight - fixedHeight - contentHeight;
    return originalHeight <= available;
  })();

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
      <TouchableOpacity
        activeOpacity={0.95}
        onPress={handleFlip}
        style={styles.face}
        testID={`experience-card-${item.id}`}
        accessibilityRole="button"
        accessibilityLabel={`经验卡片，${item.content}${item.interpretation ? '，点击查看解读' : ''}`}
      >
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
              <Text style={styles.capsuleSource}>{isPlatform ? '精选' : '原创'}</Text>
            </View>
          </View>
          <View style={[styles.frontContent, {paddingTop: contentTop + 68}]}>
            {showScore && (
              <View style={styles.experienceStarRow}>
                <Text style={styles.stars}>{'★'.repeat(stars)}{'☆'.repeat(5 - stars)}</Text>
              </View>
            )}
            <Text style={styles.quoteMark}>"</Text>
            <Text style={styles.content}>{item.content}</Text>
            <View style={styles.divider} />
            <View style={styles.creatorRow}>
              <View style={styles.avatar}><Text style={styles.avatarText}>{displayName.charAt(0)}</Text></View>
              <View>
                <View style={styles.creatorNameRow}>
                  <Ionicons name="person-circle-outline" size={16} color="#6f6f65" />
                  <Text style={styles.creatorName}>{displayName}</Text>
                </View>
                {item.source_label ? <Text style={styles.sourceLabel}>{item.source_label}</Text> : null}
                {item.author_title ? <Text style={styles.titleText}>{item.author_title}</Text> : null}
              </View>
            </View>
            {shouldShowOriginal ? (
              <View style={styles.originalCard}>
                <View style={styles.originalLabelRow}>
                  <Ionicons name="book-outline" size={12} color="#b5b0a8" />
                  <Text style={styles.originalLabel}>原文</Text>
                </View>
                <Text style={styles.originalText} numberOfLines={3}>{item.original_text}</Text>
              </View>
            ) : null}
          </View>
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
                <Text style={styles.capsuleSource}>{isPlatform ? '精选' : '原创'}</Text>
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
      {showActions && (
        <View style={styles.bottomActions}>
          <TouchableOpacity
            style={[styles.actionBtn, item.is_inspired && styles.actionLiked]}
            onPress={(e) => { e.stopPropagation(); onLike?.(item.id); }}
            accessibilityRole="button"
            accessibilityLabel="标记有启发"
          >
            <View style={styles.actionContent}>
              <Ionicons
                name="sparkles"
                size={15}
                color={item.is_inspired ? '#e85d5d' : '#9a9a9a'}
              />
              <Text style={[styles.actionText, item.is_inspired && styles.actionLikedText]}>
                {item.inspiration_count > 0 ? String(item.inspiration_count) : '有启发'}
              </Text>
            </View>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.actionBtn, item.is_collected && styles.actionSaved]}
            onPress={(e) => { e.stopPropagation(); onBookmark?.(item.id); }}
            accessibilityRole="button"
            accessibilityLabel={item.is_collected ? '取消收藏经验' : '收藏经验'}
          >
            <View style={styles.actionContent}>
              <Ionicons
                name={item.is_collected ? 'star' : 'star-outline'}
                size={15}
                color={item.is_collected ? '#e8a850' : '#9a9a9a'}
              />
              <Text style={[styles.actionText, item.is_collected && styles.actionSavedText]}>
                {item.is_collected ? '已收藏' : '收藏'}
              </Text>
            </View>
          </TouchableOpacity>
          {isOwner && (
            <TouchableOpacity
              style={styles.deleteBtn}
              onPress={(e) => { e.stopPropagation(); onDelete?.(item.id); }}
            >
              <Text style={styles.deleteText}>删除</Text>
            </TouchableOpacity>
          )}
        </View>
      )}
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
  capsuleDivider: {width: 1, height: 12, backgroundColor: '#e8e4db'},
  capsuleSource: {fontSize: 10, color: '#8a8a7a', fontWeight: '500'},
  frontContent: {flex: 1, alignItems: 'center', paddingHorizontal: 32, paddingBottom: 100},
  experienceStarRow: {marginBottom: 10},
  quoteMark: {fontSize: 72, color: '#4a7c59', opacity: 0.12, fontFamily: 'Georgia', lineHeight: 72, marginBottom: -20},
  content: {fontSize: 26, fontWeight: '700', lineHeight: 40, color: '#1a1a1a', textAlign: 'center'},
  divider: {width: 40, height: 2, backgroundColor: '#d4e0d6', marginVertical: 24},
  creatorRow: {flexDirection: 'row', alignItems: 'center', gap: 8},
  avatar: {
    width: 32, height: 32, borderRadius: 16, backgroundColor: '#eaf2e8',
    justifyContent: 'center', alignItems: 'center',
  },
  avatarText: {fontSize: 13, fontWeight: '700', color: '#4a7c59'},
  creatorNameRow: {flexDirection: 'row', alignItems: 'center', gap: 4},
  creatorName: {fontSize: 14, fontWeight: '600', color: '#4a4a4a'},
  sourceLabel: {fontSize: 11, color: '#9a9a9a', marginTop: 1},
  titleText: {fontSize: 11, color: '#4a7c59', fontWeight: '500', marginTop: 2},
  stars: {fontSize: 16, color: '#e8a850', letterSpacing: 2},
  originalCard: {
    marginTop: 16, marginHorizontal: 16,
    padding: 14, backgroundColor: '#f7f4ee',
    borderRadius: 10, borderWidth: 1, borderColor: '#e0dbd0', borderStyle: 'dashed',
    alignSelf: 'stretch',
  },
  originalLabelRow: {flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 4, marginBottom: 6},
  originalLabel: {fontSize: 10, fontWeight: '600', color: '#b5b0a8', letterSpacing: 1},
  originalText: {
    fontSize: 13, lineHeight: 20, color: '#6a6a5a',
    fontStyle: 'italic', textAlign: 'center',
    fontFamily: 'Georgia',
  },
  bottomActions: {position: 'absolute', bottom: -32, left: 0, right: 0, flexDirection: 'row', justifyContent: 'center', gap: 12, zIndex: 2},
  actionBtn: {paddingHorizontal: 16, paddingVertical: 8, borderRadius: 20, backgroundColor: '#fff', borderWidth: 0.5, borderColor: '#f0ece7'},
  actionContent: {flexDirection: 'row', alignItems: 'center', gap: 5},
  actionText: {fontSize: 13, color: '#9a9a9a', fontWeight: '500'},
  actionLiked: {borderColor: '#fce8e8', backgroundColor: '#fff5f5'},
  actionLikedText: {color: '#e85d5d'},
  actionSaved: {borderColor: '#fdf3e4', backgroundColor: '#fffcf5'},
  actionSavedText: {color: '#e8a850'},
  deleteBtn: {paddingHorizontal: 16, paddingVertical: 8, borderRadius: 20, backgroundColor: '#fff5f5', borderWidth: 0.5, borderColor: '#fce8e8'},
  deleteText: {fontSize: 13, color: '#e85d5d', fontWeight: '500'},
  unavailableContent: {flex: 1, alignItems: 'center', justifyContent: 'center', paddingHorizontal: 44, paddingBottom: 40},
  unavailableIconWrap: {
    width: 56, height: 56, borderRadius: 28, backgroundColor: '#edf0e7',
    alignItems: 'center', justifyContent: 'center', marginBottom: 18,
  },
  unavailableTitle: {fontSize: 22, lineHeight: 30, fontWeight: '700', color: '#252820', textAlign: 'center'},
  unavailableBody: {marginTop: 10, fontSize: 15, lineHeight: 23, color: '#74796b', textAlign: 'center'},
  unavailableActionBtn: {minHeight: 44, justifyContent: 'center'},
  capsuleFlowWrapper: {alignItems: 'center', marginBottom: 24},
  backQuoteArea: {paddingHorizontal: 32, alignItems: 'center'},
  backQuoteSmall: {fontSize: 14, fontWeight: '500', color: '#8a8a8a', lineHeight: 22, textAlign: 'center', fontStyle: 'italic'},
  backDivider: {width: 40, height: 2, backgroundColor: '#d4e0d6', alignSelf: 'center', marginVertical: 16},
  backInterpArea: {flex: 1, paddingHorizontal: 32, paddingBottom: 24},
  backInterpTitle: {fontSize: 17, fontWeight: '700', color: '#4a7c59', textAlign: 'center', marginBottom: 16},
  interpText: {fontSize: 15, lineHeight: 26, color: '#3d3d3d', textAlign: 'left'},
});
