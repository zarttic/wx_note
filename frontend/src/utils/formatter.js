/**
 * One-click Markdown formatting: CJK-Latin spacing, punctuation fixes, paragraph normalization.
 */

// CJK ranges: Hiragana, Katakana, CJK Unified Ideographs, CJK Compatibility, Hangul
const CJK = '\\u2e80-\\u9fff\\uf900-\\ufaff\\ufe30-\\ufe4f'
const CJK_RE = new RegExp(`[${CJK}]`)

/**
 * Add space between CJK and Latin/number characters.
 * "你好world" → "你好 world", "123个" → "123 个"
 */
function cjkLatinSpacing(text) {
  // Space between CJK char and following Latin/digit
  text = text.replace(
    new RegExp(`([${CJK}])([a-zA-Z0-9])`, 'g'),
    '$1 $2'
  )
  // Space between Latin/digit and following CJK char
  text = text.replace(
    new RegExp(`([a-zA-Z0-9])([${CJK}])`, 'g'),
    '$1 $2'
  )
  return text
}

/**
 * Fix common Chinese punctuation issues:
 * - Replace half-width comma/period between CJK chars with full-width
 * - Remove spaces around full-width punctuation
 */
function fixPunctuation(text) {
  // Half-width comma between CJK chars → full-width
  text = text.replace(
    new RegExp(`([${CJK}])\\s*,\\s*([${CJK}])`, 'g'),
    '$1，$2'
  )
  // Half-width period between CJK chars → full-width
  text = text.replace(
    new RegExp(`([${CJK}])\\s*\\.\\s*([${CJK}])`, 'g'),
    '$1。$2'
  )
  // Half-width colon after CJK → full-width
  text = text.replace(
    new RegExp(`([${CJK}])\\s*:\\s*`, 'g'),
    '$1：'
  )
  // Half-width semicolon between CJK → full-width
  text = text.replace(
    new RegExp(`([${CJK}])\\s*;\\s*([${CJK}])`, 'g'),
    '$1；$2'
  )
  // Remove space before full-width punctuation
  text = text.replace(/\s+([，。：；！？、）】」』])/g, '$1')
  // Remove space after opening full-width brackets
  text = text.replace(/([（【「『])\s+/g, '$1')
  return text
}

/**
 * Normalize paragraph spacing: ensure single blank line between paragraphs.
 */
function normalizeParagraphs(text) {
  // Collapse 3+ consecutive blank lines into 2 (one blank line between paragraphs)
  text = text.replace(/\n{3,}/g, '\n\n')
  // Remove trailing whitespace from lines
  text = text.replace(/[ \t]+$/gm, '')
  return text
}

/**
 * Remove duplicate spaces in inline text (not in code blocks or links).
 * Preserves markdown syntax.
 */
function normalizeInlineSpaces(text) {
  const lines = text.split('\n')
  const result = []
  let inCodeBlock = false

  for (const line of lines) {
    if (line.startsWith('```')) {
      inCodeBlock = !inCodeBlock
      result.push(line)
      continue
    }
    if (inCodeBlock) {
      result.push(line)
      continue
    }
    // Skip lines that are markdown structural elements
    if (line.startsWith('#') || line.startsWith('>') || line.startsWith('- ') ||
        line.startsWith('* ') || line.startsWith('|') || line.startsWith('```') ||
        line.match(/^\d+\.\s/)) {
      result.push(line)
      continue
    }
    // Collapse multiple spaces in regular text (preserve single space)
    let normalized = line.replace(/ {2,}/g, ' ')
    result.push(normalized)
  }

  return result.join('\n')
}

/**
 * Main formatting entry point. Applies all formatting rules.
 * Skips code blocks (``` ... ```).
 */
export function formatMarkdown(text) {
  if (!text || !text.trim()) return text

  // Split into code blocks and non-code sections to avoid formatting code
  const parts = text.split(/(```[\s\S]*?```)/g)

  const formatted = parts.map((part, i) => {
    // Even indices are non-code, odd indices are code blocks
    if (i % 2 === 1) return part

    let result = part

    // Apply formatting rules in order
    result = cjkLatinSpacing(result)
    result = fixPunctuation(result)
    result = normalizeInlineSpaces(result)
    result = normalizeParagraphs(result)

    return result
  })

  return formatted.join('')
}
