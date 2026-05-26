import React from 'react';
import {Alert} from 'react-native';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import DetailScreen from '../screens/DetailScreen';
import * as api from '../services/api';
import * as config from '../services/config';

jest.mock('../services/api');
jest.mock('../services/config');

describe('DetailScreen', () => {
  const mockExp = {
    id: '1',
    author_id: 'author-1',
    content: '测试经验',
    domain: 'meaning',
    sub_domain: 'self',
    author_name: 'Tester',
    inspiration_count: 5,
    collection_count: 3,
    is_inspired: false,
    is_collected: false,
    is_official: false,
    is_private: false,
    review_status: 'approved',
    quality_score: 7.5,
    score_details: '{"value":8,"actionable":7,"universal":7,"original":6,"clarity":9}',
    created_at: '2026-01-01T00:00:00Z',
  };

  beforeEach(() => {
    jest.clearAllMocks();
    jest.spyOn(Alert, 'alert').mockImplementation(() => {});
    (config.getUserInfo as jest.Mock).mockResolvedValue(null);
    (config.getToken as jest.Mock).mockResolvedValue('fake-token');
  });

  afterEach(() => {
    (Alert.alert as jest.Mock).mockRestore();
  });

  it('shows quality score as stars when available', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    const { findByText, getAllByText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    expect(await findByText('价值度')).toBeTruthy();
    expect(getAllByText('★')).toHaveLength(4);
    expect(getAllByText('☆')).toHaveLength(1);
  });

  it('shows private marker for private experiences', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue({...mockExp, is_private: true, review_status: 'private'});
    const { findByLabelText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    expect(await findByLabelText('私密经验')).toBeTruthy();
  });

  it('shows sub-domain label when available', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    const { findByText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    expect(await findByText('自我')).toBeTruthy();
  });

  it('marks an experience as inspiring once', async () => {
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    (api.markInspired as jest.Mock).mockResolvedValue({inspired: true});
    const { findByLabelText } = render(<DetailScreen route={{params: {id: '1'}}} navigation={{}} />);
    const likeBtn = await findByLabelText('标记有启发');
    fireEvent.press(likeBtn);
    await waitFor(() => {
      expect(api.markInspired).toHaveBeenCalledWith('1');
    });
  });

  it('offers turn-private as the safer path when deleting a public own experience', async () => {
    (config.getUserInfo as jest.Mock).mockResolvedValue({id: 'author-1'});
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    (api.updateExperience as jest.Mock).mockResolvedValue({status: 'ok'});

    const navigation = {goBack: jest.fn(), navigate: jest.fn()};
    const {findByText} = render(<DetailScreen route={{params: {id: '1'}}} navigation={navigation} />);

    fireEvent.press(await findByText('删除'));
    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    const makePrivateButton = buttons.find((button: any) => button.text === '转为私密');
    const deleteButton = buttons.find((button: any) => button.text === '删除');
    expect(buttons.indexOf(makePrivateButton)).toBeLessThan(buttons.indexOf(deleteButton));
    makePrivateButton.onPress();

    await waitFor(() => {
      expect(api.updateExperience).toHaveBeenCalledWith(
        '1',
        '测试经验',
        'meaning',
        'self',
        true,
        undefined,
        undefined,
      );
    });
    expect(api.deleteExperience).not.toHaveBeenCalled();
    expect(navigation.goBack).not.toHaveBeenCalled();
  });

  it('still allows explicit deletion from the public-own delete confirmation', async () => {
    (config.getUserInfo as jest.Mock).mockResolvedValue({id: 'author-1'});
    (api.fetchExperience as jest.Mock).mockResolvedValue(mockExp);
    (api.deleteExperience as jest.Mock).mockResolvedValue({status: 'ok'});

    const navigation = {goBack: jest.fn(), navigate: jest.fn()};
    const {findByText} = render(<DetailScreen route={{params: {id: '1'}}} navigation={navigation} />);

    fireEvent.press(await findByText('删除'));
    const buttons = (Alert.alert as jest.Mock).mock.calls[0][2];
    const deleteButton = buttons.find((button: any) => button.text === '删除');
    deleteButton.onPress();

    await waitFor(() => {
      expect(api.deleteExperience).toHaveBeenCalledWith('1');
      expect(navigation.goBack).toHaveBeenCalledTimes(1);
    });
  });

  it('shows owner actions when V4 owner_user_id matches the current user', async () => {
    (config.getUserInfo as jest.Mock).mockResolvedValue({id: 'author-1'});
    (api.fetchExperience as jest.Mock).mockResolvedValue({
      ...mockExp,
      author_id: 'legacy-author',
      owner_user_id: 'author-1',
    });

    const {findByText} = render(<DetailScreen route={{params: {id: '1'}}} navigation={{navigate: jest.fn()}} />);

    expect(await findByText('编辑')).toBeTruthy();
    expect(await findByText('转为私密')).toBeTruthy();
    expect(await findByText('删除')).toBeTruthy();
  });
});
