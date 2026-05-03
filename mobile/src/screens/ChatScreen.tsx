import React, {useState, useRef, useCallback} from 'react';
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
import {sendMessage, ChatMessage} from '../services/api';

// 临时：登录未实现前使用系统用户 ID
const TEMP_USER_ID = '00000000-0000-0000-0000-000000000000';

interface MessageBubble {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  references?: string[];
}

export default function ChatScreen() {
  const [messages, setMessages] = useState<MessageBubble[]>([
    {
      id: 'welcome',
      role: 'assistant',
      content: '你好呀，我是年糕。有什么想聊聊的？不管是工作上的困惑、人际关系，还是想梳理一下最近的想法，我都在这里陪你。',
    },
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const flatListRef = useRef<FlatList>(null);

  // 构建历史消息（最近 10 轮）
  const buildHistory = useCallback((): ChatMessage[] => {
    const recent = messages.filter(m => m.role !== 'system').slice(-20);
    return recent.map(m => ({role: m.role as 'user' | 'assistant', content: m.content}));
  }, [messages]);

  const handleSend = async () => {
    const text = input.trim();
    if (!text || loading) return;
    setInput('');

    const userMsg: MessageBubble = {
      id: Date.now().toString(),
      role: 'user',
      content: text,
    };
    setMessages(prev => [...prev, userMsg]);
    setLoading(true);

    // 添加占位气泡
    const aiId = (Date.now() + 1).toString();
    setMessages(prev => [...prev, {id: aiId, role: 'assistant', content: ''}]);

    try {
      const history = buildHistory();
      const result = await sendMessage(text, TEMP_USER_ID, history);

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
      setMessages(prev =>
        prev.map(m =>
          m.id === aiId ? {...m, content: '抱歉，对话服务暂时不可用，请稍后再试。'} : m,
        ),
      );
    } finally {
      setLoading(false);
    }
  };

  const domainLabels: Record<string, string> = {
    career: '职场',
    relationship: '人际',
    cognition: '认知',
    life: '生活',
    emotion: '情感',
  };

  const renderMessage = ({item}: {item: MessageBubble}) => {
    if (item.role === 'system') return null;
    const isUser = item.role === 'user';

    return (
      <View style={[styles.bubbleRow, isUser && styles.bubbleRowUser]}>
        {!isUser && <View style={styles.aiAvatar}><Text style={styles.aiAvatarText}>糕</Text></View>}
        <View style={[styles.bubble, isUser ? styles.bubbleUser : styles.bubbleAi]}>
          <Text style={[styles.bubbleText, isUser && styles.bubbleTextUser]}>
            {item.content || '…'}
          </Text>
          {!isUser && loading && item.content === '' && (
            <ActivityIndicator size="small" color="#4a7c59" style={{marginTop: 4}} />
          )}
          {item.references && item.references.length > 0 && (
            <View style={styles.referenceBar}>
              <Text style={styles.referenceText}>
                📎 引用了 {item.references.length} 条经验
              </Text>
            </View>
          )}
        </View>
        {isUser && <View style={styles.userAvatar}><Text style={styles.userAvatarText}>我</Text></View>}
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>年糕对话</Text>
        <Text style={styles.headerSub}>AI 成长伙伴</Text>
      </View>

      {/* Messages */}
      <FlatList
        ref={flatListRef}
        data={messages}
        keyExtractor={item => item.id}
        renderItem={renderMessage}
        contentContainerStyle={styles.messageList}
        onContentSizeChange={() => flatListRef.current?.scrollToEnd({animated: true})}
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
            style={[styles.sendButton, (!input.trim() || loading) && styles.sendButtonDisabled]}
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
