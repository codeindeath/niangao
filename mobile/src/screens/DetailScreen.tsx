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

export default function DetailScreen({route, navigation}: any) {
  const {id} = route.params;
  const [exp, setExp] = useState<Experience | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadExperience();
  }, [id]);

  const loadExperience = async () => {
    try {
      const data = await fetchExperience(id);
      setExp(data);
    } catch (e) {
      console.error('Failed to load experience:', e);
    } finally {
      setLoading(false);
    }
  };

  const handleLike = async () => {
    if (!exp) return;
    await toggleLike(exp.id);
    setExp({
      ...exp,
      is_liked: !exp.is_liked,
      like_count: exp.is_liked ? exp.like_count - 1 : exp.like_count + 1,
    });
  };

  const handleBookmark = async () => {
    if (!exp) return;
    await toggleBookmark(exp.id);
    setExp({...exp, is_bookmarked: !exp.is_bookmarked});
  };

  const domainLabels: Record<string, string> = {
    career: '职场成长',
    relationship: '人际关系',
    cognition: '认知升级',
    life: '生活智慧',
    emotion: '情感',
  };

  if (loading) {
    return (
      <SafeAreaView style={styles.container}>
        <ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 200}} />
      </SafeAreaView>
    );
  }

  if (!exp) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyText}>经验不存在或已被删除</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.backText}>← 返回</Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.body} contentContainerStyle={{paddingBottom: 40}}>
        {/* Author row */}
        <View style={styles.authorRow}>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{exp.author_name?.charAt(0) || '?'}</Text>
          </View>
          <Text style={styles.authorName}>{exp.author_name || '匿名'}</Text>
          <View style={styles.domainTag}>
            <Text style={styles.domainText}>{domainLabels[exp.domain] || exp.domain}</Text>
          </View>
        </View>

        {/* Content card */}
        <View style={styles.contentCard}>
          <Text style={styles.content}>{exp.content}</Text>
          {exp.source_label && (
            <Text style={styles.source}>来源：{exp.source_label}</Text>
          )}
          <Text style={styles.date}>
            {new Date(exp.created_at).toLocaleDateString('zh-CN', {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
            })}
          </Text>
        </View>

        {/* Actions */}
        <View style={styles.actions}>
          <TouchableOpacity style={styles.actionBtn} onPress={handleLike}>
            <Text style={[styles.actionText, exp.is_liked && styles.actionLiked]}>
              ♥ {exp.like_count > 0 ? exp.like_count : '点赞'}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.actionBtn} onPress={handleBookmark}>
            <Text style={[styles.actionText, exp.is_bookmarked && styles.actionSaved]}>
              ★ {exp.is_bookmarked ? '已收藏' : '收藏'}
            </Text>
          </TouchableOpacity>
        </View>

        {/* Interpretation card */}
        {exp.interpretation ? (
          <View style={styles.interpretationCard}>
            <Text style={styles.interpretationTitle}>📖 解读</Text>
            <Text style={styles.interpretationText}>{exp.interpretation}</Text>
          </View>
        ) : (
          <View style={styles.noInterpretation}>
            <Text style={styles.noInterpretationText}>暂无解读</Text>
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#faf8f5',
  },
  header: {
    paddingHorizontal: 18,
    paddingVertical: 12,
    borderBottomWidth: 0.5,
    borderBottomColor: '#e8e4df',
  },
  backText: {
    fontSize: 15,
    color: '#4a7c59',
    fontWeight: '500',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 15,
    color: '#9a9a9a',
  },
  body: {
    flex: 1,
  },
  authorRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 18,
    paddingTop: 18,
    paddingBottom: 12,
  },
  avatar: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#eaf2e8',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 10,
  },
  avatarText: {
    fontSize: 14,
    fontWeight: '700',
    color: '#4a7c59',
  },
  authorName: {
    fontSize: 14,
    fontWeight: '600',
    color: '#4a4a4a',
    flex: 1,
  },
  domainTag: {
    backgroundColor: '#eaf2e8',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 10,
  },
  domainText: {
    fontSize: 11,
    fontWeight: '600',
    color: '#4a7c59',
  },
  contentCard: {
    marginHorizontal: 18,
    backgroundColor: '#ffffff',
    borderRadius: 16,
    padding: 18,
    borderWidth: 0.5,
    borderColor: '#f0ece7',
  },
  content: {
    fontSize: 20,
    lineHeight: 30,
    fontWeight: '700',
    color: '#1a1a1a',
  },
  source: {
    fontSize: 12,
    color: '#9a9a9a',
    marginTop: 12,
  },
  date: {
    fontSize: 11,
    color: '#b5b0a8',
    marginTop: 4,
  },
  actions: {
    flexDirection: 'row',
    marginHorizontal: 18,
    marginTop: 16,
    gap: 12,
  },
  actionBtn: {
    backgroundColor: '#ffffff',
    borderRadius: 22,
    paddingHorizontal: 18,
    paddingVertical: 10,
    borderWidth: 0.5,
    borderColor: '#f0ece7',
  },
  actionText: {
    fontSize: 13,
    fontWeight: '600',
    color: '#6e6e6e',
  },
  actionLiked: {
    color: '#e85d5d',
  },
  actionSaved: {
    color: '#e8a850',
  },
  interpretationCard: {
    marginHorizontal: 18,
    marginTop: 20,
    backgroundColor: '#ffffff',
    borderRadius: 16,
    padding: 18,
    borderWidth: 0.5,
    borderColor: '#d4e0d6',
  },
  interpretationTitle: {
    fontSize: 14,
    fontWeight: '700',
    color: '#4a7c59',
    marginBottom: 10,
  },
  interpretationText: {
    fontSize: 15,
    lineHeight: 24,
    color: '#3d3d3d',
  },
  noInterpretation: {
    marginHorizontal: 18,
    marginTop: 20,
    padding: 40,
    alignItems: 'center',
  },
  noInterpretationText: {
    fontSize: 14,
    color: '#b5b0a8',
  },
});
