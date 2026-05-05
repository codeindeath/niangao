import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { fetchExperiences, Experience, toggleLike, toggleBookmark } from '../services/api';

export default function HomeScreen({ navigation }: any) {
  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const loadExperiences = useCallback(async (pageNum: number, refresh = false) => {
    try {
      const result = await fetchExperiences(pageNum);
      const data = Array.isArray(result?.data) ? result.data : [];
      setExperiences(prev => refresh ? data : [...prev, ...data]);
      setError(null);
    } catch (e) {
      console.error('Failed to load experiences:', e);
      setError('加载失败，请检查网络连接');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadExperiences(1, true);
  }, []);

  const handleRefresh = () => {
    setRefreshing(true);
    setPage(1);
    loadExperiences(1, true);
  };

  const handleLoadMore = () => {
    const nextPage = page + 1;
    setPage(nextPage);
    loadExperiences(nextPage);
  };

  const handleLike = async (id: string) => {
    setExperiences(prev =>
      prev.map(e =>
        e.id === id
          ? { ...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1 }
          : e
      )
    );
    try {
      await toggleLike(id);
    } catch (e) {
      console.error('toggleLike failed:', e);
      // rollback
      setExperiences(prev =>
        prev.map(e =>
          e.id === id
            ? { ...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1 }
            : e
        )
      );
    }
  };

  const handleBookmark = async (id: string) => {
    setExperiences(prev =>
      prev.map(e => (e.id === id ? { ...e, is_bookmarked: !e.is_bookmarked } : e))
    );
    try {
      await toggleBookmark(id);
    } catch (e) {
      console.error('toggleBookmark failed:', e);
      // rollback
      setExperiences(prev =>
        prev.map(e => (e.id === id ? { ...e, is_bookmarked: !e.is_bookmarked } : e))
      );
    }
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
        <ActivityIndicator size="large" color="#4a7c59" style={{ marginTop: 200 }} />
      </SafeAreaView>
    );
  }

  if (error && experiences.length === 0) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
          <TouchableOpacity style={styles.retryButton} onPress={() => { setError(null); handleRefresh(); }}>
            <Text style={styles.retryButtonText}>重试</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>为你推荐</Text>
        <Text style={styles.headerSub}>基于你的兴趣</Text>
      </View>
      <FlatList
        data={experiences}
        keyExtractor={item => item.id}
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={handleRefresh} tintColor="#4a7c59" />}
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.5}
        contentContainerStyle={styles.list}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>暂无推荐内容</Text>
          </View>
        }
        renderItem={({ item }) => (
          <TouchableOpacity
            style={styles.card}
            onPress={() => navigation.navigate('detail', { id: item.id })}
            activeOpacity={0.8}
          >
            {/* Author row */}
            <View style={styles.authorRow}>
              <View style={styles.avatar}>
                <Text style={styles.avatarText}>
                  {item.author_name?.charAt(0) || '?'}
                </Text>
              </View>
              <Text style={styles.authorName}>{item.author_name || '匿名'}</Text>
              <View style={styles.domainTag}>
                <Text style={styles.domainText}>
                  {domainLabels[item.domain] || item.domain}
                </Text>
              </View>
            </View>
            {/* Content */}
            <Text style={styles.content}>{item.content}</Text>
            {/* Actions */}
            <View style={styles.actions}>
              <TouchableOpacity onPress={() => handleLike(item.id)} style={styles.actionButton}>
                <Text style={[styles.actionText, item.is_liked && styles.actionLiked]}>
                  ♥ {item.like_count}
                </Text>
              </TouchableOpacity>
              <TouchableOpacity onPress={() => handleBookmark(item.id)} style={styles.actionButton}>
                <Text style={[styles.actionText, item.is_bookmarked && styles.actionSaved]}>
                  ★ {item.is_bookmarked ? '已收藏' : '收藏'}
                </Text>
              </TouchableOpacity>
            </View>
          </TouchableOpacity>
        )}
      />
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
    paddingTop: 4,
    paddingBottom: 8,
  },
  headerTitle: {
    fontSize: 13,
    fontWeight: '700',
    color: '#9a9a9a',
    letterSpacing: 1,
    textTransform: 'uppercase',
  },
  headerSub: {
    fontSize: 11,
    color: '#9a9a9a',
    marginTop: 1,
  },
  list: {
    paddingHorizontal: 14,
    paddingBottom: 20,
  },
  card: {
    backgroundColor: '#ffffff',
    borderRadius: 16,
    padding: 16,
    marginBottom: 10,
    borderWidth: 0.5,
    borderColor: '#f0ece7',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.04,
    shadowRadius: 6,
    elevation: 1,
  },
  authorRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  avatar: {
    width: 22,
    height: 22,
    borderRadius: 11,
    backgroundColor: '#eaf2e8',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 6,
  },
  avatarText: {
    fontSize: 11,
    fontWeight: '700',
    color: '#4a7c59',
  },
  authorName: {
    fontSize: 12,
    fontWeight: '500',
    color: '#6e6e6e',
    flex: 1,
  },
  domainTag: {
    backgroundColor: '#eaf2e8',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 8,
  },
  domainText: {
    fontSize: 10,
    fontWeight: '600',
    color: '#4a7c59',
  },
  content: {
    fontSize: 15,
    lineHeight: 23,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 10,
  },
  actions: {
    flexDirection: 'row',
    gap: 16,
    paddingTop: 8,
    borderTopWidth: 0.5,
    borderTopColor: '#f0ece7',
  },
  actionButton: {
    paddingVertical: 2,
  },
  actionText: {
    fontSize: 11,
    color: '#9a9a9a',
  },
  actionLiked: {
    color: '#e85d5d',
  },
  actionSaved: {
    color: '#e8a850',
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingBottom: 80,
  },
  errorText: {
    fontSize: 15,
    color: '#9a9a9a',
    marginBottom: 16,
  },
  retryButton: {
    backgroundColor: '#4a7c59',
    borderRadius: 20,
    paddingHorizontal: 24,
    paddingVertical: 10,
  },
  retryButtonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '600',
  },
  emptyContainer: {
    paddingTop: 100,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 15,
    color: '#9a9a9a',
  },
});
