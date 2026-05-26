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
});
