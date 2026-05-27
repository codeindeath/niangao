import {deleteAccount} from '../services/api';
import {apiDelete} from '../services/config';

jest.mock('../services/config', () => ({
  apiDelete: jest.fn(),
}));

describe('account API', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (apiDelete as jest.Mock).mockResolvedValue({message: '账号已注销'});
  });

  it('uses the V4 me account endpoint for account deletion', async () => {
    await expect(deleteAccount()).resolves.toEqual({message: '账号已注销'});

    expect(apiDelete).toHaveBeenCalledWith('/api/v1/me/account');
  });
});
