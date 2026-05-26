import React from 'react';
import {fireEvent, render} from '@testing-library/react-native';
import ExperienceCard from '../components/ExperienceCard';
import {Experience} from '../services/api';

jest.mock('@expo/vector-icons/Ionicons', () => {
  const React = require('react');
  const {Text} = require('react-native');
  return ({name}: {name: string}) => <Text>{name}</Text>;
});

describe('ExperienceCard unavailable placeholder', () => {
  const unavailableExperience: Experience = {
    id: 'exp-hidden',
    content: '这段原经验不应该在不可见占位卡上出现',
    unavailable_reason: 'experience_unavailable',
    domain: '',
    inspiration_count: 0,
    collection_count: 0,
    is_inspired: false,
    is_collected: true,
    interpretation: '这段解读也不应该出现',
    created_at: '2026-05-26T00:00:00Z',
  };

  it('renders a non-flippable unavailable card without original content', () => {
    const onFlipChange = jest.fn();
    const {getByLabelText, getByText, queryByText, queryByLabelText} = render(
      <ExperienceCard
        item={unavailableExperience}
        cardHeight={520}
        onFlipChange={onFlipChange}
        showActions
      />,
    );

    expect(getByLabelText('不可见经验卡片')).toBeTruthy();
    expect(getByText('该经验已不可见')).toBeTruthy();
    expect(getByText('它可能已经被删除、转为私密，或正在重新处理。')).toBeTruthy();
    expect(queryByText(unavailableExperience.content)).toBeNull();
    expect(queryByText(unavailableExperience.interpretation!)).toBeNull();
    expect(queryByLabelText('标记有启发')).toBeNull();
    expect(getByLabelText('从收藏移除')).toBeTruthy();

    fireEvent.press(getByLabelText('不可见经验卡片'));

    expect(onFlipChange).not.toHaveBeenCalled();
  });
});
