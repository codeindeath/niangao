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
import {fetchProfile, updateProfile, UserProfile} from '../services/api';

export default function ProfileEditScreen({navigation}: any) {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [nickname, setNickname] = useState('');
  const [title, setTitle] = useState('');
  const [bio, setBio] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchProfile().then(p => {
      setProfile(p);
      setNickname(p.nickname || '');
      setTitle(p.title || '');
      setBio(p.bio || '');
    }).catch(console.error).finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    const n = nickname.trim();
    if (!n) { Alert.alert('提示', '昵称不能为空'); return; }
    if ([...n].length > 30) { Alert.alert('提示', '昵称不能超过30字'); return; }
    if ([...title].length > 20) { Alert.alert('提示', '称号不能超过20字'); return; }
    if ([...bio].length > 100) { Alert.alert('提示', '简介不能超过100字'); return; }

    setSaving(true);
    try {
      await updateProfile({nickname: n, title: title.trim(), bio: bio.trim()});
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
        <TouchableOpacity onPress={() => navigation.goBack()} style={styles.backBtn}>
          <Text style={styles.backBtnText}>←</Text>
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
            <Text style={styles.avatarText}>{(nickname || '?').charAt(0)}</Text>
          </View>
        </View>

        {/* Nickname */}
        <View style={styles.field}>
          <Text style={styles.fieldLabel}>昵称</Text>
          <TextInput
            style={styles.fieldInput}
            value={nickname}
            onChangeText={setNickname}
            placeholder="输入昵称"
            placeholderTextColor="#b5b0a8"
            maxLength={30}
          />
          <Text style={styles.fieldHint}>{[...nickname].length}/30</Text>
        </View>

        {/* Title */}
        <View style={styles.field}>
          <Text style={styles.fieldLabel}>称号</Text>
          <TextInput
            style={styles.fieldInput}
            value={title}
            onChangeText={setTitle}
            placeholder="给自己一个独特称号"
            placeholderTextColor="#b5b0a8"
            maxLength={20}
          />
          <Text style={styles.fieldHint}>{[...title].length}/20</Text>
        </View>

        {/* Bio */}
        <View style={styles.field}>
          <Text style={styles.fieldLabel}>简介</Text>
          <TextInput
            style={[styles.fieldInput, styles.bioInput]}
            value={bio}
            onChangeText={setBio}
            placeholder="简单介绍一下自己"
            placeholderTextColor="#b5b0a8"
            multiline
            numberOfLines={3}
            textAlignVertical="top"
            maxLength={100}
          />
          <Text style={styles.fieldHint}>{[...bio].length}/100</Text>
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
  backBtnText: {fontSize: 20, color: '#5c5548', fontWeight: '600'},
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
