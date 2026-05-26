import {fetchExperience, fetchMyExperiences} from '../services/api';
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

  it('normalizes legacy action aliases from detail responses into V4 fields', async () => {
    (apiGet as jest.Mock).mockResolvedValue({
      id: 'exp-legacy',
      author_id: 'user-1',
      content: '先把复杂事写成一句能行动的话',
      domain: 'meaning',
      sub_domain: 'self',
      like_count: 4,
      bookmark_count: 5,
      is_liked: true,
      is_bookmarked: false,
      is_official: false,
      source_type: 'user',
      is_private: true,
      review_status: 'private',
      creator_name: '旧创建者',
      author_name: '旧作者',
      created_at: '2026-05-26T00:00:00Z',
    });

    const result = await fetchExperience('exp-legacy');

    expect(result.owner_user_id).toBe('user-1');
    expect(result.visibility).toBe('private');
    expect(result.experience_type).toBe('user_original');
    expect(result.creator_display_name).toBe('旧创建者');
    expect(result.inspiration_count).toBe(4);
    expect(result.collection_count).toBe(5);
    expect(result.is_inspired).toBe(true);
    expect(result.is_collected).toBe(false);
    expect(result).not.toHaveProperty('author_id');
    expect(result).not.toHaveProperty('author_name');
    expect(result).not.toHaveProperty('creator_name');
    expect(result).not.toHaveProperty('is_private');
    expect(result).not.toHaveProperty('is_official');
    expect(result).not.toHaveProperty('source_type');
    expect(result).not.toHaveProperty('review_status');
    expect(result).not.toHaveProperty('like_count');
    expect(result).not.toHaveProperty('bookmark_count');
    expect(result).not.toHaveProperty('is_liked');
    expect(result).not.toHaveProperty('is_bookmarked');
  });

  it('normalizes V4 and legacy topic fields into singular topic for App screens', async () => {
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

    (apiGet as jest.Mock).mockResolvedValue({
      id: 'exp-legacy-topic',
      content: '把不确定的事先说清楚。',
      topics: '#边界沟通',
      domain: 'relationship',
      sub_domain: 'friendship',
      created_at: '2026-05-26T00:00:00Z',
    });

    const legacyResult = await fetchExperience('exp-legacy-topic');
    expect(legacyResult.topic).toBe('#边界沟通');
  });

});
