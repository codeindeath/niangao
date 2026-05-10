import React, {useState, useCallback} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  FlatList,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {searchExperiences, Experience, toggleLike, toggleBookmark} from '../services/api';

const DOMAINS: {key: string; label: string}[] = [
  {key: '', label: '全部'},
  {key: 'vitality', label: '生命'},
  {key: 'living', label: '生活'},
  {key: 'work', label: '工作'},
  {key: 'relationship', label: '关系'},
  {key: 'cognition', label: '认知'},
  {key: 'meaning', label: '意义'},
];

const domainLabels: Record<string, string> = {
  vitality: '生命', living: '生活', work: '工作',
  relationship: '关系', cognition: '认知', meaning: '意义',
};

export default function SearchScreen({navigation}: any) {
  const [keyword, setKeyword] = useState('');
  const [domain, setDomain] = useState('');
  const [results, setResults] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = useCallback(async () => {
    if (!keyword.trim()) return;
    setLoading(true);
    setSearched(true);
    try {
      const result = await searchExperiences(keyword.trim());
      let data = result.data || [];
      if (domain) {
        data = data.filter((e: Experience) => e.domain === domain);
      }
      setResults(data);
      setError(null);
    } catch (e) {
      console.error('Search failed:', e);
      setError('搜索失败，请检查网络连接');
    } finally {
      setLoading(false);
    }
  }, [keyword, domain]);

  const handleLike = async (id: string) => {
    setResults(prev =>
      prev.map(e =>
        e.id === id
          ? {...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1}
          : e,
      ),
    );
    try {
      await toggleLike(id);
    } catch (e) {
      console.error('toggleLike failed:', e);
      // rollback
      setResults(prev =>
        prev.map(e =>
          e.id === id
            ? {...e, is_liked: !e.is_liked, like_count: e.is_liked ? e.like_count - 1 : e.like_count + 1}
            : e,
        ),
      );
    }
  };

  const handleBookmark = async (id: string) => {
    setResults(prev =>
      prev.map(e => (e.id === id ? {...e, is_bookmarked: !e.is_bookmarked} : e)),
    );
    try {
      await toggleBookmark(id);
    } catch (e) {
      console.error('toggleBookmark failed:', e);
      // rollback
      setResults(prev =>
        prev.map(e => (e.id === id ? {...e, is_bookmarked: !e.is_bookmarked} : e)),
      );
    }
  };

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Search bar */}
      <View style={styles.searchBar}>
        <TextInput
          style={styles.searchInput}
          value={keyword}
          onChangeText={setKeyword}
          placeholder="搜索经验..."
          placeholderTextColor="#b5b0a8"
          returnKeyType="search"
          onSubmitEditing={handleSearch}
        />
        <TouchableOpacity style={styles.searchButton} onPress={handleSearch}>
          <Text style={styles.searchButtonText}>搜索</Text>
        </TouchableOpacity>
      </View>

      {/* Domain chips */}
      <View style={styles.domainRow}>
        {DOMAINS.map(d => (
          <TouchableOpacity
            key={d.key}
            style={[styles.domainChip, domain === d.key && styles.domainChipActive]}
            onPress={() => setDomain(d.key)}>
            <Text style={[styles.domainChipText, domain === d.key && styles.domainChipTextActive]}>
              {d.label}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {/* Results */}
      {loading ? (
        <ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 80}} />
      ) : error ? (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
          <TouchableOpacity style={styles.retryButton} onPress={() => { setError(null); handleSearch(); }}>
            <Text style={styles.retryButtonText}>重试</Text>
          </TouchableOpacity>
        </View>
      ) : searched && results.length === 0 ? (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyText}>没有找到相关经验</Text>
          <Text style={styles.emptyHint}>试试其他关键词</Text>
        </View>
      ) : (
        <FlatList
          data={results}
          keyExtractor={item => item.id}
          contentContainerStyle={styles.list}
          renderItem={({item}) => (
            <TouchableOpacity
              style={styles.card}
              onPress={() => navigation.navigate('detail', {id: item.id})}
              activeOpacity={0.8}>
              <View style={styles.authorRow}>
                <View style={styles.avatar}>
                  <Text style={styles.avatarText}>{item.author_name?.charAt(0) || '?'}</Text>
                </View>
                <Text style={styles.authorName}>{item.author_name || '匿名'}</Text>
                <View style={styles.domainTag}>
                  <Text style={styles.domainTagText}>
                    {domainLabels[item.domain] || item.domain}
                  </Text>
                </View>
              </View>
              <Text style={styles.content}>{item.content}</Text>
              <View style={styles.actions}>
                <TouchableOpacity onPress={() => handleLike(item.id)}>
                  <Text style={[styles.actionText, item.is_liked && styles.actionLiked]}>
                    ♥ {item.like_count}
                  </Text>
                </TouchableOpacity>
                <TouchableOpacity onPress={() => handleBookmark(item.id)}>
                  <Text style={[styles.actionText, item.is_bookmarked && styles.actionSaved]}>
                    ★ {item.is_bookmarked ? '已收藏' : '收藏'}
                  </Text>
                </TouchableOpacity>
              </View>
            </TouchableOpacity>
          )}
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#faf8f5',
  },
  searchBar: {
    flexDirection: 'row',
    paddingHorizontal: 14,
    paddingVertical: 12,
    gap: 8,
  },
  searchInput: {
    flex: 1,
    backgroundColor: '#ffffff',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 10,
    fontSize: 15,
    color: '#1a1a1a',
    borderWidth: 0.5,
    borderColor: '#e8e4df',
  },
  searchButton: {
    backgroundColor: '#4a7c59',
    borderRadius: 20,
    paddingHorizontal: 18,
    justifyContent: 'center',
  },
  searchButtonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '600',
  },
  domainRow: {
    flexDirection: 'row',
    paddingHorizontal: 14,
    paddingBottom: 10,
    gap: 6,
  },
  domainChip: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 14,
    backgroundColor: '#ffffff',
    borderWidth: 0.5,
    borderColor: '#e8e4df',
  },
  domainChipActive: {
    backgroundColor: '#4a7c59',
    borderColor: '#4a7c59',
  },
  domainChipText: {
    fontSize: 12,
    fontWeight: '500',
    color: '#6e6e6e',
  },
  domainChipTextActive: {
    color: '#ffffff',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingBottom: 100,
  },
  emptyText: {
    fontSize: 15,
    color: '#9a9a9a',
    marginBottom: 6,
  },
  emptyHint: {
    fontSize: 13,
    color: '#b5b0a8',
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
  domainTagText: {
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
    paddingBottom: 100,
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
});
