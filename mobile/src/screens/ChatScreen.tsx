import React, {useState, useRef, useCallback, useEffect} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  FlatList,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  ActivityIndicator,
  Animated,
  Dimensions,
  Modal,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {
  createChatTempSession,
  fetchRecentChatTopics,
  fetchChatTopicMessages,
  sendChatTopicMessage,
  sendTempChatMessage,
  setCollected,
  recordExperienceEvent,
  ChatTopic,
  ChatReferenceCard,
  ChatNoteSuggestion,
  ChatMessageItem,
} from '../services/api';
import {clearToken} from '../services/config';
import {reportHandledError} from '../utils/logging';
import Ionicons from '@expo/vector-icons/Ionicons';

interface MessageBubble {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  referenceCards?: ChatReferenceCard[];
  noteSuggestion?: ChatNoteSuggestion;
  failed?: boolean;
  retryText?: string;
  clientMessageId?: string;
}

export default function ChatScreen({navigation}: any) {
  const [tempSessionId, setTempSessionId] = useState<string>('');
  const [activeTopic, setActiveTopic] = useState<ChatTopic | null>(null);
  const [topics, setTopics] = useState<ChatTopic[]>([]);
  const [topicModalVisible, setTopicModalVisible] = useState(false);
  const [topicLoading, setTopicLoading] = useState(false);
  const [messages, setMessages] = useState<MessageBubble[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const flatListRef = useRef<FlatList>(null);
  const slideAnim = useRef(new Animated.Value(0)).current;
  const screenWidth = Dimensions.get('window').width;

  const openLogin = useCallback(() => {
    const parent = navigation.getParent?.();
    if (parent?.navigate) {
      parent.navigate('login');
      return;
    }
    navigation.navigate('login');
  }, [navigation]);

  const handleAuthExpired = useCallback((err: any): boolean => {
    if (err?.status !== 401) return false;
    clearToken().catch(clearErr => reportHandledError('ChatScreen.clearExpiredAuth', clearErr));
    openLogin();
    return true;
  }, [openLogin]);

  const handleBack = () => {
    Animated.timing(slideAnim, {
      toValue: -screenWidth,
      duration: 220,
      useNativeDriver: true,
    }).start(() => navigation.navigate('home'));
  };

  // Reset slide position when screen gains focus
  useEffect(() => {
    const unsub = navigation.addListener('focus', () => {
      slideAnim.setValue(0);
    });
    return unsub;
  }, [navigation, slideAnim]);

  const startTempSession = useCallback(() => {
    setInitialLoading(true);
    setActiveTopic(null);
    return createChatTempSession(false)
      .then(session => {
        setTempSessionId(session.id);
        setMessages([
          {
            id: 'welcome',
            role: 'assistant',
            content: '我在。你可以从任何一点开始说，不用先想清楚。',
          },
        ]);
      })
      .catch(err => {
        if (handleAuthExpired(err)) {
          setMessages([
            {
              id: 'auth-expired',
              role: 'assistant',
              content: '登录状态过期了，重新登录后可以继续聊。',
            },
          ]);
          return;
        }
        reportHandledError('ChatScreen.startTempSession', err);
        setMessages([
          {
            id: 'welcome',
            role: 'assistant',
            content: '现在连不上聊聊服务。你可以先把想说的放在这里，稍后再试。',
          },
        ]);
      })
      .finally(() => setInitialLoading(false));
  }, [handleAuthExpired]);

  // 初始化：进入聊聊先创建临时会话；用户发消息后再判断是否形成稳定议题
  useEffect(() => {
    startTempSession();
  }, [startTempSession]);

  const openTopicList = async () => {
    setTopicModalVisible(true);
    setTopicLoading(true);
    try {
      const result = await fetchRecentChatTopics();
      setTopics(result.data || []);
    } catch (err: any) {
      if (handleAuthExpired(err)) return;
      reportHandledError('ChatScreen.openTopicList', err);
      setTopics([]);
    } finally {
      setTopicLoading(false);
    }
  };

  const selectTopic = async (topic: ChatTopic) => {
    setTopicModalVisible(false);
    setTopicLoading(true);
    try {
      const result = await fetchChatTopicMessages(topic.id);
      setActiveTopic(topic);
      setMessages((result.data || []).map(toBubble));
    } catch (err: any) {
      if (handleAuthExpired(err)) return;
      reportHandledError('ChatScreen.selectTopic', err);
      setActiveTopic(topic);
      setMessages([{
        id: 'topic-load-failed',
        role: 'assistant',
        content: '这个议题的消息暂时没恢复出来，可以先接着说。',
      }]);
    } finally {
      setTopicLoading(false);
    }
  };

  const handleStartNewTopic = () => {
    setTopicModalVisible(false);
    startTempSession();
  };

  const toBubble = (message: ChatMessageItem): MessageBubble => ({
    id: message.id,
    role: message.role,
    content: message.content,
    referenceCards: message.reference_cards || [],
  });

  const sendToCurrentScope = async (text: string, clientMessageId: string) => {
    if (activeTopic) {
      return sendChatTopicMessage(activeTopic.id, text, clientMessageId);
    }
    return sendTempChatMessage(tempSessionId, text, clientMessageId);
  };

  const citationMetadata = (messageId: string, topicOverride?: ChatTopic | null) => ({
    message_id: messageId,
    ...(topicOverride?.id ? {topic_id: topicOverride.id} : activeTopic ? {topic_id: activeTopic.id} : {temp_session_id: tempSessionId}),
  });

  const recordCitationShows = (messageId: string, cards: ChatReferenceCard[] = [], topicOverride?: ChatTopic | null) => {
    cards.forEach(card => {
      recordExperienceEvent(card.experience_id, 'chat_citation_show', 'chat', citationMetadata(messageId, topicOverride));
    });
  };

  const handleReferencePress = (messageId: string, experienceId: string) => {
    recordExperienceEvent(experienceId, 'chat_citation_click', 'chat', citationMetadata(messageId));
    navigation.navigate('detail', {id: experienceId, from: 'chat'});
  };

  const applyPromotedTopic = (topic?: ChatTopic) => {
    if (!topic?.id) return null;
    setActiveTopic(topic);
    setTempSessionId('');
    return topic;
  };

  const openNoteSuggestion = (item: MessageBubble) => {
    const sourceMessageIds = item.noteSuggestion?.source_message_ids || [];
    const params: Record<string, unknown> = {
      prefillContent: item.noteSuggestion?.suggested_text || '',
      defaultVisibility: 'private',
      sourceScene: 'chat',
      sourceMessageIds,
      sourceChatMessageId: item.id,
    };
    if (sourceMessageIds.length > 0) {
      params.sourceChatMessageSnapshot = sourceMessageIds.join(',');
    }
    if (activeTopic?.id) {
      params.sourceChatTopicId = activeTopic.id;
    }
    navigation.navigate('create', params);
  };

  const handleSend = async () => {
    const text = input.trim();
    if (!text || loading || (!activeTopic && !tempSessionId)) return;
    setInput('');

    const clientMessageId = `m-${Date.now()}`;
    const userMsg: MessageBubble = {
      id: clientMessageId,
      role: 'user',
      content: text,
    };
    setMessages(prev => [...prev, userMsg]);
    setLoading(true);

    // 占位气泡
    const aiId = (Date.now() + 1).toString();
    setMessages(prev => [...prev, {id: aiId, role: 'assistant', content: ''}]);

    try {
      const result = await sendToCurrentScope(text, clientMessageId);
      const referenceCards = result.reference_cards || [];
      const promotedTopic = applyPromotedTopic(result.promoted_topic);
      recordCitationShows(result.message.id || aiId, referenceCards, promotedTopic);

      setMessages(prev =>
        prev.map(m =>
          m.id === aiId
            ? {
                ...m,
                id: result.message.id || aiId,
                content: result.message.content,
                referenceCards,
                noteSuggestion: result.note_suggestion,
              }
            : m,
        ),
      );
    } catch (e: any) {
      if (handleAuthExpired(e)) {
        setMessages(prev =>
          prev.map(m =>
            m.id === aiId
              ? {
                  ...m,
                  content: '登录状态过期了，重新登录后可以继续聊。',
                  failed: false,
                  retryText: undefined,
                  clientMessageId: undefined,
                }
              : m,
          ),
        );
        return;
      }
      let errMsg = '抱歉，对话服务暂时不可用，请稍后再试。';
      if (e?.status === 429) {
        errMsg = '今日对话已达上限（100轮），明天再来聊吧。';
      }
      setMessages(prev =>
        prev.map(m =>
          m.id === aiId ? {...m, content: errMsg, failed: true, retryText: text, clientMessageId} : m,
        ),
      );
    } finally {
      setLoading(false);
    }
  };

  const retryAssistant = async (item: MessageBubble) => {
    if (!item.retryText || loading || (!activeTopic && !tempSessionId)) return;
    setLoading(true);
    setMessages(prev => prev.map(m => m.id === item.id ? {...m, content: '', failed: false} : m));
    try {
      const result = await sendToCurrentScope(item.retryText, item.clientMessageId || `retry-${Date.now()}`);
      const referenceCards = result.reference_cards || [];
      const promotedTopic = applyPromotedTopic(result.promoted_topic);
      recordCitationShows(result.message.id || item.id, referenceCards, promotedTopic);
      setMessages(prev =>
        prev.map(m =>
          m.id === item.id
            ? {
                ...m,
                id: result.message.id || item.id,
                content: result.message.content,
                referenceCards,
                noteSuggestion: result.note_suggestion,
                failed: false,
                retryText: undefined,
                clientMessageId: undefined,
              }
            : m,
        ),
      );
    } catch (err: any) {
      if (handleAuthExpired(err)) {
        setMessages(prev => prev.map(m => m.id === item.id ? {
          ...m,
          content: '登录状态过期了，重新登录后可以继续聊。',
          failed: false,
          retryText: undefined,
          clientMessageId: undefined,
        } : m));
        return;
      }
      setMessages(prev => prev.map(m => m.id === item.id ? {
        ...m,
        content: '还是没连上。你这条消息已经保留了，可以稍后再试。',
        failed: true,
      } : m));
    } finally {
      setLoading(false);
    }
  };

  const collectReferenceCard = async (messageId: string, experienceId: string) => {
    setMessages(prev => prev.map(message => {
      if (message.id !== messageId || !message.referenceCards) return message;
      return {
        ...message,
        referenceCards: message.referenceCards.map(card =>
          card.experience_id === experienceId ? {...card, is_collected: true} : card,
        ),
      };
    }));
    try {
      await setCollected(experienceId, true);
    } catch (err: any) {
      setMessages(prev => prev.map(message => {
        if (message.id !== messageId || !message.referenceCards) return message;
        return {
          ...message,
          referenceCards: message.referenceCards.map(card =>
            card.experience_id === experienceId ? {...card, is_collected: false} : card,
          ),
        };
      }));
      handleAuthExpired(err);
    }
  };

  const renderMessage = ({item}: {item: MessageBubble}) => {
    if (item.role === 'system') return null;
    const isUser = item.role === 'user';

    return (
      <View style={[styles.messageBlock, isUser && styles.messageBlockUser]}>
      <View style={[styles.bubbleRow, isUser && styles.bubbleRowUser]}>
        {!isUser && (
          <View style={styles.aiAvatar}>
            <Text style={styles.aiAvatarText}>糕</Text>
          </View>
        )}
        <View
          style={[
            styles.bubble,
            isUser ? styles.bubbleUser : styles.bubbleAi,
          ]}>
          <Text
            style={[
              styles.bubbleText,
              isUser && styles.bubbleTextUser,
            ]}>
            {item.content || '…'}
          </Text>
          {!isUser && loading && item.content === '' && (
            <ActivityIndicator
              size="small"
              color="#4a7c59"
              style={{marginTop: 4}}
            />
          )}
          {!isUser && item.failed && (
            <TouchableOpacity style={styles.retryBtn} onPress={() => retryAssistant(item)}>
              <Text style={styles.retryText}>重试</Text>
            </TouchableOpacity>
          )}
        </View>
        {isUser && (
          <View style={styles.userAvatar}>
            <Text style={styles.userAvatarText}>我</Text>
          </View>
        )}
      </View>
      {!isUser && item.referenceCards && item.referenceCards.length > 0 && (
        <View style={styles.referenceWrap}>
          <Text style={styles.referenceTitle}>参考经验</Text>
          {item.referenceCards.map(card => {
            const isUnavailable = Boolean(card.unavailable_reason);
            return (
              <TouchableOpacity
                key={card.experience_id}
                style={[styles.referenceCard, isUnavailable && styles.referenceCardUnavailable]}
                onPress={() => {
                  if (!isUnavailable) handleReferencePress(item.id, card.experience_id);
                }}
                accessibilityRole="button"
                accessibilityLabel={isUnavailable ? '不可见参考经验' : '参考经验'}
                activeOpacity={0.76}>
                {isUnavailable ? (
                  <View style={styles.referenceUnavailableBody}>
                    <Text style={styles.referenceUnavailableTitle}>该经验已不可见</Text>
                    <Text style={styles.referenceUnavailableText}>
                      它可能已经被删除、转为私密，或正在重新处理。
                    </Text>
                  </View>
                ) : (
                  <>
                    <Text style={styles.referenceText}>{card.content}</Text>
                    <TouchableOpacity
                      style={styles.referenceSaveBtn}
                      onPress={() => {
                        if (!card.is_collected) collectReferenceCard(item.id, card.experience_id);
                      }}
                      accessibilityRole="button"
                      accessibilityLabel={card.is_collected ? '已收藏参考经验' : '收藏参考经验'}
                      activeOpacity={0.72}>
                      <Ionicons
                        name={card.is_collected ? 'bookmark' : 'bookmark-outline'}
                        size={18}
                        color={card.is_collected ? '#d59a3d' : '#8a8173'}
                      />
                    </TouchableOpacity>
                  </>
                )}
              </TouchableOpacity>
            );
          })}
        </View>
      )}
      {!isUser && item.noteSuggestion?.should_show && (
        <View style={styles.noteSuggestion}>
          <Text style={styles.noteSuggestionText}>
            你刚才的思考很适合总结记下一条经验，要记下吗？
          </Text>
          <TouchableOpacity
            style={styles.noteSuggestionBtn}
            onPress={() => openNoteSuggestion(item)}>
            <Text style={styles.noteSuggestionBtnText}>记下</Text>
          </TouchableOpacity>
        </View>
      )}
      </View>
    );
  };

  if (initialLoading) {
    return (
      <SafeAreaView style={styles.container} edges={['top']}>
        <Animated.View style={[styles.flex, {transform: [{translateX: slideAnim}]}]}>
        <View style={styles.header}>
          <TouchableOpacity onPress={handleBack} style={styles.backBtn} accessibilityRole="button" accessibilityLabel="返回">
            <Ionicons name="chevron-back" size={22} color="#5c5548" />
          </TouchableOpacity>
          <View style={styles.headerCenter}>
            <Text style={styles.headerTitle}>和年糕聊聊</Text>
          </View>
          <TouchableOpacity style={styles.topicBtn} onPress={openTopicList} accessibilityRole="button" accessibilityLabel="打开议题列表">
            <Ionicons name="chatbubble-ellipses-outline" size={14} color="#4a7c59" />
            <Text style={styles.topicBtnText} numberOfLines={1}>
              {activeTopic?.title || '当前议题'}
            </Text>
          </TouchableOpacity>
        </View>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="small" color="#4a7c59" />
        </View>
        </Animated.View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <Animated.View style={[styles.flex, {transform: [{translateX: slideAnim}]}]}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={handleBack} style={styles.backBtn} accessibilityRole="button" accessibilityLabel="返回">
          <Ionicons name="chevron-back" size={22} color="#5c5548" />
        </TouchableOpacity>
        <View style={styles.headerCenter}>
          <Text style={styles.headerTitle}>和年糕聊聊</Text>
        </View>
        <TouchableOpacity style={styles.topicBtn} onPress={openTopicList} accessibilityRole="button" accessibilityLabel="打开议题列表">
          <Ionicons name="chatbubble-ellipses-outline" size={14} color="#4a7c59" />
          <Text style={styles.topicBtnText} numberOfLines={1}>
            {activeTopic?.title || '当前议题'}
          </Text>
        </TouchableOpacity>
      </View>

      {/* Messages */}
      <FlatList
        ref={flatListRef}
        data={messages}
        keyExtractor={item => item.id}
        renderItem={renderMessage}
        contentContainerStyle={styles.messageList}
        onContentSizeChange={() =>
          flatListRef.current?.scrollToEnd({animated: true})
        }
        showsVerticalScrollIndicator={false}
      />

      {/* Input */}
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        keyboardVerticalOffset={40}>
        <View style={styles.inputBar}>
          <TextInput
            style={styles.input}
            value={input}
            onChangeText={setInput}
            placeholder="输入你想聊的..."
            placeholderTextColor="#b5b0a8"
            multiline
            maxLength={500}
            editable={!loading}
            returnKeyType="send"
            blurOnSubmit={false}
            onSubmitEditing={handleSend}
          />
          <TouchableOpacity
            style={[
              styles.sendButton,
              (!input.trim() || loading) && styles.sendButtonDisabled,
            ]}
            onPress={handleSend}
            disabled={!input.trim() || loading}>
            <Text style={styles.sendButtonText}>发送</Text>
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>
      </Animated.View>
      <Modal visible={topicModalVisible} transparent animationType="fade">
        <View style={styles.topicOverlay}>
          <View style={styles.topicPanel}>
            <View style={styles.topicPanelHead}>
              <Text style={styles.topicPanelTitle}>最近聊过</Text>
              <TouchableOpacity onPress={() => setTopicModalVisible(false)}>
                <Text style={styles.topicClose}>关闭</Text>
              </TouchableOpacity>
            </View>
            {topicLoading ? (
              <ActivityIndicator color="#4a7c59" style={{marginVertical: 24}} />
            ) : topics.length > 0 ? (
              topics.map(topic => (
                <TouchableOpacity
                  key={topic.id}
                  style={styles.topicItem}
                  onPress={() => selectTopic(topic)}
                  activeOpacity={0.72}>
                  <Text style={styles.topicItemTitle} numberOfLines={1}>{topic.title || '未命名议题'}</Text>
                  <Text style={styles.topicItemMeta} numberOfLines={1}>
                    {[topic.domain, topic.sub_domain, topic.topic].filter(Boolean).join(' · ') || '私密议题'}
                  </Text>
                </TouchableOpacity>
              ))
            ) : (
              <Text style={styles.topicEmpty}>还没有稳定议题。</Text>
            )}
            <TouchableOpacity style={styles.newTopicBtn} onPress={handleStartNewTopic}>
              <Text style={styles.newTopicText}>换个事聊</Text>
            </TouchableOpacity>
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
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 12,
    borderBottomWidth: 0.5,
    borderBottomColor: '#e8e4df',
    backgroundColor: '#faf8f5',
  },
  backBtn: {
    width: 36,
    height: 36,
    justifyContent: 'center',
    alignItems: 'center',
  },
  headerCenter: {
    flex: 1,
    alignItems: 'flex-start',
  },
  headerTitle: {
    fontSize: 19,
    fontWeight: '700',
    color: '#1a1a1a',
  },
  topicBtn: {
    minHeight: 34,
    paddingHorizontal: 10,
    borderRadius: 10,
    backgroundColor: '#f0ece3',
    justifyContent: 'center',
    alignItems: 'center',
    flexDirection: 'row',
    gap: 5,
    maxWidth: 138,
  },
  topicBtnText: {
    fontSize: 12,
    color: '#4a7c59',
    fontWeight: '700',
    flexShrink: 1,
  },
  messageList: {
    paddingHorizontal: 14,
    paddingVertical: 12,
    flexGrow: 1,
  },
  messageBlock: {
    marginBottom: 14,
  },
  messageBlockUser: {
    alignItems: 'flex-end',
  },
  bubbleRow: {
    flexDirection: 'row',
    marginBottom: 0,
    alignItems: 'flex-end',
    maxWidth: '85%',
  },
  bubbleRowUser: {
    alignSelf: 'flex-end',
    maxWidth: '85%',
  },
  aiAvatar: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: '#eaf2e8',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 8,
  },
  aiAvatarText: {
    fontSize: 13,
    fontWeight: '700',
    color: '#4a7c59',
  },
  userAvatar: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: '#d4d4d4',
    justifyContent: 'center',
    alignItems: 'center',
    marginLeft: 8,
  },
  userAvatarText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#6e6e6e',
  },
  bubble: {
    borderRadius: 18,
    paddingHorizontal: 14,
    paddingVertical: 10,
    maxWidth: '100%',
  },
  bubbleAi: {
    backgroundColor: '#ffffff',
    borderWidth: 0.5,
    borderColor: '#f0ece7',
    borderBottomLeftRadius: 4,
  },
  bubbleUser: {
    backgroundColor: '#4a7c59',
    borderBottomRightRadius: 4,
  },
  bubbleText: {
    fontSize: 15,
    lineHeight: 22,
    color: '#1a1a1a',
  },
  retryBtn: {
    marginTop: 8,
    alignSelf: 'flex-start',
    borderRadius: 9,
    backgroundColor: '#edf4e9',
    paddingHorizontal: 10,
    paddingVertical: 6,
  },
  retryText: {
    fontSize: 12,
    color: '#4a7c59',
    fontWeight: '800',
  },
  referenceWrap: {
    marginLeft: 36,
    marginTop: 8,
    width: '78%',
  },
  referenceTitle: {
    fontSize: 11,
    color: '#8a8173',
    fontWeight: '700',
    marginBottom: 6,
  },
  referenceCard: {
    borderRadius: 10,
    borderWidth: 1,
    borderColor: '#e4ded2',
    backgroundColor: '#fffdf8',
    paddingHorizontal: 12,
    paddingVertical: 10,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  referenceCardUnavailable: {
    borderColor: '#e6e2d8',
    backgroundColor: '#f4f1ea',
  },
  referenceUnavailableBody: {
    flex: 1,
  },
  referenceUnavailableTitle: {
    fontSize: 13,
    lineHeight: 19,
    color: '#4e5348',
    fontWeight: '800',
  },
  referenceUnavailableText: {
    marginTop: 3,
    fontSize: 12,
    lineHeight: 18,
    color: '#7a7f72',
  },
  referenceText: {
    flex: 1,
    fontSize: 13,
    lineHeight: 20,
    color: '#3d3a34',
  },
  referenceSaveBtn: {
    minWidth: 42,
    minHeight: 36,
    justifyContent: 'center',
    alignItems: 'flex-end',
  },
  noteSuggestion: {
    alignSelf: 'center',
    marginTop: 12,
    maxWidth: '88%',
    borderRadius: 12,
    backgroundColor: '#edf4e9',
    paddingHorizontal: 14,
    paddingVertical: 12,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  noteSuggestionText: {
    flex: 1,
    fontSize: 13,
    lineHeight: 19,
    color: '#3d5d43',
  },
  noteSuggestionBtn: {
    borderRadius: 9,
    backgroundColor: '#4a7c59',
    paddingHorizontal: 12,
    paddingVertical: 7,
  },
  noteSuggestionBtnText: {
    fontSize: 12,
    color: '#fff',
    fontWeight: '700',
  },
  bubbleTextUser: {
    color: '#ffffff',
  },
  inputBar: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    paddingHorizontal: 14,
    paddingVertical: 20,
    backgroundColor: '#ffffff',
    borderTopWidth: 0.5,
    borderTopColor: '#e8e4df',
  },
  input: {
    flex: 1,
    backgroundColor: '#f5f3f0',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 10,
    fontSize: 15,
    color: '#1a1a1a',
    maxHeight: 100,
  },
  sendButton: {
    marginLeft: 10,
    backgroundColor: '#4a7c59',
    borderRadius: 20,
    paddingHorizontal: 18,
    paddingVertical: 10,
    marginBottom: 2,
  },
  sendButtonDisabled: {
    backgroundColor: '#c5d4c9',
  },
  sendButtonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '600',
  },
  topicOverlay: {
    flex: 1,
    backgroundColor: 'rgba(31,29,24,0.28)',
    justifyContent: 'flex-end',
  },
  topicPanel: {
    backgroundColor: '#fffaf0',
    borderTopLeftRadius: 18,
    borderTopRightRadius: 18,
    paddingHorizontal: 18,
    paddingTop: 16,
    paddingBottom: 26,
    borderWidth: 1,
    borderColor: '#eadfcd',
  },
  topicPanelHead: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 12,
  },
  topicPanelTitle: {
    fontSize: 17,
    fontWeight: '800',
    color: '#211f1a',
  },
  topicClose: {
    fontSize: 13,
    color: '#8a8173',
    fontWeight: '700',
  },
  topicItem: {
    borderRadius: 12,
    backgroundColor: '#f3eadc',
    paddingHorizontal: 14,
    paddingVertical: 12,
    marginBottom: 8,
  },
  topicItemTitle: {
    fontSize: 15,
    color: '#211f1a',
    fontWeight: '800',
  },
  topicItemMeta: {
    fontSize: 12,
    color: '#8a8173',
    marginTop: 4,
    fontWeight: '600',
  },
  topicEmpty: {
    fontSize: 13,
    color: '#8a8173',
    textAlign: 'center',
    paddingVertical: 18,
  },
  newTopicBtn: {
    minHeight: 44,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#d9ccb9',
    justifyContent: 'center',
    alignItems: 'center',
    marginTop: 4,
  },
  newTopicText: {
    fontSize: 14,
    color: '#4a7c59',
    fontWeight: '800',
  },
});
