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
});
