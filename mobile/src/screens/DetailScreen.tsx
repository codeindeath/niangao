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
import {fetchExperience, toggleLike, toggleBookmark, deleteExperience, Experience} from '../services/api';
import {getUserInfo} from '../services/config';

const DOMAIN_LABELS: Record<string, string> = {
  career: '职场成长', relationship: '人际关系', cognition: '认知升级',
  life: '生活智慧', emotion: '情感',
};
const SUB_DOMAIN_LABELS: Record<string, string> = {
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
    if (!exp) return;
    setExp({...exp, is_liked: !exp.is_liked, like_count: exp.is_liked ? exp.like_count - 1 : exp.like_count + 1});
    try { await toggleLike(exp.id); } catch (e) {
      setExp(prev => prev ? {...prev, is_liked: !prev.is_liked, like_count: prev.is_liked ? prev.like_count - 1 : prev.like_count + 1} : null);
    }
  };

  const handleBookmark = async () => {
    if (!exp) return;
    setExp({...exp, is_bookmarked: !exp.is_bookmarked});
    try { await toggleBookmark(exp.id); } catch (e) {
      setExp(prev => prev ? {...prev, is_bookmarked: !prev.is_bookmarked} : null);
    }
  };

  const handleDelete = () => {
    Alert.alert('删除经验', '确定要删除这条经验吗？', [
      {text: '取消', style: 'cancel'},
      {text: '删除', style: 'destructive', onPress: async () => {
        try { await deleteExperience(exp!.id); navigation.goBack(); }
        catch (e: any) { Alert.alert('删除失败', e?.message || '请稍后再试'); }
      }},
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

  const isPlatform = exp.source_type === 'platform';
  const isRejected = exp.review_status === 'rejected';
  const displayName = exp.creator_name || exp.author_name || '匿名';
  const domainLabel = (exp.sub_domain && SUB_DOMAIN_LABELS[exp.sub_domain]) || DOMAIN_LABELS[exp.domain] || exp.domain;
  const showScore = exp.quality_score != null && exp.quality_score > 0;
  const stars = showScore ? Math.round(exp.quality_score! / 2) : 0;

  return (
    <SafeAreaView style={s.container} edges={['top']}>
      <View style={s.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={s.backText}>← 返回</Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={s.body} contentContainerStyle={{paddingBottom: 40}}>
        {/* Creator row */}
        <View style={s.creatorRow}>
          <View style={s.avatar}><Text style={s.avatarText}>{displayName.charAt(0)}</Text></View>
          <Text style={s.creatorName}>{displayName}</Text>
          {isPlatform && (
            <View style={s.platformTag}>
              <Text style={s.platformTagIcon}>官</Text>
              <Text style={s.platformTagText}>平台生产</Text>
            </View>
          )}
          <View style={s.domainTag}><Text style={s.domainTagText}>{domainLabel}</Text></View>
          {exp.is_private && <Text style={s.privateMark}>🔒</Text>}
        </View>

        {/* Content card */}
        <View style={s.contentCard}>
          <Text style={s.content}>{exp.content}</Text>
          {exp.source_label && <Text style={s.source}>来源：{exp.source_label}</Text>}
          <Text style={s.date}>{new Date(exp.created_at).toLocaleDateString('zh-CN', {year: 'numeric', month: 'long', day: 'numeric'})}</Text>
        </View>

        {/* Rejected indicator (only for rejected experiences) */}
        {isRejected && (
          <View style={s.rejectedCard}>
            <Text style={s.rejectedIcon}>❕</Text>
            <View style={{flex: 1}}>
              <Text style={s.rejectedTitle}>这条经验未通过审核，仅你自己可见</Text>
              {exp.review_reason ? <Text style={s.rejectedReason}>{exp.review_reason}</Text> : null}
            </View>
          </View>
        )}

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
          <TouchableOpacity style={s.actionBtn} onPress={handleLike}>
            <Text style={[s.actionText, exp.is_liked && s.actionLiked]}>♥ {exp.like_count > 0 ? exp.like_count : '点赞'}</Text>
          </TouchableOpacity>
          <TouchableOpacity style={s.actionBtn} onPress={handleBookmark}>
            <Text style={[s.actionText, exp.is_bookmarked && s.actionSaved]}>★ {exp.is_bookmarked ? '已收藏' : '收藏'}</Text>
          </TouchableOpacity>
        </View>

        {/* Delete — only for author */}
        {currentUserId && exp.author_id === currentUserId && (
          <TouchableOpacity style={s.deleteBtn} onPress={handleDelete}>
            <Text style={s.deleteText}>🗑 删除此经验</Text>
          </TouchableOpacity>
        )}

        {/* Interpretation */}
        {exp.interpretation ? (
          <View style={s.interpCard}>
            <Text style={s.interpTitle}>📖 经验解读</Text>
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
  privateMark: {fontSize: 14, marginLeft: 4},

  // Content
  contentCard: {marginHorizontal: 18, backgroundColor: '#ffffff', borderRadius: 16, padding: 18, borderWidth: 0.5, borderColor: '#f0ece7'},
  content: {fontSize: 20, lineHeight: 30, fontWeight: '700', color: '#1a1a1a'},
  source: {fontSize: 12, color: '#9a9a9a', marginTop: 12},
  date: {fontSize: 11, color: '#b5b0a8', marginTop: 4},

  // Rejected
  rejectedCard: {
    marginHorizontal: 18, marginTop: 12, backgroundColor: '#f8f8f8',
    borderRadius: 12, padding: 14, flexDirection: 'row', alignItems: 'flex-start', gap: 10,
  },
  rejectedIcon: {fontSize: 20, color: '#9a9a9a', marginTop: 1},
  rejectedTitle: {fontSize: 13, fontWeight: '600', color: '#6e6e6e'},
  rejectedReason: {fontSize: 12, color: '#9a9a9a', marginTop: 3},

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
  actionText: {fontSize: 13, fontWeight: '600', color: '#6e6e6e'},
  actionLiked: {color: '#e85d5d'},
  actionSaved: {color: '#e8a850'},

  // Delete
  deleteBtn: {marginHorizontal: 18, marginTop: 12, backgroundColor: '#fff5f5', borderRadius: 12, paddingVertical: 12, alignItems: 'center', borderWidth: 0.5, borderColor: '#fce8e8'},
  deleteText: {fontSize: 14, color: '#e85d5d', fontWeight: '500'},

  // Interpretation
  interpCard: {marginHorizontal: 18, marginTop: 20, backgroundColor: '#ffffff', borderRadius: 16, padding: 18, borderWidth: 0.5, borderColor: '#d4e0d6'},
  interpTitle: {fontSize: 14, fontWeight: '700', color: '#4a7c59', marginBottom: 10},
  interpText: {fontSize: 15, lineHeight: 24, color: '#3d3d3d'},
  noInterp: {marginHorizontal: 18, marginTop: 20, padding: 40, alignItems: 'center'},
  noInterpText: {fontSize: 14, color: '#b5b0a8'},
});
