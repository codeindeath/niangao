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
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {createExperience, generateInterpretation} from '../services/api';

const DOMAINS: {key: string; label: string}[] = [
  {key: 'career', label: '职场成长'},
  {key: 'relationship', label: '人际关系'},
  {key: 'cognition', label: '认知升级'},
  {key: 'life', label: '生活智慧'},
  {key: 'emotion', label: '情感'},
];

export default function CreateScreen({navigation}: any) {
  const [content, setContent] = useState('');
  const [domain, setDomain] = useState('');
  const [interpretation, setInterpretation] = useState('');
  const [aiLoading, setAiLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const handleGenerateAI = async () => {
    if (!content.trim() || !domain) {
      Alert.alert('提示', '请先填写经验内容和选择领域');
      return;
    }
    setAiLoading(true);
    try {
      const result = await generateInterpretation(content.trim(), domain);
      setInterpretation(result.interpretation || '');
    } catch (e: any) {
      Alert.alert('生成失败', 'AI 解读生成失败，请稍后再试');
    } finally {
      setAiLoading(false);
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
    setSubmitting(true);
    try {
      await createExperience(
        content.trim(),
        domain,
        interpretation.trim() || undefined,
      );
      Alert.alert('发布成功', '你的经验已发布', [
        {text: '好的', onPress: () => navigation.goBack()},
      ]);
    } catch (e: any) {
      Alert.alert('发布失败', '请稍后再试');
    } finally {
      setSubmitting(false);
    }
  };

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
          style={[styles.publishBtn, (!content.trim() || !domain || submitting) && styles.publishBtnDisabled]}>
          {submitting ? (
            <ActivityIndicator size="small" color="#fff" />
          ) : (
            <Text style={[styles.publishBtnText, (!content.trim() || !domain) && styles.publishBtnTextDisabled]}>
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

          {/* Domain */}
          <Text style={styles.label}>领域</Text>
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

          {/* AI Interpretation */}
          <View style={styles.aiSection}>
            <TouchableOpacity
              style={[styles.aiButton, aiLoading && styles.aiButtonLoading]}
              onPress={handleGenerateAI}
              disabled={aiLoading}>
              {aiLoading ? (
                <ActivityIndicator size="small" color="#4a7c59" />
              ) : (
                <Text style={styles.aiButtonText}>🤖 AI 生成解读</Text>
              )}
            </TouchableOpacity>
          </View>

          {interpretation !== '' && (
            <>
              <Text style={styles.label}>AI 解读</Text>
              <View style={styles.interpretationBox}>
                <Text style={styles.interpretationText}>{interpretation}</Text>
              </View>
              <TouchableOpacity onPress={() => setInterpretation('')}>
                <Text style={styles.clearInterpretation}>清除重新生成</Text>
              </TouchableOpacity>
            </>
          )}
        </ScrollView>
      </KeyboardAvoidingView>

      {/* Bottom bar */}
      <View style={styles.bottomBar}>
        <Text style={styles.bottomHint}>
          {content.trim() && domain ? '内容就绪，可以发布' : '请填写经验和选择领域'}
        </Text>
        <TouchableOpacity
          style={[styles.submitButton, (!content.trim() || !domain || submitting) && styles.submitButtonDisabled]}
          onPress={handlePublish}
          disabled={!content.trim() || !domain || submitting}>
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
  aiSection: {
    marginTop: 24,
    alignItems: 'center',
  },
  aiButton: {
    backgroundColor: '#eaf2e8',
    borderRadius: 20,
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderWidth: 1,
    borderColor: '#4a7c59',
  },
  aiButtonLoading: {
    opacity: 0.6,
  },
  aiButtonText: {
    fontSize: 15,
    fontWeight: '600',
    color: '#4a7c59',
  },
  interpretationBox: {
    backgroundColor: '#ffffff',
    borderRadius: 14,
    padding: 14,
    borderWidth: 0.5,
    borderColor: '#d4e0d6',
  },
  interpretationText: {
    fontSize: 14,
    lineHeight: 22,
    color: '#3d3d3d',
  },
  clearInterpretation: {
    fontSize: 12,
    color: '#9a9a9a',
    textAlign: 'right',
    marginTop: 6,
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
