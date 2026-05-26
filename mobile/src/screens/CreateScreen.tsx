import React, {useState, useRef, useEffect} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  Alert,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Animated,
  Dimensions,
  Modal,
} from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import {SafeAreaView} from 'react-native-safe-area-context';
import {createExperience, updateExperience, rewriteExperience, updateProfile, ApiError, Experience} from '../services/api';
import {triggerTabRefresh} from './HomeScreen';
import {handleAuthExpired} from '../utils/authGate';
import Ionicons from '@expo/vector-icons/Ionicons';

const PRIMARY_DOMAINS: {key: string; label: string}[] = [
  {key: 'vitality', label: '生命'},
  {key: 'living', label: '生活'},
  {key: 'work', label: '工作'},
  {key: 'relationship', label: '关系'},
  {key: 'cognition', label: '认知'},
  {key: 'meaning', label: '意义'},
];

const SUB_DOMAINS: Record<string, {key: string; label: string}[]> = {
  vitality: [
    {key: 'health', label: '健康'},
    {key: 'housing', label: '居住'},
    {key: 'transit', label: '出行'},
    {key: 'diet', label: '饮食'},
    {key: 'exercise', label: '运动'},
  ],
  living: [
    {key: 'pets', label: '宠物'},
    {key: 'travel', label: '旅行'},
    {key: 'fashion', label: '衣着'},
    {key: 'selfcare', label: '养护'},
    {key: 'shopping', label: '购物'},
    {key: 'fun', label: '娱乐'},
  ],
  work: [
    {key: 'jobhunt', label: '求职'},
    {key: 'promotion', label: '升职'},
    {key: 'startup', label: '创业'},
    {key: 'work-comm', label: '沟通'},
    {key: 'management', label: '管理'},
    {key: 'productivity', label: '效率'},
  ],
  relationship: [
    {key: 'marriage', label: '夫妻'},
    {key: 'romance', label: '恋人'},
    {key: 'friendship', label: '朋友'},
    {key: 'parenting', label: '亲子'},
    {key: 'parents', label: '父母'},
    {key: 'siblings', label: '兄妹'},
  ],
  cognition: [
    {key: 'cog-learning', label: '学习'},
    {key: 'thinking', label: '思维'},
    {key: 'info', label: '信息'},
    {key: 'tools', label: '工具'},
    {key: 'creativity', label: '创造'},
    {key: 'expression', label: '表达'},
  ],
  meaning: [
    {key: 'self', label: '自我'},
    {key: 'happiness', label: '幸福'},
    {key: 'emotion', label: '情绪'},
    {key: 'faith', label: '信仰'},
    {key: 'mission', label: '使命'},
    {key: 'belonging', label: '归属'},
  ],
};

const DRAFT_KEY = '@niangao_create_draft';

interface DraftData {
  content: string;
  domain: string;
  subDomain: string;
  topic: string;
  isPrivate: boolean;
}

interface RewriteCandidate {
  content: string;
  domain?: string;
  subDomain?: string;
  topic?: string;
}

export default function CreateScreen({navigation, route}: any) {
  const editingExperience = route?.params?.experience as Experience | undefined;
  const isEditing = !!editingExperience;
  const sourceScene = route?.params?.sourceScene === 'chat' ? 'chat' : 'note';
  const sourceMessageIds = Array.isArray(route?.params?.sourceMessageIds)
    ? route.params.sourceMessageIds
    : undefined;
  const [content, setContent] = useState('');
  const [domain, setDomain] = useState('');
  const [subDomain, setSubDomain] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [rewriting, setRewriting] = useState(false);
  const [rewriteCandidate, setRewriteCandidate] = useState<RewriteCandidate | null>(null);
  const [isFocused, setIsFocused] = useState(false);
  const [topic, setTopic] = useState('');
  const [isTopicEditing, setIsTopicEditing] = useState(false);
  const [displayNameModalVisible, setDisplayNameModalVisible] = useState(false);
  const [displayNameDraft, setDisplayNameDraft] = useState('');
  const [displayNameSubmitting, setDisplayNameSubmitting] = useState(false);
  const [displayNameRetryPrivate, setDisplayNameRetryPrivate] = useState<boolean | null>(null);

  const inputRef = useRef<TextInput>(null);
  const topicPageInputRef = useRef<TextInput>(null);
  const [topicDraft, setTopicDraft] = useState('');
  const subdomainOpacity = useRef(new Animated.Value(0)).current;
  const subdomainHeight = useRef(new Animated.Value(0)).current;
  const slideAnim = useRef(new Animated.Value(0)).current;
  const screenWidth = Dimensions.get('window').width;

  // Auto-focus on mount
  useEffect(() => {
    const timer = setTimeout(() => {
      inputRef.current?.focus();
    }, 150);
    return () => clearTimeout(timer);
  }, []);

  // Reset slide position when screen gains focus
  useEffect(() => {
    const unsub = navigation.addListener('focus', () => {
      slideAnim.setValue(0);
    });
    return unsub;
  }, [navigation, slideAnim]);

  // Load cached draft on mount
  useEffect(() => {
    if (editingExperience) {
      setContent(editingExperience.content || '');
      setDomain(editingExperience.domain || '');
      setSubDomain(editingExperience.sub_domain || '');
      setTopic(editingExperience.topic || '');
      setIsPrivate(editingExperience.visibility === 'private');
      return;
    }

    if (typeof route?.params?.prefillContent === 'string') {
      setContent(route.params.prefillContent);
      setDomain(route.params.domain || '');
      setSubDomain(route.params.subDomain || route.params.sub_domain || '');
      setTopic(route.params.topic || '');
      setIsPrivate(route.params.defaultVisibility === 'private' || sourceScene === 'chat');
    }
  }, [editingExperience, route?.params, sourceScene]);

  useEffect(() => {
    if (isEditing || typeof route?.params?.prefillContent === 'string') return;
    AsyncStorage.getItem(DRAFT_KEY).then(data => {
      if (data) {
        try {
          const draft: DraftData = JSON.parse(data);
          if (draft.content) {
            setContent(draft.content);
            setDomain(draft.domain || '');
            setSubDomain(draft.subDomain || '');
            setTopic(draft.topic || '');
            setIsPrivate(draft.isPrivate || false);
          }
        } catch {}
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isEditing, route?.params?.prefillContent]);

  useEffect(() => {
    Animated.parallel([
      Animated.timing(subdomainOpacity, {
        toValue: domain ? 1 : 0,
        duration: domain ? 180 : 120,
        useNativeDriver: false,
      }),
      Animated.timing(subdomainHeight, {
        toValue: domain ? 1 : 0,
        duration: domain ? 180 : 120,
        useNativeDriver: false,
      }),
    ]).start();
  }, [domain, subdomainHeight, subdomainOpacity]);

  useEffect(() => {
    setRewriteCandidate(null);
  }, [content]);

  const handleBack = () => {
    if (isEditing) {
      navigation.goBack();
      return;
    }
    if (content.trim()) {
      const draft: DraftData = {content, domain, subDomain, topic, isPrivate};
      AsyncStorage.setItem(DRAFT_KEY, JSON.stringify(draft));
    } else {
      AsyncStorage.removeItem(DRAFT_KEY);
    }
    Animated.timing(slideAnim, {
      toValue: -screenWidth,
      duration: 220,
      useNativeDriver: true,
    }).start(() => navigation.navigate('home'));
  };

  const handleDomainSelect = (key: string) => {
    if (domain === key) {
      setDomain('');
      setSubDomain('');
    } else {
      setDomain(key);
      setSubDomain('');
    }
  };

  const isDisplayNameRequiredError = (e: any) => (
    e?.code === 'display_name_required' ||
    (e instanceof ApiError && e.code === 'display_name_required')
  );

  const persistExperience = async (trimmedContent: string, privateSave: boolean = isPrivate) => {
    if (isEditing && editingExperience) {
      await updateExperience(
        editingExperience.id,
        trimmedContent,
        domain,
        subDomain,
        privateSave,
        editingExperience.interpretation,
        topic.trim(),
      );
    } else {
      await createExperience(
        trimmedContent,
        domain,
        subDomain,
        privateSave,
        undefined,
        topic.trim(),
        {
          source_scene: sourceScene,
          source_message_ids: sourceMessageIds,
        },
      );
      await AsyncStorage.removeItem(DRAFT_KEY);
    }
    Alert.alert(isEditing ? '已保存' : '已记下', isEditing && !privateSave ? '修改后会重新处理' : '你的经验已保存', [
      {text: '好的', onPress: () => {
        triggerTabRefresh('my');
        navigation.goBack();
      }},
    ]);
  };

  const handlePersistError = async (e: any, privateSave: boolean = isPrivate) => {
    if (await handleAuthExpired(navigation, e)) {
      return;
    } else if (!isEditing && !privateSave && isDisplayNameRequiredError(e)) {
      setDisplayNameRetryPrivate(privateSave);
      setDisplayNameModalVisible(true);
    } else {
      Alert.alert('发布失败', e?.message || String(e));
    }
  };

  const saveWithVisibility = async (trimmedContent: string, privateSave: boolean) => {
    setSubmitting(true);
    try {
      await persistExperience(trimmedContent, privateSave);
    } catch (e: any) {
      await handlePersistError(e, privateSave);
    } finally {
      setSubmitting(false);
    }
  };

  const handlePublish = async () => {
    const trimmedContent = content.trim();
    const contentLength = Array.from(trimmedContent).length;

    if (!trimmedContent) {
      Alert.alert('提示', '请输入经验内容');
      return;
    }
    if (contentLength < 1 || contentLength > 100) {
      Alert.alert('提示', '经验内容需 1-100 字');
      return;
    }

    if (!isEditing && sourceScene === 'chat' && isPrivate) {
      Alert.alert('要匿名贡献给相似处境的人吗？', '不会公开原聊天内容、议题标题或上下文。', [
        {text: '保持私密', onPress: () => saveWithVisibility(trimmedContent, true)},
        {text: '匿名贡献', onPress: () => saveWithVisibility(trimmedContent, false)},
        {text: '取消', style: 'cancel'},
      ]);
      return;
    }

    await saveWithVisibility(trimmedContent, isPrivate);
  };

  const handleDisplayNameContinue = async () => {
    const name = displayNameDraft.trim();
    if (!name) {
      Alert.alert('提示', '请输入名字');
      return;
    }
    if (Array.from(name).length > 20) {
      Alert.alert('提示', '名字最多 20 个字');
      return;
    }

    setDisplayNameSubmitting(true);
    try {
      await updateProfile({display_name: name});
      setDisplayNameModalVisible(false);
      await persistExperience(content.trim(), displayNameRetryPrivate ?? isPrivate);
      setDisplayNameRetryPrivate(null);
    } catch (e: any) {
      await handlePersistError(e, displayNameRetryPrivate ?? isPrivate);
    } finally {
      setDisplayNameSubmitting(false);
    }
  };

  const handleRewrite = async () => {
    const trimmedContent = content.trim();
    if (!trimmedContent || rewriting || submitting) return;
    setRewriting(true);
    try {
      const result = await rewriteExperience(trimmedContent, {
        source: sourceScene === 'chat' ? 'chat_note' : 'manual_note',
        default_visibility: isPrivate ? 'private' : 'public',
        user_selected_domain: domain || undefined,
        user_selected_sub_domain: subDomain || undefined,
        topic_context: topic.trim() || undefined,
        source_message_ids: sourceMessageIds,
      });
      if (result.can_rewrite && result.rewritten_content) {
        setRewriteCandidate({
          content: result.rewritten_content,
          domain: result.domain,
          subDomain: result.sub_domain,
          topic: result.topic,
        });
      } else {
        Alert.alert('先这样记下也可以', '这段内容暂时不适合整理成经验，你可以按原文记下。');
      }
    } catch (e: any) {
      if (await handleAuthExpired(navigation, e)) return;
      Alert.alert('暂时改不了', '原文可以直接记下。');
    } finally {
      setRewriting(false);
    }
  };

  const applyRewriteCandidate = () => {
    if (!rewriteCandidate) return;
    setContent(rewriteCandidate.content);
    if (!domain && rewriteCandidate.domain) setDomain(rewriteCandidate.domain);
    if (!subDomain && rewriteCandidate.subDomain) setSubDomain(rewriteCandidate.subDomain);
    if (!topic && rewriteCandidate.topic) setTopic(rewriteCandidate.topic);
    setRewriteCandidate(null);
  };

  const trimmedContent = content.trim();
  const contentLength = Array.from(trimmedContent).length;
  const isPublishReady = !!(trimmedContent) && contentLength >= 1 && contentLength <= 100;

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Animated.View style={[styles.flex, {transform: [{translateX: slideAnim}]}]}>
      {isTopicEditing ? (
        /* ── Full-page topic editor ── */
        <View style={styles.flex}>
          <View style={styles.header}>
            <TouchableOpacity onPress={() => setIsTopicEditing(false)} style={styles.backBtn}>
              <Text style={styles.backBtnText}>取消</Text>
            </TouchableOpacity>
            <Text style={styles.headerTitle}>话题</Text>
            <TouchableOpacity
              onPress={() => {
                setTopic(topicDraft);
                setIsTopicEditing(false);
              }}
              style={styles.backBtn}>
              <Text style={styles.topicDoneText}>完成</Text>
            </TouchableOpacity>
          </View>

          <View style={styles.topicPageContainer}>
            <TextInput
              ref={topicPageInputRef}
              style={styles.topicPageInput}
              value={topicDraft}
              onChangeText={setTopicDraft}
              placeholder="#"
              placeholderTextColor="#b5b0a8"
              multiline
              maxLength={200}
              autoFocus
              textAlignVertical="top"
            />
            <TouchableOpacity
              style={styles.topicPageHashBtn}
              onPress={() => setTopicDraft(prev => prev + '#')}>
              <Text style={styles.topicPageHashBtnText}>#话题</Text>
            </TouchableOpacity>
          </View>
        </View>
      ) : (
        <>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={handleBack} style={styles.backBtn} accessibilityRole="button" accessibilityLabel="返回">
          <Ionicons name="chevron-back" size={22} color="#5c5548" />
        </TouchableOpacity>
        <Text style={styles.headerTitle}>{isEditing ? '编辑经验' : '记下'}</Text>
        <View style={styles.backBtn} />
      </View>

      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.flex}>
        <ScrollView
          style={styles.flex}
          contentContainerStyle={styles.scrollContent}
          keyboardShouldPersistTaps="handled">

          {/* Central input area */}
          <View style={styles.inputArea}>
            <TextInput
              ref={inputRef}
              style={styles.centralInput}
              value={content}
              onChangeText={setContent}
              placeholder="此刻你有什么想说的？"
              placeholderTextColor="#c5bfb3"
              multiline
              maxLength={100}
              textAlign="center"
              textAlignVertical="center"
              onFocus={() => setIsFocused(true)}
              onBlur={() => setIsFocused(false)}
            />
            <Text style={styles.charCount}>{content.length}/100</Text>
          </View>

          <>
            <Text style={styles.domainHint}>领域</Text>
            <Text style={styles.domainDash}>—</Text>

            {/* Domain row */}
            <View style={styles.domainSection}>
              <View style={styles.domainRow}>
                {PRIMARY_DOMAINS.map(d => {
                  const isSelected = domain === d.key;
                  const hasSelection = domain !== '';
                  return (
                    <TouchableOpacity
                      key={d.key}
                      style={[styles.domainChip, isSelected && styles.domainChipActive]}
                      onPress={() => handleDomainSelect(d.key)}>
                      <Text style={[
                        styles.domainChipText,
                        isSelected && styles.domainChipTextActive,
                        hasSelection && !isSelected && styles.domainChipTextDimmed,
                      ]}>
                        {d.label}
                      </Text>
                    </TouchableOpacity>
                  );
                })}
              </View>

              {/* Animated subdomain row */}
              <Animated.View
                style={[
                  styles.subdomainWrapper,
                  {
                    opacity: subdomainOpacity,
                    maxHeight: subdomainHeight.interpolate({
                      inputRange: [0, 1],
                      outputRange: [0, 100],
                    }),
                  },
                ]}>
                {domain !== '' && SUB_DOMAINS[domain] && (
                  <View style={styles.subdomainRow}>
                    {SUB_DOMAINS[domain].map(sd => {
                      const isSelected = subDomain === sd.key;
                      const hasSelection = subDomain !== '';
                      return (
                        <TouchableOpacity
                          key={sd.key}
                          style={[styles.subDomainChip, isSelected && styles.subDomainChipActive]}
                          onPress={() => setSubDomain(subDomain === sd.key ? '' : sd.key)}>
                          <Text style={[
                            styles.subDomainChipText,
                            isSelected && styles.subDomainChipTextActive,
                            hasSelection && !isSelected && styles.subDomainChipTextDimmed,
                          ]}>
                            {sd.label}
                          </Text>
                        </TouchableOpacity>
                      );
                    })}
                  </View>
                )}
              </Animated.View>
            </View>
          </>

          {/* Topic section */}
          {!isFocused && (
            <View style={styles.topicSection}>
              <Text style={styles.domainHint}>话题</Text>
              <Text style={styles.domainDash}>—</Text>

              <TouchableOpacity
                style={styles.topicBtnWrap}
                onPress={() => {
                  setTopicDraft(topic.trim() || '#');
                  setIsTopicEditing(true);
                  setTimeout(() => topicPageInputRef.current?.focus(), 150);
                }}>
                <Text style={styles.addTopicText}>
                  {topic.trim() ? topic.trim() : '#添加话题'}
                </Text>
              </TouchableOpacity>
            </View>
          )}

        </ScrollView>
      </KeyboardAvoidingView>

      {/* Bottom bar */}
      <View style={styles.bottomBar}>
        <TouchableOpacity
          style={[styles.rewriteButton, (!trimmedContent || rewriting || submitting) && styles.rewriteButtonDisabled]}
          onPress={handleRewrite}
          disabled={!trimmedContent || rewriting || submitting}>
          {rewriting ? (
            <ActivityIndicator size="small" color="#4a7c59" />
          ) : (
            <Text style={styles.rewriteButtonText}>帮我改改</Text>
          )}
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.saveButton, (!isPublishReady || submitting) && styles.saveButtonDisabled]}
          onPress={handlePublish}
          disabled={!isPublishReady || submitting}>
          {submitting ? (
            <ActivityIndicator size="small" color="#fff" />
          ) : (
            <Text style={styles.saveButtonText}>{isEditing ? '保存修改' : '保存'}</Text>
          )}
        </TouchableOpacity>
        <TouchableOpacity style={styles.privacyRow} onPress={() => setIsPrivate(!isPrivate)}>
          <View style={[styles.privacyDot, isPrivate && styles.privacyDotActive]} />
          <Text style={[styles.privacyLabel, isPrivate && styles.privacyLabelActive]}>
            {isPrivate ? '🔒 私密' : '公开'}
          </Text>
        </TouchableOpacity>
      </View>
        </>
      )}
      </Animated.View>
      <Modal visible={!!rewriteCandidate} transparent animationType="fade">
        <View style={styles.rewriteOverlay}>
          <View style={styles.rewriteModal}>
            <Text style={styles.rewriteModalTitle}>帮你改成这样</Text>
            <Text style={styles.rewriteModalContent}>{rewriteCandidate?.content}</Text>
            <View style={styles.rewriteModalActions}>
              <TouchableOpacity
                style={styles.rewriteModalSecondary}
                onPress={() => setRewriteCandidate(null)}
                activeOpacity={0.72}>
                <Text style={styles.rewriteModalSecondaryText}>取消</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={styles.rewriteModalPrimary}
                onPress={applyRewriteCandidate}
                activeOpacity={0.78}>
                <Text style={styles.rewriteModalPrimaryText}>替换原文</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
      <Modal visible={displayNameModalVisible} transparent animationType="fade">
        <View style={styles.rewriteOverlay}>
          <View style={styles.displayNameModal}>
            <Text style={styles.displayNameTitle}>先取个名字</Text>
            <Text style={styles.displayNameHint}>公开经验会显示这个名字。</Text>
            <TextInput
              style={styles.displayNameInput}
              value={displayNameDraft}
              onChangeText={setDisplayNameDraft}
              placeholder="别人会在经验卡上看到这个名字"
              placeholderTextColor="#b5b0a8"
              maxLength={20}
              autoFocus
            />
            <View style={styles.rewriteModalActions}>
              <TouchableOpacity
                style={styles.rewriteModalSecondary}
                onPress={() => {
                  setDisplayNameModalVisible(false);
                  setDisplayNameRetryPrivate(null);
                }}
                activeOpacity={0.72}
                disabled={displayNameSubmitting}>
                <Text style={styles.rewriteModalSecondaryText}>取消</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.rewriteModalPrimary, (!displayNameDraft.trim() || displayNameSubmitting) && styles.saveButtonDisabled]}
                onPress={handleDisplayNameContinue}
                activeOpacity={0.78}
                disabled={!displayNameDraft.trim() || displayNameSubmitting}>
                {displayNameSubmitting ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text style={styles.rewriteModalPrimaryText}>保存并继续</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#faf8f5',
  },
  flex: {
    flex: 1,
  },
  // Header
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  backBtn: {
    width: 44,
    minHeight: 36,
    justifyContent: 'center',
  },
  backBtnText: {
    fontSize: 15,
    color: '#9a9a9a',
  },
  headerTitle: {
    fontSize: 13,
    color: '#a0a0a0',
  },
  // Central input
  scrollContent: {
    flexGrow: 1,
    paddingHorizontal: 24,
    paddingBottom: 24,
  },
  inputArea: {
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: 200,
    paddingTop: 140,
  },
  centralInput: {
    width: '100%',
    fontSize: 18,
    lineHeight: 30,
    color: '#1a1a1a',
    minHeight: 100,
    paddingHorizontal: 8,
    paddingVertical: 16,
  },
  charCount: {
    fontSize: 12,
    color: '#c5bfb3',
    marginTop: 4,
    marginBottom: 40,
  },
  domainHint: {
    fontSize: 11,
    color: '#b5b0a8',
    textAlign: 'center',
    marginTop: 8,
  },
  domainDash: {
    fontSize: 13,
    color: '#d5d0c8',
    textAlign: 'center',
    marginTop: 2,
    marginBottom: 6,
  },
  // Domain
  domainSection: {
    alignItems: 'center',
  },
  domainRow: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'flex-end',
    width: '70%',
    alignSelf: 'center',
    gap: 10,
  },
  domainChip: {
    paddingHorizontal: 4,
    paddingVertical: 4,
  },
  domainChipActive: {
    // no background — just text color changes
  },
  domainChipText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#8a8a8a',
  },
  domainChipTextActive: {
    color: '#4a7c59',
    fontWeight: '700',
    fontSize: 15,
  },
  domainChipTextDimmed: {
    color: '#cdc8c0',
  },
  // Domain
  subdomainWrapper: {
    overflow: 'hidden',
    width: '100%',
  },
  subdomainRow: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'flex-end',
    width: '70%',
    alignSelf: 'center',
    gap: 10,
    paddingTop: 8,
    paddingBottom: 4,
  },
  subDomainChip: {
    paddingHorizontal: 4,
    paddingVertical: 4,
  },
  subDomainChipActive: {
    // no background — just text color changes
  },
  subDomainChipText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#8a8a8a',
  },
  subDomainChipTextActive: {
    color: '#4a7c59',
    fontWeight: '700',
    fontSize: 15,
  },
  subDomainChipTextDimmed: {
    color: '#cdc8c0',
  },
  // Bottom bar
  bottomBar: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    paddingBottom: 28,
    backgroundColor: '#faf8f5',
  },
  rewriteButton: {
    borderRadius: 8,
    paddingVertical: 12,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 42,
    width: '100%',
    marginBottom: 10,
    backgroundColor: '#f5faf2',
    borderWidth: 1,
    borderColor: '#dbe7d8',
  },
  rewriteButtonDisabled: {
    opacity: 0.45,
  },
  rewriteButtonText: {
    color: '#4a7c59',
    fontSize: 15,
    fontWeight: '700',
  },
  rewriteOverlay: {
    flex: 1,
    backgroundColor: 'rgba(25,24,20,0.34)',
    justifyContent: 'center',
    paddingHorizontal: 28,
  },
  rewriteModal: {
    backgroundColor: '#fffaf0',
    borderRadius: 18,
    padding: 18,
    borderWidth: 1,
    borderColor: '#eadfcd',
  },
  displayNameModal: {
    backgroundColor: '#fffaf0',
    borderRadius: 16,
    padding: 18,
    borderWidth: 1,
    borderColor: '#eadfcd',
  },
  displayNameTitle: {
    fontSize: 16,
    fontWeight: '800',
    color: '#2a2722',
  },
  displayNameHint: {
    marginTop: 6,
    fontSize: 13,
    lineHeight: 19,
    color: '#8f887b',
  },
  displayNameInput: {
    minHeight: 46,
    marginTop: 16,
    paddingHorizontal: 12,
    borderRadius: 10,
    borderWidth: 1,
    borderColor: '#ded2c0',
    backgroundColor: '#fffdf8',
    color: '#24231f',
    fontSize: 15,
  },
  rewriteModalTitle: {
    fontSize: 15,
    fontWeight: '800',
    color: '#2a2722',
    marginBottom: 12,
  },
  rewriteModalContent: {
    fontSize: 18,
    lineHeight: 28,
    color: '#1f211c',
    fontWeight: '700',
  },
  rewriteModalActions: {
    flexDirection: 'row',
    gap: 10,
    marginTop: 18,
  },
  rewriteModalSecondary: {
    flex: 1,
    minHeight: 44,
    borderRadius: 11,
    borderWidth: 1,
    borderColor: '#ded2c0',
    justifyContent: 'center',
    alignItems: 'center',
  },
  rewriteModalSecondaryText: {
    color: '#6d6256',
    fontSize: 14,
    fontWeight: '700',
  },
  rewriteModalPrimary: {
    flex: 1,
    minHeight: 44,
    borderRadius: 11,
    backgroundColor: '#4a7c59',
    justifyContent: 'center',
    alignItems: 'center',
  },
  rewriteModalPrimaryText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '800',
  },
  saveButton: {
    backgroundColor: '#4a7c59',
    borderRadius: 8,
    paddingVertical: 14,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 46,
    width: '100%',
  },
  saveButtonDisabled: {
    backgroundColor: '#c5d4c9',
  },
  saveButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '600',
  },
  privacyRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 10,
    gap: 6,
    paddingVertical: 4,
  },
  privacyDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
    backgroundColor: '#d5d0c8',
  },
  privacyDotActive: {
    backgroundColor: '#4a7c59',
  },
  privacyLabel: {
    fontSize: 11,
    color: '#b0b0b0',
  },
  privacyLabelActive: {
    color: '#4a7c59',
    fontWeight: '600',
  },
  // Topic
  topicSection: {
    marginTop: 20,
  },
  topicBtnWrap: {
    alignItems: 'center',
  },
  addTopicText: {
    fontSize: 13,
    color: '#8a8a8a',
    fontWeight: '500',
  },
  // Topic page (full-screen editor)
  topicDoneText: {
    fontSize: 15,
    color: '#4a7c59',
    fontWeight: '600',
    textAlign: 'right',
  },
  topicPageContainer: {
    flex: 1,
    paddingHorizontal: 24,
    paddingTop: 100,
  },
  topicPageInput: {
    fontSize: 18,
    color: '#1a1a1a',
    minHeight: 120,
    lineHeight: 30,
    textAlign: 'center',
  },
  topicPageHashBtn: {
    alignSelf: 'center',
    marginTop: 16,
    paddingHorizontal: 12,
    paddingVertical: 6,
  },
  topicPageHashBtnText: {
    fontSize: 13,
    color: '#4a7c59',
    fontWeight: '500',
  },
});
