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
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {initChat, sendChatMessage, ChatMessageItem} from '../services/api';

interface MessageBubble {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  references?: string[];
}

export default function ChatScreen() {
  const [conversationId, setConversationId] = useState<string>('');
  const [messages, setMessages] = useState<MessageBubble[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const flatListRef = useRef<FlatList>(null);

  // 初始化：加载历史消息 + 自动打招呼
  useEffect(() => {
    initChat()
      .then(data => {
        setConversationId(data.conversation_id);
        const msgs: MessageBubble[] = (data.messages || []).map(
          (m: ChatMessageItem) => ({
            id: m.id,
            role: m.role,
            content: m.content,
            references: m.referenced_experience_ids?.length
              ? m.referenced_experience_ids
              : undefined,
          }),
        );
        setMessages(msgs);
      })
      .catch(err => {
        console.warn('[chat] init failed:', err?.message);
        // Fallback: show welcome if history load fails
        setMessages([
          {
            id: 'welcome',
            role: 'assistant',
            content: '嗨，我是年糕。想聊什么都可以，随便说说。',
          },
        ]);
      })
      .finally(() => setInitialLoading(false));
  }, []);

  const handleSend = async () => {
    const text = input.trim();
    if (!text || loading || !conversationId) return;
    setInput('');

    const userMsg: MessageBubble = {
      id: Date.now().toString(),
      role: 'user',
      content: text,
    };
    setMessages(prev => [...prev, userMsg]);
    setLoading(true);

    // 占位气泡
    const aiId = (Date.now() + 1).toString();
    setMessages(prev => [...prev, {id: aiId, role: 'assistant', content: ''}]);

    try {
      const result = await sendChatMessage(conversationId, text);

      setMessages(prev =>
        prev.map(m =>
          m.id === aiId
            ? {
                ...m,
                content: result.reply,
                references: result.referenced_experience_ids?.length
                  ? result.referenced_experience_ids
                  : undefined,
              }
            : m,
        ),
      );
    } catch (e: any) {
      let errMsg = '抱歉，对话服务暂时不可用，请稍后再试。';
      if (e?.status === 429) {
        errMsg = '今日对话已达上限（100轮），明天再来聊吧。';
      }
      setMessages(prev =>
        prev.map(m =>
          m.id === aiId ? {...m, content: errMsg} : m,
        ),
      );
    } finally {
      setLoading(false);
    }
  };

  const renderMessage = ({item}: {item: MessageBubble}) => {
    if (item.role === 'system') return null;
    const isUser = item.role === 'user';

    return (
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
          {item.references && item.references.length > 0 && (
            <View style={styles.referenceBar}>
              <Text style={styles.referenceText}>
                引用了你收藏的 {item.references.length} 条经验
              </Text>
            </View>
          )}
        </View>
        {isUser && (
          <View style={styles.userAvatar}>
            <Text style={styles.userAvatarText}>我</Text>
          </View>
        )}
      </View>
    );
  };

  if (initialLoading) {
    return (
      <SafeAreaView style={styles.container} edges={['top']}>
        <View style={styles.header}>
          <Text style={styles.headerTitle}>随便聊聊</Text>
          <Text style={styles.headerSub}>收藏的经验我都记着</Text>
        </View>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="small" color="#4a7c59" />
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>随便聊聊</Text>
        <Text style={styles.headerSub}>收藏的经验我都记着</Text>
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
        keyboardVerticalOffset={Platform.OS === 'ios' ? 90 : 0}>
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
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#faf8f5',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  header: {
    paddingHorizontal: 18,
    paddingVertical: 12,
    borderBottomWidth: 0.5,
    borderBottomColor: '#e8e4df',
    backgroundColor: '#faf8f5',
  },
  headerTitle: {
    fontSize: 17,
    fontWeight: '700',
    color: '#1a1a1a',
  },
  headerSub: {
    fontSize: 12,
    color: '#9a9a9a',
    marginTop: 2,
  },
  messageList: {
    paddingHorizontal: 14,
    paddingVertical: 12,
    flexGrow: 1,
  },
  bubbleRow: {
    flexDirection: 'row',
    marginBottom: 14,
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
  bubbleTextUser: {
    color: '#ffffff',
  },
  referenceBar: {
    marginTop: 8,
    paddingTop: 8,
    borderTopWidth: 0.5,
    borderTopColor: '#e8e4df',
  },
  referenceText: {
    fontSize: 11,
    color: '#9a9a9a',
  },
  inputBar: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    paddingHorizontal: 14,
    paddingVertical: 10,
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
});
