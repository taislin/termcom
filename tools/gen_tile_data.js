const fs = require('fs');
const path = require('path');

function parseJSONC(text) {
  var result = '';
  var inString = false;
  var escape = false;
  for (var i = 0; i < text.length; i++) {
    var ch = text[i];
    if (escape) { result += ch; escape = false; continue; }
    if (ch === '\\' && inString) { result += ch; escape = true; continue; }
    if (ch === '"') { inString = !inString; result += ch; continue; }
    if (!inString && ch === '/' && i + 1 < text.length && text[i + 1] === '/') {
      while (i < text.length && text[i] !== '\n') i++;
      result += '\n';
      continue;
    }
    if (!inString && ch === '/' && i + 1 < text.length && text[i + 1] === '*') {
      while (i + 1 < text.length && !(text[i] === '*' && text[i + 1] === '/')) i++;
      i += 2;
      continue;
    }
    result += ch;
  }
  return JSON.parse(result);
}

const dir = path.join(__dirname, '..', 'data', 'tiles');
const files = fs.readdirSync(dir).filter(f => f.endsWith('.jsonc'));
let allTiles = [];
for (const file of files) {
  const text = fs.readFileSync(path.join(dir, file), 'utf8');
  const tiles = parseJSONC(text);
  allTiles = allTiles.concat(tiles);
}

const js = '// Auto-generated from data/tiles/*.jsonc\nconst TILE_DATA = ' + JSON.stringify(allTiles, null, 2) + ';\n';
fs.writeFileSync(path.join(__dirname, 'tile_data.js'), js);
console.log('Generated tile_data.js with ' + allTiles.length + ' tiles');
