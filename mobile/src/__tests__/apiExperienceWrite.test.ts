import {createExperience, updateExperience} from '../services/api';
import {apiPost, apiPut} from '../services/config';

jest.mock('../services/config', () => ({
  apiGet: jest.fn(),
  apiPost: jest.fn(),
  apiPut: jest.fn(),
  apiPatch: jest.fn(),
  apiDelete: jest.fn(),
  getToken: jest.fn(),
}));

describe('experience write API contract', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (apiPost as jest.Mock).mockResolvedValue({
      experience: {
        id: 'exp-1',
        content: '把经验写短一点',
        domain: 'meaning',
        visibility: 'private',
        inspiration_count: 0,
        collection_count: 0,
        is_inspired: false,
        is_collected: false,
        created_at: '2026-05-26T00:00:00Z',
      },
    });
    (apiPut as jest.Mock).mockResolvedValue({status: 'ok'});
  });

  it('creates experiences with V4 visibility instead of legacy is_private payloads', async () => {
    await createExperience('把经验写短一点', 'meaning', 'self', true, '解读', '#表达');

    expect(apiPost).toHaveBeenCalledWith('/api/v1/experiences', expect.objectContaining({
      content: '把经验写短一点',
      visibility: 'private',
      topic: '#表达',
      source_scene: 'note',
    }));
    expect((apiPost as jest.Mock).mock.calls[0][1]).not.toHaveProperty('is_private');
    expect((apiPost as jest.Mock).mock.calls[0][1]).not.toHaveProperty('topics');
  });

  it('updates experiences with V4 visibility instead of legacy is_private payloads', async () => {
    await updateExperience('exp-1', '把经验写短一点', 'meaning', 'self', false, '解读', '#表达');

    expect(apiPut).toHaveBeenCalledWith('/api/v1/experiences/exp-1', expect.objectContaining({
      content: '把经验写短一点',
      visibility: 'public',
      topic: '#表达',
    }));
    expect((apiPut as jest.Mock).mock.calls[0][1]).not.toHaveProperty('is_private');
    expect((apiPut as jest.Mock).mock.calls[0][1]).not.toHaveProperty('topics');
  });
});
