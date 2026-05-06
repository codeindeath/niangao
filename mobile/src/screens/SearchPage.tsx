import React, {useState, useCallback, useEffect, useRef} from 'react';
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
import {searchExperiences, Experience} from '../services/api';

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

export default function SearchPage({navigation}: any) {
  const [keyword, setKeyword] = useState('');
  const [results, setResults] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<TextInput>(null);

  // Auto-focus on mount
  useEffect(() => {
    const t = setTimeout(() => inputRef.current?.focus(), 100);
    return () => clearTimeout(t);
  }, []);

  const handleSearch = useCallback(async () => {
    if (!keyword.trim()) return;
    setLoading(true);
    setSearched(true);
    setError(null);
    try {
      const result = await searchExperiences(keyword.trim());
      setResults(result.data || []);
    } catch (e) {
      console.error('Search failed:', e);
      setError('搜索失败，请检查网络连接');
    } finally {
      setLoading(false);
    }
  }, [keyword]);

  const handleResultPress = (item: Experience, index: number) => {
    navigation.navigate('searchCard', {
      results: results,
      initialIndex: index,
      keyword: keyword.trim(),
    });
  };

  const renderResult = ({item, index}: {item: Experience; index: number}) => {
    const isPlatform = item.source_type === 'platform';
    const primaryDomain = DOMAIN_LABELS[item.domain] || item.domain;
    const subDomain = item.sub_domain ? SUB_LABELS[item.sub_domain] : null;
    const domainLabel = subDomain || primaryDomain;
    const displayName = item.creator_name || item.author_name || '匿名';
    const showScore = item.quality_score != null && item.quality_score > 0;
    const stars = showScore ? Math.round(item.quality_score! / 2) : 0;

    return (
      <TouchableOpacity
        style={s.resultItem}
        onPress={() => handleResultPress(item, index)}
        activeOpacity={0.7}
      >
        <View style={s.resultAvatar}>
          <Text style={s.resultAvatarText}>{displayName.charAt(0)}</Text>
        </View>
        <View style={s.resultInfo}>
          <Text style={s.resultContent} numberOfLines={2}>{item.content}</Text>
          <Text style={s.resultMeta}>
            {displayName} · {domainLabel}
            {isPlatform && <Text style={{color: '#4a7c59'}}> · 官</Text>}
            {showScore ? ` · ${'★'.repeat(stars)}` : ''}
          </Text>
        </View>
        <Text style={s.resultArrow}>›</Text>
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView style={s.container} edges={['top']}>
      {/* Search header */}
      <View style={s.header}>
        <TouchableOpacity style={s.backBtn} onPress={() => navigation.goBack()}>
          <Text style={s.backBtnText}>←</Text>
        </TouchableOpacity>
        <View style={s.inputWrap}>
          <TextInput
            ref={inputRef}
            style={s.input}
            value={keyword}
            onChangeText={setKeyword}
            placeholder="输入你想找的经验、创作者..."
            placeholderTextColor="#b5b0a8"
            returnKeyType="search"
            onSubmitEditing={handleSearch}
          />
        </View>
        <TouchableOpacity style={s.searchBtn} onPress={handleSearch}>
          <Text style={s.searchBtnText}>搜索</Text>
        </TouchableOpacity>
      </View>

      {/* Results */}
      {loading ? (
        <ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 80}} />
      ) : error ? (
        <View style={s.center}>
          <Text style={s.emptyText}>{error}</Text>
          <TouchableOpacity style={s.retryBtn} onPress={() => { setError(null); handleSearch(); }}>
            <Text style={s.retryBtnText}>重试</Text>
          </TouchableOpacity>
        </View>
      ) : searched && results.length === 0 ? (
        <View style={s.center}>
          <Text style={s.emptyText}>没有找到相关经验</Text>
          <Text style={s.emptyHint}>试试其他关键词</Text>
        </View>
      ) : (
        <FlatList
          data={results}
          keyExtractor={item => item.id}
          contentContainerStyle={s.list}
          renderItem={renderResult}
        />
      )}
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  container: {flex: 1, backgroundColor: '#faf8f5'},
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 8,
    gap: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#e8e4dc',
  },
  backBtn: {
    width: 36, height: 36,
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  backBtnText: {fontSize: 20, color: '#5c5548', fontWeight: '600'},
  inputWrap: {
    flex: 1,
    backgroundColor: '#ece8df',
    borderRadius: 10,
    paddingHorizontal: 12,
    height: 36,
    justifyContent: 'center',
  },
  input: {fontSize: 15, color: '#2a2722', padding: 0},
  searchBtn: {
    paddingHorizontal: 16,
    height: 36,
    backgroundColor: '#4a7c59',
    borderRadius: 10,
    justifyContent: 'center',
  },
  searchBtnText: {color: '#fff', fontSize: 14, fontWeight: '500'},
  list: {paddingVertical: 8},
  resultItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 14,
    paddingHorizontal: 16,
    gap: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f0ece7',
  },
  resultAvatar: {
    width: 40, height: 40,
    borderRadius: 20,
    backgroundColor: '#eaf2e8',
    justifyContent: 'center',
    alignItems: 'center',
  },
  resultAvatarText: {fontSize: 14, fontWeight: '700', color: '#4a7c59'},
  resultInfo: {flex: 1, minWidth: 0},
  resultContent: {fontSize: 15, color: '#2a2722', lineHeight: 22, marginBottom: 3},
  resultMeta: {fontSize: 12, color: '#9b9487'},
  resultArrow: {fontSize: 16, color: '#c5bfb3', flexShrink: 0},
  center: {flex: 1, justifyContent: 'center', alignItems: 'center', paddingBottom: 100},
  emptyText: {fontSize: 15, color: '#9a9a9a', marginBottom: 6},
  emptyHint: {fontSize: 13, color: '#b5b0a8'},
  retryBtn: {
    marginTop: 16,
    backgroundColor: '#4a7c59',
    borderRadius: 20,
    paddingHorizontal: 24,
    paddingVertical: 10,
  },
  retryBtnText: {color: '#fff', fontSize: 14, fontWeight: '600'},
});
