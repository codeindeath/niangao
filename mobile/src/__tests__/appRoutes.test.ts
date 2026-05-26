import fs from 'fs';
import path from 'path';

describe('App route registry', () => {
  const mobileRoot = path.join(__dirname, '..', '..');

  it('does not register legacy placeholder routes in the V4 app shell', () => {
    const appSource = fs.readFileSync(path.join(mobileRoot, 'App.tsx'), 'utf8');

    expect(appSource).not.toContain('PlaceholderScreen');
    expect(appSource).not.toContain('name="placeholder"');
  });

  it('does not keep legacy unfinished placeholder screens', () => {
    const screensDir = path.join(mobileRoot, 'src', 'screens');

    expect(fs.existsSync(path.join(screensDir, 'PlaceholderScreen.tsx'))).toBe(false);
    expect(fs.existsSync(path.join(screensDir, 'FriendsScreen.tsx'))).toBe(false);
  });

  it('does not keep the legacy search screen beside the V4 search flow', () => {
    const screensDir = path.join(mobileRoot, 'src', 'screens');

    expect(fs.existsSync(path.join(screensDir, 'SearchScreen.tsx'))).toBe(false);
  });

  it('uses real icon components instead of raw glyph placeholders for bottom tabs', () => {
    const tabSource = fs.readFileSync(
      path.join(mobileRoot, 'src', 'navigation', 'BottomTabNavigator.tsx'),
      'utf8',
    );

    expect(tabSource).toContain('@expo/vector-icons/Ionicons');
    expect(tabSource).not.toContain("home: '◇'");
    expect(tabSource).not.toContain("chat: '◌'");
    expect(tabSource).not.toContain("profile: '◎'");
  });

  it('uses a real icon for the home search entry instead of an emoji glyph', () => {
    const homeSource = fs.readFileSync(
      path.join(mobileRoot, 'src', 'screens', 'HomeScreen.tsx'),
      'utf8',
    );

    expect(homeSource).toContain('@expo/vector-icons/Ionicons');
    expect(homeSource).toContain('search-outline');
    expect(homeSource).toContain('accessibilityLabel="搜索经验"');
    expect(homeSource).toContain("navigation.navigate('searchPage')");
    expect(homeSource).not.toContain('🔍');
  });

  it('uses real icons for experience card creator and action controls', () => {
    const cardSource = fs.readFileSync(
      path.join(mobileRoot, 'src', 'components', 'ExperienceCard.tsx'),
      'utf8',
    );

    expect(cardSource).toContain('@expo/vector-icons/Ionicons');
    expect(cardSource).toContain('person-circle-outline');
    expect(cardSource).toContain('sparkles');
    expect(cardSource).toContain('star');
    expect(cardSource).not.toContain('◎ {displayName}');
    expect(cardSource).not.toContain('✦ {item.inspiration_count');
    expect(cardSource).not.toContain('★ {item.is_collected');
    expect(cardSource).not.toContain('📖 原文');
  });

  it('uses real icons for chat topic and reference controls', () => {
    const chatSource = fs.readFileSync(
      path.join(mobileRoot, 'src', 'screens', 'ChatScreen.tsx'),
      'utf8',
    );

    expect(chatSource).toContain('@expo/vector-icons/Ionicons');
    expect(chatSource).toContain('chatbubble-ellipses-outline');
    expect(chatSource).toContain('bookmark');
    expect(chatSource).not.toContain('◎ {activeTopic');
    expect(chatSource).not.toContain("'☆'");
  });

  it('uses real icons for detail actions and private markers', () => {
    const detailSource = fs.readFileSync(
      path.join(mobileRoot, 'src', 'screens', 'DetailScreen.tsx'),
      'utf8',
    );

    expect(detailSource).toContain('@expo/vector-icons/Ionicons');
    expect(detailSource).toContain('sparkles');
    expect(detailSource).toContain('bookmark');
    expect(detailSource).toContain('lock-closed-outline');
    expect(detailSource).toContain('book-outline');
    expect(detailSource).not.toContain('✦ {exp.inspiration_count');
    expect(detailSource).not.toContain('★ {exp.is_collected');
    expect(detailSource).not.toContain('🔒');
    expect(detailSource).not.toContain('📖 经验解读');
  });
});
