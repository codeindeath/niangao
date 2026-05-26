import React, {useState, useEffect} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ActivityIndicator,
} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';
import {fetchProfile, updateProfile} from '../services/api';
import Ionicons from '@expo/vector-icons/Ionicons';

export default function ProfileEditScreen({navigation}: any) {
  const [displayName, setDisplayName] = useState('');
  const [freeDescription, setFreeDescription] = useState('');
  const [commonIssues, setCommonIssues] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchProfile().then(p => {
      setDisplayName(p.display_name || p.nickname || '');
      setFreeDescription(p.free_description || p.bio || '');
      setCommonIssues((p.common_issues || []).join('、'));
    }).catch(console.error).finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    const name = displayName.trim();
    const description = freeDescription.trim();
    const issues = commonIssues
      .split(/[、,，\n]/)
      .map(item => item.trim())
      .filter(Boolean)
      .slice(0, 6);
    if (!name) { Alert.alert('提示', '名字不能为空'); return; }
    if ([...name].length > 30) { Alert.alert('提示', '名字不能超过30字'); return; }
    if ([...description].length > 200) { Alert.alert('提示', '一句介绍不能超过200字'); return; }
    if (issues.some(item => [...item].length > 20)) { Alert.alert('提示', '常聊的事每项不超过20字'); return; }

    setSaving(true);
    try {
      await updateProfile({
        display_name: name,
        free_description: description,
        common_issues: issues,
      });
      navigation.goBack();
    } catch (e: any) {
      Alert.alert('保存失败', e?.message || '请稍后再试');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <SafeAreaView style={styles.container} edges={['top']}>
        <ActivityIndicator size="large" color="#4a7c59" style={{marginTop: 200}} />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()} style={styles.backBtn} accessibilityRole="button" accessibilityLabel="返回">
          <Ionicons name="chevron-back" size={22} color="#5c5548" />
        </TouchableOpacity>
        <Text style={styles.headerTitle}>编辑个人信息</Text>
        <TouchableOpacity onPress={handleSave} style={styles.saveBtn} disabled={saving}>
          <Text style={[styles.saveBtnText, saving && {opacity: 0.5}]}>保存</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.content}>
        {/* Avatar */}
        <View style={styles.avatarRow}>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{(displayName || '年').charAt(0)}</Text>
          </View>
        </View>

        {/* Display name */}
        <View style={styles.field}>
          <Text style={styles.fieldLabel}>名字</Text>
          <TextInput
            style={styles.fieldInput}
            value={displayName}
            onChangeText={setDisplayName}
            placeholder="输入名字"
            placeholderTextColor="#b5b0a8"
            maxLength={30}
          />
          <Text style={styles.fieldHint}>{[...displayName].length}/30</Text>
        </View>

        {/* Free description */}
        <View style={styles.field}>
          <Text style={styles.fieldLabel}>一句介绍</Text>
          <TextInput
            style={[styles.fieldInput, styles.bioInput]}
            value={freeDescription}
            onChangeText={setFreeDescription}
            placeholder="比如：正在认真生活的人"
            placeholderTextColor="#b5b0a8"
            multiline
            numberOfLines={3}
            textAlignVertical="top"
            maxLength={200}
          />
          <Text style={styles.fieldHint}>{[...freeDescription].length}/200</Text>
        </View>

        {/* Common issues */}
        <View style={styles.field}>
          <Text style={styles.fieldLabel}>常聊的事</Text>
          <TextInput
            style={styles.fieldInput}
            value={commonIssues}
            onChangeText={setCommonIssues}
            placeholder="用顿号分隔，比如：工作、关系、情绪"
            placeholderTextColor="#b5b0a8"
          />
          <Text style={styles.fieldHint}>最多6项</Text>
        </View>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {flex: 1, backgroundColor: '#faf8f5'},
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#e8e4dc',
    backgroundColor: '#faf8f5',
  },
  backBtn: {width: 36, height: 36, justifyContent: 'center', alignItems: 'center', borderRadius: 8},
  headerTitle: {flex: 1, textAlign: 'center', fontSize: 16, fontWeight: '600', color: '#2a2722'},
  saveBtn: {paddingHorizontal: 16, paddingVertical: 8, borderRadius: 8, backgroundColor: '#4a7c59'},
  saveBtnText: {color: '#fff', fontSize: 14, fontWeight: '600'},
  content: {paddingTop: 24, paddingHorizontal: 24},
  avatarRow: {alignItems: 'center', marginBottom: 28},
  avatar: {
    width: 80, height: 80, borderRadius: 40,
    backgroundColor: '#4a7c59',
    justifyContent: 'center', alignItems: 'center',
  },
  avatarText: {fontSize: 32, fontWeight: '700', color: '#fff'},
  field: {marginBottom: 20},
  fieldLabel: {fontSize: 14, fontWeight: '600', color: '#4a4a4a', marginBottom: 8},
  fieldInput: {
    backgroundColor: '#fff',
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 12,
    fontSize: 15,
    color: '#2a2722',
    borderWidth: 0.5,
    borderColor: '#e8e4dc',
  },
  bioInput: {minHeight: 80},
  fieldHint: {fontSize: 11, color: '#b5b0a8', marginTop: 4, textAlign: 'right'},
});
