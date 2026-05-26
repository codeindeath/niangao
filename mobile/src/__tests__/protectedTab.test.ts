import {openProtectedMainTab} from '../utils/protectedTab';
import * as config from '../services/config';

jest.mock('../services/config');

describe('openProtectedMainTab', () => {
  const makeNavigation = () => {
    const parent = {navigate: jest.fn()};
    return {
      navigate: jest.fn(),
      getParent: jest.fn(() => parent),
      parent,
    };
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('opens login before authenticated tabs when the user is a guest', async () => {
    (config.getToken as jest.Mock).mockResolvedValue(null);
    const event = {preventDefault: jest.fn()};
    const navigation = makeNavigation();

    await openProtectedMainTab(event, navigation, 'chat');

    expect(event.preventDefault).toHaveBeenCalledTimes(1);
    expect(navigation.navigate).not.toHaveBeenCalled();
    expect(navigation.parent.navigate).toHaveBeenCalledWith('login');
  });

  it('opens the requested tab when a token exists', async () => {
    (config.getToken as jest.Mock).mockResolvedValue('token-1');
    const event = {preventDefault: jest.fn()};
    const navigation = makeNavigation();

    await openProtectedMainTab(event, navigation, 'create');

    expect(event.preventDefault).toHaveBeenCalledTimes(1);
    expect(navigation.navigate).toHaveBeenCalledWith('create');
    expect(navigation.parent.navigate).not.toHaveBeenCalled();
  });

  it('falls back to login when token lookup fails', async () => {
    (config.getToken as jest.Mock).mockRejectedValue(new Error('storage failed'));
    const event = {preventDefault: jest.fn()};
    const navigation = makeNavigation();

    await openProtectedMainTab(event, navigation, 'chat');

    expect(event.preventDefault).toHaveBeenCalledTimes(1);
    expect(navigation.navigate).not.toHaveBeenCalled();
    expect(navigation.parent.navigate).toHaveBeenCalledWith('login');
  });
});
