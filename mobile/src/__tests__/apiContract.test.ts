import fs from 'fs';
import path from 'path';

describe('App V4 API contract', () => {
  const mobileRoot = path.join(__dirname, '..', '..');

  function read(relativePath: string): string {
    return fs.readFileSync(path.join(mobileRoot, relativePath), 'utf8');
  }

  it('uses V4 inspire and collect semantics for experience actions', () => {
    const apiSource = read('src/services/api.ts');
    const interactionSources = [
      apiSource,
      read('src/screens/HomeScreen.tsx'),
      read('src/screens/DetailScreen.tsx'),
      read('src/screens/SearchCardScreen.tsx'),
      read('src/screens/ChatScreen.tsx'),
    ].join('\n');

    expect(apiSource).toContain('/api/v1/experiences/${id}/inspire');
    expect(apiSource).toContain('/api/v1/experiences/${id}/collect');
    expect(interactionSources).not.toMatch(/\btoggleLike\b/);
    expect(interactionSources).not.toMatch(/\btoggleBookmark\b/);
  });

  it('does not call deprecated app-facing endpoints from mobile source', () => {
    const appSource = [
      read('App.tsx'),
      read('src/services/api.ts'),
      read('src/services/auth.ts'),
    ].join('\n');

    const deprecatedPaths = [
      '/api/v1/experiences/recommend',
      '/api/v1/experiences?',
      '/api/v1/me/bookmarks',
      '/api/v1/me/experiences',
      '/api/v1/user/profile',
      '/api/v1/user/stats',
      '/api/v1/chat/send',
      '/api/v1/experiences/${id}/view',
      '/api/v1/experiences/${id}/like',
      '/api/v1/experiences/${id}/bookmark',
    ];

    for (const endpoint of deprecatedPaths) {
      expect(appSource).not.toContain(endpoint);
    }
  });

  it('keeps legacy like/bookmark action aliases out of mobile UI runtime source', () => {
    const uiSources = [
      read('src/components/ExperienceCard.tsx'),
      read('src/screens/HomeScreen.tsx'),
      read('src/screens/DetailScreen.tsx'),
      read('src/screens/SearchCardScreen.tsx'),
      read('src/screens/SearchPage.tsx'),
    ].join('\n');

    expect(uiSources).not.toMatch(/\bis_liked\b/);
    expect(uiSources).not.toMatch(/\bis_bookmarked\b/);
    expect(uiSources).not.toMatch(/\blike_count\b/);
    expect(uiSources).not.toMatch(/\bbookmark_count\b/);
  });

  it('keeps legacy source classification aliases out of mobile UI runtime source', () => {
    const uiSources = [
      read('src/components/ExperienceCard.tsx'),
      read('src/screens/DetailScreen.tsx'),
      read('src/screens/SearchPage.tsx'),
    ].join('\n');

    expect(uiSources).not.toMatch(/\bsource_type\b/);
    expect(uiSources).not.toMatch(/\bis_official\b/);
  });

  it('keeps legacy privacy aliases out of mobile UI runtime source', () => {
    const uiSources = [
      read('src/screens/HomeScreen.tsx'),
      read('src/screens/DetailScreen.tsx'),
      read('src/screens/SearchCardScreen.tsx'),
      read('src/screens/CreateScreen.tsx'),
    ].join('\n');

    expect(uiSources).not.toMatch(/\bis_private\b/);
    expect(uiSources).not.toMatch(/\breview_status\b/);
  });

  it('keeps legacy creator and owner aliases out of mobile UI runtime source', () => {
    const uiSources = [
      read('src/components/ExperienceCard.tsx'),
      read('src/screens/DetailScreen.tsx'),
      read('src/screens/SearchPage.tsx'),
    ].join('\n');

    expect(uiSources).not.toMatch(/\bauthor_id\b/);
    expect(uiSources).not.toMatch(/\bauthor_name\b/);
    expect(uiSources).not.toMatch(/\bcreator_name\b/);
  });

  it('keeps legacy plural topics out of experience UI runtime source', () => {
    const uiSources = [
      read('src/screens/HomeScreen.tsx'),
      read('src/screens/DetailScreen.tsx'),
      read('src/screens/SearchCardScreen.tsx'),
      read('src/screens/CreateScreen.tsx'),
    ].join('\n');

    expect(uiSources).not.toMatch(/\btopics\b/);
  });

  it('keeps legacy experience aliases out of the exported App experience type', () => {
    const apiSource = read('src/services/api.ts');
    const start = apiSource.indexOf('export interface Experience {');
    const end = apiSource.indexOf('export interface ExperienceCard {');
    const experienceType = apiSource.slice(start, end);

    expect(experienceType).not.toMatch(/\bauthor_id\b/);
    expect(experienceType).not.toMatch(/\bauthor_name\b/);
    expect(experienceType).not.toMatch(/\bcreator_name\b/);
    expect(experienceType).not.toMatch(/\bis_private\b/);
    expect(experienceType).not.toMatch(/\breview_status\b/);
    expect(experienceType).not.toMatch(/\bis_official\b/);
    expect(experienceType).not.toMatch(/\bsource_type\b/);
  });

  it('keeps legacy detail/create aliases out of mobile API normalization', () => {
    const apiSource = read('src/services/api.ts');
    const start = apiSource.indexOf('function normalizeExperience(');
    const end = apiSource.indexOf('function normalizeFeedCard(');
    const normalizeSource = apiSource.slice(start, end);

    for (const legacy of [
      'author_id',
      'author_name',
      'creator_name',
      'topics',
      'is_private',
      'is_official',
      'source_type',
      'like_count',
      'bookmark_count',
      'is_liked',
      'is_bookmarked',
    ]) {
      expect(normalizeSource).not.toMatch(new RegExp(`\\b${legacy}\\b`));
    }
  });
});
