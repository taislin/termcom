// JSONC parser — strips // single-line and /* */ block comments from JSONC text.
// Include this before any script that loads .jsonc files.
// Usage: const data = parseJSONC(rawText);
// This handles strings correctly (does not strip inside "...").

function parseJSONC(text) {
  // Strip // comments line-by-line (respects strings)
  var result = "";
  var inString = false;
  var escape = false;
  for (var i = 0; i < text.length; i++) {
    var ch = text[i];
    if (escape) { result += ch; escape = false; continue; }
    if (ch === "\\" && inString) { result += ch; escape = true; continue; }
    if (ch === '"') { inString = !inString; result += ch; continue; }
    if (!inString && ch === "/" && i + 1 < text.length && text[i + 1] === "/") {
      while (i < text.length && text[i] !== "\n") i++;
      result += "\n";
      continue;
    }
    if (!inString && ch === "/" && i + 1 < text.length && text[i + 1] === "*") {
      while (i + 1 < text.length && !(text[i] === "*" && text[i + 1] === "/")) i++;
      i += 2;
      continue;
    }
    result += ch;
  }
  return JSON.parse(result);
}

// Also export for Node.js (bundle script)
if (typeof module !== "undefined") module.exports = { parseJSONC };
