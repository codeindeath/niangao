import React from 'react';
import { render, waitFor, fireEvent } from '@testing-library/react-native';
import DetailScreen from '../screens/DetailScreen';
import * as api from '../services/api';

jest.mock('../services/api');

describe('DetailScreen', () => {
  const mockExp = {
    id: '1',
    content: '测试经验',
    domain: 'life',
    sub_domain: 'time-mgmt',
    author_name: 'Tester',
    like_count: 5,
    bookmark_count: 3,
    is_liked: false,
    is_bookmarked: false,
    is_private: false,
    review_status: 'approved',
    quality_score: 7.5,
    score_details: '{"value":8,"actionable":7,"universal":7,"original":6,"clarity":9}',
    created_at: '2026-01-01T00:00:00Z',
  };

  beforeEach(() => { jest.clearAllMocks(); });

  it('shows quality score when available', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    const { findByText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    expect(await findByText('7.5')).toBeTruthy();
    expect(await findByText('/ 10')).toBeTruthy();
  });

  it('shows review status badge', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    const { findByText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    expect(await findByText('✓ 已通过')).toBeTruthy();
  });

  it('shows private marker for private experiences', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue({...mockExp, is_private: true, review_status: 'private'});
    const { findByText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    expect(await findByText('🔒 私密')).toBeTruthy();
  });

  it('shows sub-domain label when available', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    const { findByText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    expect(await findByText('时间管理')).toBeTruthy();
  });

  it('toggles like on press', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    (api.toggleLike as jest.Mock).mockResolvedValue({liked: true});
    const { findByText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    const likeBtn = await findByText('♥ 5');
    fireEvent.press(likeBtn);
    expect(api.toggleLike).toHaveBeenCalledWith('1');
  });
});
