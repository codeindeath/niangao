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
import {searchExperiences, recordSearchClick, Experience} from '../services/api';
import {handleAuthExpired, requireLogin} from '../utils/authGate';
import {reportHandledError} from '../utils/logging';
import Ionicons from '@expo/vector-icons/Ionicons';

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

export default function SearchPage({navigation}: any) {
  const [keyword, setKeyword] = useState('');
  const [results, setResults] = useState<Experience[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<TextInput>(null);
  const searchInFlightRef = useRef(false);

  // Auto-focus on mount
  useEffect(() => {
    const t = setTimeout(() => inputRef.current?.focus(), 100);
    return () => clearTimeout(t);
  }, []);

  const handleSearch = useCallback(async () => {
    const trimmedKeyword = keyword.trim();
    if (!trimmedKeyword || searchInFlightRef.current) return;
    searchInFlightRef.current = true;
    setLoading(true);
    setSearched(true);
    setError(null);
    try {
      const result = await searchExperiences(trimmedKeyword);
      setResults(result.data || []);
    } catch (e) {
      if (await handleAuthExpired(navigation, e)) return;
      reportHandledError('SearchPage.search', e);
      setError('搜索失败，请检查网络连接');
    } finally {
      searchInFlightRef.current = false;
      setLoading(false);
    }
  }, [keyword, navigation]);

  const handleResultPress = (item: Experience, index: number) => {
    recordSearchClick(item.id, keyword.trim(), index);
    navigation.navigate('searchCard', {
      results: results,
      initialIndex: index,
      keyword: keyword.trim(),
    });
  };

  const handleChatEntry = async () => {
    if (!(await requireLogin(navigation, '登录后可以和年糕聊聊这件事。'))) return;
    navigation.navigate('main', {screen: 'chat'});
  };

  const renderResult = ({item, index}: {item: Experience; index: number}) => {
    const isPlatform = item.experience_type === 'platform_selected';
    const primaryDomain = DOMAIN_LABELS[item.domain] || item.domain;
    const subDomain = item.sub_domain ? SUB_LABELS[item.sub_domain] : null;
    const domainLabel = subDomain || primaryDomain;
    const displayName = item.creator_display_name || '匿名';
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
            {isPlatform && <Text style={{color: '#4a7c59'}}> · 精选</Text>}
            {showScore ? ` · ${'★'.repeat(stars)}` : ''}
          </Text>
        </View>
        <Ionicons name="chevron-forward" size={17} color="#c5bfb3" />
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView style={s.container} edges={['top']}>
      {/* Search header */}
      <View style={s.header}>
        <TouchableOpacity style={s.backBtn} onPress={() => navigation.goBack()} accessibilityRole="button" accessibilityLabel="返回">
          <Ionicons name="chevron-back" size={21} color="#5c5548" />
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
        <TouchableOpacity
          style={[s.searchBtn, loading && s.searchBtnDisabled]}
          onPress={handleSearch}
          disabled={loading}
          accessibilityRole="button"
          accessibilityLabel="搜索">
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
          <Text style={s.emptyText}>没找到特别贴近的，换个说法试试</Text>
          <TouchableOpacity style={s.chatBtn} onPress={handleChatEntry}>
            <Text style={s.chatBtnText}>去聊聊</Text>
          </TouchableOpacity>
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
  searchBtnDisabled: {opacity: 0.55},
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
  center: {flex: 1, justifyContent: 'center', alignItems: 'center', paddingBottom: 100},
  emptyText: {fontSize: 15, color: '#9a9a9a', marginBottom: 6},
  emptyHint: {fontSize: 13, color: '#b5b0a8'},
  chatBtn: {
    marginTop: 14,
    borderRadius: 14,
    borderWidth: 1,
    borderColor: '#d9e4d5',
    backgroundColor: '#fff',
    paddingHorizontal: 18,
    paddingVertical: 9,
  },
  chatBtnText: {fontSize: 13, color: '#4a7c59', fontWeight: '700'},
  retryBtn: {
    marginTop: 16,
    backgroundColor: '#4a7c59',
    borderRadius: 20,
    paddingHorizontal: 24,
    paddingVertical: 10,
  },
  retryBtnText: {color: '#fff', fontSize: 14, fontWeight: '600'},
});
