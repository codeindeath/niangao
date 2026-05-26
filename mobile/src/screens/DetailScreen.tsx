import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  Modal,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {fetchExperience, markInspired, setCollected, updateExperience, deleteExperience, Experience} from '../services/api';
import {getUserInfo} from '../services/config';
import {handleAuthExpired, requireLogin} from '../utils/authGate';
import Ionicons from '@expo/vector-icons/Ionicons';

const DOMAIN_LABELS: Record<string, string> = {
  vitality: '生命', living: '生活', work: '工作',
  relationship: '关系', cognition: '认知', meaning: '意义',
};
const SUB_DOMAIN_LABELS: Record<string, string> = {
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

export default function DetailScreen({route, navigation}: any) {
  const {id} = route.params;
  const [exp, setExp] = useState<Experience | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [showScoreReason, setShowScoreReason] = useState(false);

  useEffect(() => {
    getUserInfo().then(u => setCurrentUserId(u?.id || null));
  }, []);

  useEffect(() => { loadExperience(); }, [id]);

  const loadExperience = async () => {
    try {
      const data = await fetchExperience(id);
      setExp(data);
      setError(null);
    } catch (e) {
      setError('加载失败，请检查网络连接');
    } finally {
      setLoading(false);
    }
  };

  const handleLike = async () => {
    if (!(await requireLogin(navigation, '登录后可以标记有启发，年糕也会更懂你的偏好。'))) return;
    if (!exp || exp.is_inspired) return;
    const previous = exp;
    setExp({...exp, is_inspired: true, inspiration_count: exp.inspiration_count + 1});
    try { await markInspired(exp.id); } catch (e) {
      setExp(previous);
      await handleAuthExpired(navigation, e);
    }
  };

  const handleBookmark = async () => {
    if (!(await requireLogin(navigation, '登录后可以收藏经验，之后在看看里随时翻回来。'))) return;
    if (!exp) return;
    const previous = exp;
    const nextCollected = !exp.is_collected;
    setExp({
      ...exp,
      is_collected: nextCollected,
      collection_count: Math.max(exp.collection_count + (nextCollected ? 1 : -1), 0),
    });
    try { await setCollected(exp.id, nextCollected); } catch (e) {
      setExp(previous);
      await handleAuthExpired(navigation, e);
    }
  };

  const performDelete = async () => {
    if (!exp) return;
    try { await deleteExperience(exp.id); navigation.goBack(); }
    catch (e: any) {
      if (await handleAuthExpired(navigation, e)) return;
      Alert.alert('删除失败', e?.message || '请稍后再试');
    }
  };

  const performMakePrivate = async () => {
    if (!exp) return;
    try {
      await updateExperience(
        exp.id,
        exp.content,
        exp.domain,
        exp.sub_domain,
        true,
        exp.interpretation,
        exp.topics,
      );
      setExp(prev => prev ? {
        ...prev,
        is_private: true,
        visibility: 'private',
        review_status: 'private',
      } : prev);
    } catch (e: any) {
      if (await handleAuthExpired(navigation, e)) return;
      Alert.alert('操作失败', e?.message || '请稍后再试');
    }
  };

  const handleDelete = () => {
    if (!currentUserId) {
      requireLogin(navigation, '登录后可以管理自己记下的经验。');
      return;
    }
    const isPublic = exp && !exp.is_private && exp.visibility !== 'private';
    Alert.alert(
      '删除经验',
      isPublic
        ? '删除后，他人收藏和历史引用会显示不可见。你也可以先转为私密，只停止公开展示、推荐和 AI 引用。'
        : '确定要删除这条经验吗？',
      [
        {text: '取消', style: 'cancel'},
        ...(isPublic ? [{text: '转为私密', onPress: performMakePrivate}] : []),
        {text: '删除', style: 'destructive', onPress: performDelete},
      ],
    );
  };

  const handleEdit = () => {
    if (!exp) return;
    if (!currentUserId) {
      requireLogin(navigation, '登录后可以编辑自己记下的经验。');
      return;
    }
    navigation.navigate('createEdit', {experience: exp});
  };

  const handleMakePrivate = () => {
    if (!exp || exp.is_private || exp.visibility === 'private') return;
    if (!currentUserId) {
      requireLogin(navigation, '登录后可以管理自己记下的经验。');
      return;
    }
    Alert.alert('转为私密', '转为私密后，这条经验会停止公开展示、推荐和 AI 引用。确定继续吗？', [
      {text: '取消', style: 'cancel'},
      {text: '转为私密', style: 'destructive', onPress: performMakePrivate},
    ]);
  };

  if (loading) {
    return <SafeAreaView style={s.container}><ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 200}} /></SafeAreaView>;
  }
  if (error && !exp) {
    return <SafeAreaView style={s.container}><View style={s.center}><Text style={s.errorText}>{error}</Text>
      <TouchableOpacity style={s.retryBtn} onPress={() => { setError(null); setLoading(true); loadExperience(); }}>
        <Text style={s.retryText}>重试</Text></TouchableOpacity></View></SafeAreaView>;
  }
  if (!exp) {
    return <SafeAreaView style={s.container}><View style={s.center}><Text style={s.emptyText}>经验不存在或已被删除</Text></View></SafeAreaView>;
  }

  const isPlatform = exp.experience_type === 'platform_selected';
  const displayName = exp.creator_display_name || exp.creator_name || exp.author_name || '匿名';
  const domainLabel = (exp.sub_domain && SUB_DOMAIN_LABELS[exp.sub_domain]) || DOMAIN_LABELS[exp.domain] || exp.domain;
  const showScore = exp.quality_score != null && exp.quality_score > 0;
  const stars = showScore ? Math.round(exp.quality_score! / 2) : 0;
  const ownerUserId = exp.owner_user_id || exp.author_id;
  const isOwner = Boolean(currentUserId && ownerUserId === currentUserId);

  return (
    <SafeAreaView style={s.container} edges={['top']}>
      <View style={s.header}>
        <TouchableOpacity style={s.backButton} onPress={() => navigation.goBack()} accessibilityRole="button" accessibilityLabel="返回">
          <Ionicons name="chevron-back" size={19} color="#4a7c59" />
          <Text style={s.backText}>返回</Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={s.body} contentContainerStyle={{paddingBottom: 40}}>
        {/* Creator row */}
        <View style={s.creatorRow}>
          <View style={s.avatar}><Text style={s.avatarText}>{displayName.charAt(0)}</Text></View>
          <Text style={s.creatorName}>{displayName}</Text>
          {isPlatform && (
            <View style={s.platformTag}>
              <Text style={s.platformTagText}>精选</Text>
            </View>
          )}
          <View style={s.domainTag}><Text style={s.domainTagText}>{domainLabel}</Text></View>
          {exp.is_private && (
            <View style={s.privateMark} accessible accessibilityLabel="私密经验">
              <Ionicons name="lock-closed-outline" size={14} color="#8a8173" />
            </View>
          )}
        </View>

        {/* Content card */}
        <View style={s.contentCard}>
          <Text style={s.content}>{exp.content}</Text>
          {exp.source_label && <Text style={s.source}>来源：{exp.source_label}</Text>}
          <Text style={s.date}>{new Date(exp.created_at).toLocaleDateString('zh-CN', {year: 'numeric', month: 'long', day: 'numeric'})}</Text>
        </View>

        {/* Score stars — click for reason */}
        {showScore && (
          <View style={s.scoreCard}>
            <Text style={s.scoreTitle}>价值度</Text>
            <TouchableOpacity
              style={s.scoreStarsRow}
              onPress={() => exp.score_reason && setShowScoreReason(true)}
              activeOpacity={exp.score_reason ? 0.6 : 1}
            >
              {[1, 2, 3, 4, 5].map(i => (
                <Text key={i} style={s.star}>{i <= stars ? '★' : '☆'}</Text>
              ))}
            </TouchableOpacity>
            {exp.score_reason && (
              <Modal visible={showScoreReason} transparent animationType="fade">
                <TouchableOpacity style={s.scoreOverlay} activeOpacity={1} onPress={() => setShowScoreReason(false)}>
                  <View style={s.scorePopup}>
                    <Text style={s.scorePopupTitle}>价值度 · {stars}星</Text>
                    <Text style={s.scorePopupText}>{exp.score_reason}</Text>
                  </View>
                </TouchableOpacity>
              </Modal>
            )}
          </View>
        )}

        {/* Actions */}
        <View style={s.actions}>
          <TouchableOpacity
            style={s.actionBtn}
            onPress={handleLike}
            accessibilityRole="button"
            accessibilityLabel="标记有启发"
          >
            <View style={s.actionContent}>
              <Ionicons
                name="sparkles"
                size={15}
                color={exp.is_inspired ? '#e85d5d' : '#6e6e6e'}
              />
              <Text style={[s.actionText, exp.is_inspired && s.actionLiked]}>
                {exp.inspiration_count > 0 ? String(exp.inspiration_count) : '有启发'}
              </Text>
            </View>
          </TouchableOpacity>
          <TouchableOpacity
            style={s.actionBtn}
            onPress={handleBookmark}
            accessibilityRole="button"
            accessibilityLabel={exp.is_collected ? '取消收藏经验' : '收藏经验'}
          >
            <View style={s.actionContent}>
              <Ionicons
                name={exp.is_collected ? 'bookmark' : 'bookmark-outline'}
                size={15}
                color={exp.is_collected ? '#e8a850' : '#6e6e6e'}
              />
              <Text style={[s.actionText, exp.is_collected && s.actionSaved]}>
                {exp.is_collected ? '已收藏' : `收藏${exp.collection_count > 0 ? ` ${exp.collection_count}` : ''}`}
              </Text>
            </View>
          </TouchableOpacity>
        </View>

        {/* Owner actions */}
        {isOwner && (
          <View style={s.ownerActions}>
            <TouchableOpacity style={s.ownerActionBtn} onPress={handleEdit}>
              <Text style={s.ownerActionText}>编辑</Text>
            </TouchableOpacity>
            {!exp.is_private && exp.visibility !== 'private' ? (
              <TouchableOpacity style={s.ownerActionBtn} onPress={handleMakePrivate}>
                <Text style={s.ownerActionText}>转为私密</Text>
              </TouchableOpacity>
            ) : null}
            <TouchableOpacity style={[s.ownerActionBtn, s.deleteBtn]} onPress={handleDelete}>
              <Text style={s.deleteText}>删除</Text>
            </TouchableOpacity>
          </View>
        )}

        {/* Interpretation */}
        {exp.interpretation ? (
          <View style={s.interpCard}>
            <View style={s.interpTitleRow}>
              <Ionicons name="book-outline" size={16} color="#4a7c59" />
              <Text style={s.interpTitle}>经验解读</Text>
            </View>
            <Text style={s.interpText}>{exp.interpretation}</Text>
          </View>
        ) : (
          <View style={s.noInterp}><Text style={s.noInterpText}>暂无解读</Text></View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  container: {flex: 1, backgroundColor: '#faf8f5'},
  header: {paddingHorizontal: 18, paddingVertical: 12, borderBottomWidth: 0.5, borderBottomColor: '#e8e4df'},
  backButton: {flexDirection: 'row', alignItems: 'center', gap: 2, minHeight: 36},
  backText: {fontSize: 15, color: '#4a7c59', fontWeight: '500'},
  center: {flex: 1, justifyContent: 'center', alignItems: 'center'},
  emptyText: {fontSize: 15, color: '#9a9a9a'},
  errorText: {fontSize: 15, color: '#9a9a9a', marginBottom: 16},
  retryBtn: {backgroundColor: '#4a7c59', borderRadius: 20, paddingHorizontal: 24, paddingVertical: 10},
  retryText: {color: '#ffffff', fontSize: 14, fontWeight: '600'},
  body: {flex: 1},

  // Creator row
  creatorRow: {flexDirection: 'row', alignItems: 'center', paddingHorizontal: 18, paddingTop: 18, paddingBottom: 12},
  avatar: {width: 36, height: 36, borderRadius: 18, backgroundColor: '#eaf2e8', justifyContent: 'center', alignItems: 'center', marginRight: 10},
  avatarText: {fontSize: 15, fontWeight: '700', color: '#4a7c59'},
  creatorName: {fontSize: 15, fontWeight: '700', color: '#1a1a1a', marginRight: 8},
  platformTag: {
    flexDirection: 'row', alignItems: 'center', backgroundColor: '#e8f0e8',
    paddingHorizontal: 6, paddingVertical: 3, borderRadius: 8, marginRight: 8,
  },
  platformTagIcon: {
    fontSize: 10, fontWeight: '800', color: '#ffffff', backgroundColor: '#4a7c59',
    width: 16, height: 16, borderRadius: 8, textAlign: 'center', lineHeight: 16, marginRight: 4,
  },
  platformTagText: {fontSize: 10, color: '#4a7c59', fontWeight: '600'},
  domainTag: {backgroundColor: '#eaf2e8', paddingHorizontal: 10, paddingVertical: 4, borderRadius: 10},
  domainTagText: {fontSize: 11, fontWeight: '600', color: '#4a7c59'},
  privateMark: {marginLeft: 4, width: 18, height: 18, justifyContent: 'center', alignItems: 'center'},

  // Content
  contentCard: {marginHorizontal: 18, backgroundColor: '#ffffff', borderRadius: 16, padding: 18, borderWidth: 0.5, borderColor: '#f0ece7'},
  content: {fontSize: 20, lineHeight: 30, fontWeight: '700', color: '#1a1a1a'},
  source: {fontSize: 12, color: '#9a9a9a', marginTop: 12},
  date: {fontSize: 11, color: '#b5b0a8', marginTop: 4},

  // Score
  scoreCard: {marginHorizontal: 18, marginTop: 12, backgroundColor: '#ffffff', borderRadius: 16, padding: 18, borderWidth: 0.5, borderColor: '#f0ece7'},
  scoreTitle: {fontSize: 12, fontWeight: '600', color: '#9a9a9a', marginBottom: 8},
  scoreStarsRow: {flexDirection: 'row', gap: 4},
  star: {fontSize: 26, color: '#e8a850'},
  scoreOverlay: {flex: 1, backgroundColor: 'rgba(0,0,0,0.35)', justifyContent: 'center', alignItems: 'center'},
  scorePopup: {
    backgroundColor: '#ffffff', borderRadius: 16, padding: 24, marginHorizontal: 40,
    alignItems: 'center', shadowColor: '#000', shadowOffset: {width: 0, height: 4},
    shadowOpacity: 0.12, shadowRadius: 12, elevation: 6,
  },
  scorePopupTitle: {fontSize: 16, fontWeight: '700', color: '#1a1a1a', marginBottom: 10},
  scorePopupText: {fontSize: 14, color: '#4a4a4a', lineHeight: 22, textAlign: 'center'},

  // Actions
  actions: {flexDirection: 'row', marginHorizontal: 18, marginTop: 16, gap: 12},
  actionBtn: {backgroundColor: '#ffffff', borderRadius: 22, paddingHorizontal: 18, paddingVertical: 10, borderWidth: 0.5, borderColor: '#f0ece7'},
  actionContent: {flexDirection: 'row', alignItems: 'center', gap: 5},
  actionText: {fontSize: 13, fontWeight: '600', color: '#6e6e6e'},
  actionLiked: {color: '#e85d5d'},
  actionSaved: {color: '#e8a850'},

  // Owner actions
  ownerActions: {flexDirection: 'row', marginHorizontal: 18, marginTop: 12, gap: 10},
  ownerActionBtn: {flex: 1, backgroundColor: '#ffffff', borderRadius: 12, paddingVertical: 12, alignItems: 'center', borderWidth: 0.5, borderColor: '#e3ded5'},
  ownerActionText: {fontSize: 14, color: '#4a7c59', fontWeight: '700'},
  deleteBtn: {backgroundColor: '#fff5f5', borderColor: '#f4dede'},
  deleteText: {fontSize: 14, color: '#e85d5d', fontWeight: '500'},

  // Interpretation
  interpCard: {marginHorizontal: 18, marginTop: 20, backgroundColor: '#ffffff', borderRadius: 16, padding: 18, borderWidth: 0.5, borderColor: '#d4e0d6'},
  interpTitleRow: {flexDirection: 'row', alignItems: 'center', gap: 6, marginBottom: 10},
  interpTitle: {fontSize: 14, fontWeight: '700', color: '#4a7c59'},
  interpText: {fontSize: 15, lineHeight: 24, color: '#3d3d3d'},
  noInterp: {marginHorizontal: 18, marginTop: 20, padding: 40, alignItems: 'center'},
  noInterpText: {fontSize: 14, color: '#b5b0a8'},
});
