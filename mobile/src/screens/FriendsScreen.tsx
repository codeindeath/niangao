import React from 'react';
import {View, Text, StyleSheet} from 'react-native';

export default function FriendsScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.text}>晚点做~</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#faf8f5',
    justifyContent: 'center',
    alignItems: 'center',
  },
  text: {
    fontSize: 16,
    color: '#b5b0a8',
  },
});
