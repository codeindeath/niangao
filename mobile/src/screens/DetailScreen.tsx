import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {fetchExperience, toggleLike, toggleBookmark, Experience} from '../services/api';

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

const REVIEW_STATUS_LABELS: Record<string, {label: string; color: string; bg: string}> = {
  approved: {label: '✓ 已通过', color: '#4a7c59', bg: '#eaf2e8'},
  rejected: {label: '✗ 未通过', color: '#e85d5d', bg: '#fce8e8'},
  pending: {label: '⏳ 审核中', color: '#e8a850', bg: '#fdf3e4'},
  private: {label: '🔒 私密', color: '#9a9a9a', bg: '#f0f0f0'},
};

function ScoreBar({score, label}: {score: number; label: string}) {
  const pct = Math.min(100, Math.max(0, score * 10));
  return (
    <View style={s.scoreRow}>
      <Text style={s.scoreLabel}>{label}</Text>
      <View style={s.scoreTrack}>
        <View style={[s.scoreFill, {width: `${pct}%`}]} />
      </View>
      <Text style={s.scoreValue}>{score.toFixed(1)}</Text>
    </View>
  );
}

export default function DetailScreen({route, navigation}: any) {
  const {id} = route.params;
  const [exp, setExp] = useState<Experience | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  if (loading) {
    return <SafeAreaView style={s.container}><ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 200}} /></SafeAreaView>;
  }
  if (error && !exp) {
    return <SafeAreaView style={s.container}><View style={s.center}><Text style={s.errorText}>{error}</Text><TouchableOpacity style={s.retryBtn} onPress={() => { setError(null); setLoading(true); loadExperience(); }}><Text style={s.retryText}>重试</Text></TouchableOpacity></View></SafeAreaView>;
  }
  if (!exp) {
    return <SafeAreaView style={s.container}><View style={s.center}><Text style={s.emptyText}>经验不存在或已被删除</Text></View></SafeAreaView>;
  }

  const domainLabel = SUB_DOMAIN_LABELS[exp.sub_domain] || DOMAIN_LABELS[exp.domain] || exp.domain;
  const reviewInfo = REVIEW_STATUS_LABELS[exp.review_status] || REVIEW_STATUS_LABELS.pending;
  const hasScore = exp.quality_score != null && exp.quality_score > 0;

  return (
    <SafeAreaView style={s.container} edges={['top']}>
      <View style={s.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={s.backText}>← 返回</Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={s.body} contentContainerStyle={{paddingBottom: 40}}>
        {/* Author row */}
        <View style={s.authorRow}>
          <View style={s.avatar}><Text style={s.avatarText}>{exp.author_name?.charAt(0) || '?'}</Text></View>
          <Text style={s.authorName}>{exp.author_name || '匿名'}</Text>
          <View style={s.domainTag}><Text style={s.domainText}>{domainLabel}</Text></View>
          {exp.is_private && <Text style={s.privateMark}>🔒</Text>}
        </View>

        {/* Content card */}
        <View style={s.contentCard}>
          <Text style={s.content}>{exp.content}</Text>
          {exp.source_label && <Text style={s.source}>来源：{exp.source_label}</Text>}
          <Text style={s.date}>{new Date(exp.created_at).toLocaleDateString('zh-CN', {year: 'numeric', month: 'long', day: 'numeric'})}</Text>
        </View>

        {/* Review status badge */}
        <View style={[s.reviewBadge, {backgroundColor: reviewInfo.bg}]}>
          <Text style={[s.reviewText, {color: reviewInfo.color}]}>{reviewInfo.label}</Text>
          {exp.review_reason ? <Text style={s.reviewReason}>{exp.review_reason}</Text> : null}
        </View>

        {/* Quality score */}
        {hasScore && (
          <View style={s.scoreCard}>
            <Text style={s.scoreTitle}>📊 质量评分</Text>
            <View style={s.scoreOverall}>
              <Text style={s.scoreBig}>{exp.quality_score!.toFixed(1)}</Text>
              <Text style={s.scoreMax}>/ 10</Text>
            </View>
            {exp.score_details && (() => {
              try {
                const d = JSON.parse(exp.score_details);
                return (
                  <View style={s.scoreDetails}>
                    {d.value != null && <ScoreBar score={d.value} label="内容价值" />}
                    {d.actionable != null && <ScoreBar score={d.actionable} label="可执行度" />}
                    {d.universal != null && <ScoreBar score={d.universal} label="普适性" />}
                    {d.original != null && <ScoreBar score={d.original} label="原创性" />}
                    {d.clarity != null && <ScoreBar score={d.clarity} label="清晰度" />}
                  </View>
                );
              } catch { return null; }
            })()}
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

        {/* Interpretation */}
        {exp.interpretation ? (
          <View style={s.interpCard}>
            <Text style={s.interpTitle}>📖 解读</Text>
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
  authorRow: {flexDirection: 'row', alignItems: 'center', paddingHorizontal: 18, paddingTop: 18, paddingBottom: 12},
  avatar: {width: 32, height: 32, borderRadius: 16, backgroundColor: '#eaf2e8', justifyContent: 'center', alignItems: 'center', marginRight: 10},
  avatarText: {fontSize: 14, fontWeight: '700', color: '#4a7c59'},
  authorName: {fontSize: 14, fontWeight: '600', color: '#4a4a4a', flex: 1},
  domainTag: {backgroundColor: '#eaf2e8', paddingHorizontal: 10, paddingVertical: 4, borderRadius: 10},
  domainText: {fontSize: 11, fontWeight: '600', color: '#4a7c59'},
  privateMark: {fontSize: 14, marginLeft: 4},
  contentCard: {marginHorizontal: 18, backgroundColor: '#ffffff', borderRadius: 16, padding: 18, borderWidth: 0.5, borderColor: '#f0ece7'},
  content: {fontSize: 20, lineHeight: 30, fontWeight: '700', color: '#1a1a1a'},
  source: {fontSize: 12, color: '#9a9a9a', marginTop: 12},
  date: {fontSize: 11, color: '#b5b0a8', marginTop: 4},
  reviewBadge: {marginHorizontal: 18, marginTop: 12, borderRadius: 10, paddingHorizontal: 14, paddingVertical: 8, flexDirection: 'row', alignItems: 'center'},
  reviewText: {fontSize: 12, fontWeight: '600'},
  reviewReason: {fontSize: 11, color: '#9a9a9a', marginLeft: 8, flex: 1},
  scoreCard: {marginHorizontal: 18, marginTop: 12, backgroundColor: '#ffffff', borderRadius: 16, padding: 18, borderWidth: 0.5, borderColor: '#e8e4df'},
  scoreTitle: {fontSize: 14, fontWeight: '700', color: '#4a7c59', marginBottom: 12},
  scoreOverall: {flexDirection: 'row', alignItems: 'baseline', marginBottom: 16},
  scoreBig: {fontSize: 36, fontWeight: '800', color: '#1a1a1a'},
  scoreMax: {fontSize: 16, color: '#9a9a9a', marginLeft: 4},
  scoreDetails: {gap: 8},
  scoreRow: {flexDirection: 'row', alignItems: 'center', gap: 8},
  scoreLabel: {fontSize: 11, color: '#6e6e6e', width: 60, textAlign: 'right'},
  scoreTrack: {flex: 1, height: 6, backgroundColor: '#f0ece7', borderRadius: 3},
  scoreFill: {height: 6, backgroundColor: '#4a7c59', borderRadius: 3},
  scoreValue: {fontSize: 11, color: '#4a7c59', fontWeight: '600', width: 28},
  actions: {flexDirection: 'row', marginHorizontal: 18, marginTop: 16, gap: 12},
  actionBtn: {backgroundColor: '#ffffff', borderRadius: 22, paddingHorizontal: 18, paddingVertical: 10, borderWidth: 0.5, borderColor: '#f0ece7'},
  actionText: {fontSize: 13, fontWeight: '600', color: '#6e6e6e'},
  actionLiked: {color: '#e85d5d'},
  actionSaved: {color: '#e8a850'},
  interpCard: {marginHorizontal: 18, marginTop: 20, backgroundColor: '#ffffff', borderRadius: 16, padding: 18, borderWidth: 0.5, borderColor: '#d4e0d6'},
  interpTitle: {fontSize: 14, fontWeight: '700', color: '#4a7c59', marginBottom: 10},
  interpText: {fontSize: 15, lineHeight: 24, color: '#3d3d3d'},
  noInterp: {marginHorizontal: 18, marginTop: 20, padding: 40, alignItems: 'center'},
  noInterpText: {fontSize: 14, color: '#b5b0a8'},
});
