import {fetchMyExperiences} from '../services/api';
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

});
