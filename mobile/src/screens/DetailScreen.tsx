import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

export default function DetailScreen() {
  return (
    <SafeAreaView style={styles.container}>
      <Text>经验详情 — 开发中</Text>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#faf8f5', justifyContent: 'center', alignItems: 'center' },
});
