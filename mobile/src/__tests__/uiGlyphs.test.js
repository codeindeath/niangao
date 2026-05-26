const fs = require('fs');
const path = require('path');

const SOURCE_DIRS = [
  path.resolve(__dirname, '../screens'),
  path.resolve(__dirname, '../components'),
];

function listTsxFiles(dir, out = []) {
  for (const entry of fs.readdirSync(dir, {withFileTypes: true})) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      listTsxFiles(fullPath, out);
    } else if (entry.name.endsWith('.tsx')) {
      out.push(fullPath);
    }
  }
  return out;
}

describe('UI glyph polish', () => {
  it('does not render crude text glyphs as icons in app UI', () => {
    const forbiddenTextGlyph = /<Text\b[^>]*>\s*[←→↩↪›‹×＋✕□◇◆○●]/g;
    const offenders = SOURCE_DIRS
      .flatMap(dir => listTsxFiles(dir))
      .flatMap(file => {
        const source = fs.readFileSync(file, 'utf8');
        return [...source.matchAll(forbiddenTextGlyph)].map(match => {
          const line = source.slice(0, match.index).split('\n').length;
          return `${path.relative(path.resolve(__dirname, '..'), file)}:${line}:${match[0].trim()}`;
        });
      });

    expect(offenders).toEqual([]);
  });
});
