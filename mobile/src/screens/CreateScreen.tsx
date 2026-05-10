import React, {useState} from 'react';
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
  Switch,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {createExperience, ApiError} from '../services/api';
import {triggerTabRefresh} from './HomeScreen';

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
    {key: 'faith', label: '信仰'},
    {key: 'mission', label: '使命'},
    {key: 'belonging', label: '归属'},
  ],
};

export default function CreateScreen({navigation}: any) {
  const [content, setContent] = useState('');
  const [domain, setDomain] = useState('');
  const [subDomain, setSubDomain] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const handleDomainSelect = (key: string) => {
    if (domain === key) {
      // Deselecting the first-level domain clears everything
      setDomain('');
      setSubDomain('');
    } else {
      setDomain(key);
      setSubDomain(''); // Reset sub-domain when switching first-level domains
    }
  };

  const handlePublish = async () => {
    if (!content.trim()) {
      Alert.alert('提示', '请输入经验内容');
      return;
    }
    if (!domain) {
      Alert.alert('提示', '请选择领域');
      return;
    }
    if (!subDomain) {
      Alert.alert('提示', '请选择子领域');
      return;
    }
    setSubmitting(true);
    try {
      await createExperience(
        content.trim(),
        domain,
        subDomain,
        isPrivate,
        undefined,
      );
      Alert.alert('发布成功', '你的经验已发布', [
        {text: '好的', onPress: () => {
          triggerTabRefresh('my');
          navigation.goBack();
        }},
      ]);
    } catch (e: any) {
      if (e instanceof ApiError && e.status === 401) {
        Alert.alert('未登录', '请先登录后再发布经验');
      } else {
        Alert.alert('发布失败', e?.message || String(e));
      }
    } finally {
      setSubmitting(false);
    }
  };

  const getBottomHint = () => {
    if (!content.trim()) return '请输入经验内容';
    if (!domain) return '请选择经验领域';
    if (!subDomain) return '请选择子领域';
    return '内容就绪，可以发布';
  };
  const isPublishReady = !!(content.trim() && domain && subDomain);

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.cancelText}>取消</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>发布经验</Text>
        <TouchableOpacity
          onPress={handlePublish}
          disabled={submitting}
          style={[styles.publishBtn, (!isPublishReady || submitting) && styles.publishBtnDisabled]}>
          {submitting ? (
            <ActivityIndicator size="small" color="#fff" />
          ) : (
            <Text style={[styles.publishBtnText, !isPublishReady && styles.publishBtnTextDisabled]}>
              发布
            </Text>
          )}
        </TouchableOpacity>
      </View>

      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={{flex: 1}}>
        <ScrollView
          style={styles.body}
          contentContainerStyle={{paddingBottom: 40}}
          keyboardShouldPersistTaps="handled">
          {/* Content */}
          <Text style={styles.label}>经验内容</Text>
          <TextInput
            style={styles.contentInput}
            value={content}
            onChangeText={setContent}
            placeholder="写下你的经验，不超过 100 字..."
            placeholderTextColor="#b5b0a8"
            multiline
            maxLength={100}
            textAlignVertical="top"
          />
          <Text style={styles.charCount}>{content.length}/100</Text>

          {/* Domain - First Level */}
          <Text style={styles.label}>领域</Text>
          <View style={styles.domainRow}>
            {PRIMARY_DOMAINS.map(d => (
              <TouchableOpacity
                key={d.key}
                style={[styles.domainChip, domain === d.key && styles.domainChipActive]}
                onPress={() => handleDomainSelect(d.key)}>
                <Text style={[styles.domainChipText, domain === d.key && styles.domainChipTextActive]}>
                  {d.label}
                </Text>
              </TouchableOpacity>
            ))}
          </View>

          {/* Domain - Second Level (sub-domains) */}
          {domain !== '' && SUB_DOMAINS[domain] && (
            <>
              <Text style={styles.subLabel}>子领域</Text>
              <View style={styles.domainRow}>
                {SUB_DOMAINS[domain].map(sd => (
                  <TouchableOpacity
                    key={sd.key}
                    style={[styles.subDomainChip, subDomain === sd.key && styles.subDomainChipActive]}
                    onPress={() => setSubDomain(sd.key)}>
                    <Text style={[styles.subDomainChipText, subDomain === sd.key && styles.subDomainChipTextActive]}>
                      {sd.label}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </>
          )}

          {/* Private / Public Toggle */}
          <View style={styles.privacyRow}>
            <Text style={styles.privacyLabel}>私密经验（仅自己可见）</Text>
            <Switch
              value={isPrivate}
              onValueChange={setIsPrivate}
              trackColor={{false: '#e0dcd5', true: '#4a7c59'}}
              thumbColor={isPrivate ? '#ffffff' : '#f4f3f0'}
            />
          </View>

        </ScrollView>
      </KeyboardAvoidingView>

      {/* Bottom bar */}
      <View style={styles.bottomBar}>
        <Text style={styles.bottomHint}>
          {getBottomHint()}
        </Text>
        <TouchableOpacity
          style={[styles.submitButton, (!isPublishReady || submitting) && styles.submitButtonDisabled]}
          onPress={handlePublish}
          disabled={!isPublishReady || submitting}>
          {submitting ? (
            <ActivityIndicator size="small" color="#fff" />
          ) : (
            <Text style={styles.submitButtonText}>发布经验</Text>
          )}
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#faf8f5',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 0.5,
    borderBottomColor: '#e8e4df',
  },
  cancelText: {
    fontSize: 15,
    color: '#9a9a9a',
  },
  headerTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#1a1a1a',
  },
  publishBtn: {
    paddingHorizontal: 14,
    paddingVertical: 6,
    borderRadius: 14,
    backgroundColor: '#4a7c59',
  },
  publishBtnDisabled: {
    backgroundColor: '#c5d4c9',
  },
  publishBtnText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#ffffff',
  },
  publishBtnTextDisabled: {
    color: '#e8f0ea',
  },
  body: {
    flex: 1,
    paddingHorizontal: 18,
    paddingTop: 20,
  },
  label: {
    fontSize: 13,
    fontWeight: '600',
    color: '#6e6e6e',
    marginBottom: 8,
    marginTop: 20,
  },
  subLabel: {
    fontSize: 12,
    fontWeight: '500',
    color: '#8a8a8a',
    marginBottom: 8,
    marginTop: 14,
  },
  contentInput: {
    backgroundColor: '#ffffff',
    borderRadius: 14,
    padding: 14,
    fontSize: 16,
    lineHeight: 24,
    color: '#1a1a1a',
    minHeight: 120,
    borderWidth: 0.5,
    borderColor: '#f0ece7',
  },
  charCount: {
    fontSize: 11,
    color: '#b5b0a8',
    textAlign: 'right',
    marginTop: 4,
    marginRight: 4,
  },
  domainRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  domainChip: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 18,
    backgroundColor: '#ffffff',
    borderWidth: 0.5,
    borderColor: '#e8e4df',
  },
  domainChipActive: {
    backgroundColor: '#4a7c59',
    borderColor: '#4a7c59',
  },
  domainChipText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#6e6e6e',
  },
  domainChipTextActive: {
    color: '#ffffff',
  },
  subDomainChip: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 18,
    backgroundColor: '#f5f2ed',
    borderWidth: 0.5,
    borderColor: '#e0dcd5',
  },
  subDomainChipActive: {
    backgroundColor: '#2d5a3d',
    borderColor: '#2d5a3d',
  },
  subDomainChipText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#8a8a8a',
  },
  subDomainChipTextActive: {
    color: '#ffffff',
  },
  privacyRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 24,
    paddingVertical: 10,
    paddingHorizontal: 4,
  },
  privacyLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#4a4a4a',
  },
  bottomBar: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 18,
    paddingVertical: 12,
    backgroundColor: '#ffffff',
    borderTopWidth: 0.5,
    borderTopColor: '#e8e4df',
  },
  bottomHint: {
    fontSize: 12,
    color: '#9a9a9a',
    flex: 1,
  },
  submitButton: {
    backgroundColor: '#4a7c59',
    borderRadius: 22,
    paddingHorizontal: 24,
    paddingVertical: 12,
  },
  submitButtonDisabled: {
    backgroundColor: '#c5d4c9',
  },
  submitButtonText: {
    color: '#ffffff',
    fontSize: 15,
    fontWeight: '600',
  },
});
