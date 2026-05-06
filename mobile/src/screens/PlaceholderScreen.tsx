import React from 'react';
import {View, Text, TouchableOpacity, StyleSheet} from 'react-native';
import {SafeAreaView} from 'react-native-safe-area-context';

export default function PlaceholderScreen({route, navigation}: any) {
  const {title} = route.params as {title: string};

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()} style={styles.backBtn}>
          <Text style={styles.backBtnText}>←</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>{title}</Text>
        <View style={{width: 36}} />
      </View>
      <View style={styles.body}>
        <Text style={styles.text}>晚点做~</Text>
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
  body: {flex: 1, justifyContent: 'center', alignItems: 'center'},
  text: {fontSize: 16, color: '#b5b0a8'},
});
