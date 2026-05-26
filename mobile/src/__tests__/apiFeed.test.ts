import {fetchExperience, fetchMyBookmarks, fetchMyExperiences, fetchRecommendations} from '../services/api';
import {apiGet} from '../services/config';

jest.mock('../services/config', () => ({
  apiGet: jest.fn(),
  apiPost: jest.fn(),
  apiPut: jest.fn(),
  apiPatch: jest.fn(),
  apiDelete: jest.fn(),
  getToken: jest.fn(),
}));

describe('feed API normalization', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('preserves V4 owner_user_id from mine feed cards', async () => {
    (apiGet as jest.Mock).mockResolvedValue({
      data: [{
        id: 'exp-1',
        owner_user_id: 'user-1',
        content: '先把复杂事写成一句能行动的话',
        experience_type: 'user_original',
        visibility: 'public',
        domain: 'meaning',
        sub_domain: 'self',
        creator_display_name: '阿树',
        inspiration_count: 3,
        collection_count: 2,
        is_inspired: false,
        is_collected: true,
      }],
      has_more: false,
    });

    const result = await fetchMyExperiences(1);

    expect(result.data[0].owner_user_id).toBe('user-1');
  });

  it('normalizes V4 detail response fields for App screens', async () => {
    (apiGet as jest.Mock).mockResolvedValue({
      id: 'exp-v4',
      owner_user_id: 'user-1',
      content: '先把复杂事写成一句能行动的话',
      domain: 'meaning',
      sub_domain: 'self',
      topic: '#自我',
      inspiration_count: 4,
      collection_count: 5,
      is_inspired: true,
      is_collected: false,
      experience_type: 'user_original',
      visibility: 'private',
      creator_display_name: '阿树',
      created_at: '2026-05-26T00:00:00Z',
    });

    const result = await fetchExperience('exp-v4');

    expect(result.owner_user_id).toBe('user-1');
    expect(result.visibility).toBe('private');
    expect(result.experience_type).toBe('user_original');
    expect(result.creator_display_name).toBe('阿树');
    expect(result.topic).toBe('#自我');
    expect(result.inspiration_count).toBe(4);
    expect(result.collection_count).toBe(5);
    expect(result.is_inspired).toBe(true);
    expect(result.is_collected).toBe(false);
  });

  it('normalizes V4 topic field for App screens', async () => {
    (apiGet as jest.Mock).mockResolvedValue({
      id: 'exp-topic',
      content: '把不确定的事先说清楚。',
      topic: '#边界沟通',
      domain: 'relationship',
      sub_domain: 'friendship',
      created_at: '2026-05-26T00:00:00Z',
    });

    const v4Result = await fetchExperience('exp-topic');
    expect(v4Result.topic).toBe('#边界沟通');
  });

  it('preserves unavailable collection placeholders without original content', async () => {
    (apiGet as jest.Mock).mockResolvedValue({
      data: [{
        id: 'exp-hidden',
        unavailable_reason: 'experience_unavailable',
        is_collected: true,
      }],
      has_more: false,
    });

    const result = await fetchMyBookmarks(1);

    expect(result.data[0]).toEqual(expect.objectContaining({
      id: 'exp-hidden',
      content: '',
      unavailable_reason: 'experience_unavailable',
      is_collected: true,
    }));
  });

  it('uses backend recommendation cursors instead of numeric offsets', async () => {
    (apiGet as jest.Mock).mockResolvedValue({
      data: [],
      next_cursor: 'rec:11111111-1111-4111-8111-111111111111:20',
      session_id: '11111111-1111-4111-8111-111111111111',
      has_more: true,
    });

    const result = await fetchRecommendations(20, 'rec:11111111-1111-4111-8111-111111111111:20');

    expect(apiGet).toHaveBeenCalledWith(
      '/api/v1/feed/recommend?limit=20&cursor=rec%3A11111111-1111-4111-8111-111111111111%3A20',
    );
    expect(result.next_cursor).toBe('rec:11111111-1111-4111-8111-111111111111:20');
  });

});
