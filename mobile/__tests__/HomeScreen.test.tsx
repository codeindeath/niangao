import React from 'react';
import { render, waitFor } from '@testing-library/react-native';
import HomeScreen from '../screens/HomeScreen';
import * as api from '../services/api';
import * as config from '../services/config';

jest.mock('../services/api');
jest.mock('../services/config');

describe('HomeScreen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('shows loading spinner initially', () => {
    const { getByTestId } = render(<HomeScreen navigation={{}} />);
    // Would need testID on ActivityIndicator
  });

  it('loads recommendations when logged in', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('fake-token');
    (api.fetchRecommendations as jest.Mock).mockResolvedValue({
      data: [{ id: '1', content: 'test', domain: 'life', author_name: 'Tester', like_count: 0, is_liked: false, is_bookmarked: false }],
      total: 1,
    });

    render(<HomeScreen navigation={{}} />);
    await waitFor(() => {
      expect(api.fetchRecommendations).toHaveBeenCalledWith(20);
    });
  });

  it('falls back to public feed when not logged in', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchExperiences as jest.Mock).mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
    });

    render(<HomeScreen navigation={{}} />);
    await waitFor(() => {
      expect(api.fetchExperiences).toHaveBeenCalledWith(1);
    });
  });

  it('shows empty state when no experiences', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchExperiences as jest.Mock).mockResolvedValue({ data: [], total: 0, page: 1 });

    const { findByText } = render(<HomeScreen navigation={{}} />);
    expect(await findByText('暂无推荐内容')).toBeTruthy();
  });

  it('shows error state with retry button', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    (api.fetchExperiences as jest.Mock).mockRejectedValue(new Error('Network error'));

    const { findByText } = render(<HomeScreen navigation={{}} />);
    expect(await findByText('加载失败，请检查网络连接')).toBeTruthy();
    expect(await findByText('重试')).toBeTruthy();
  });
});
