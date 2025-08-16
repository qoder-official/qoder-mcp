/**
 * Canonical identifiers for MCP server implementation languages.
 * Using a frozen object emulates an enum in JavaScript while remaining
 * completely type-safe when imported via JSDoc or TypeScript.
 */
const ServerLang = Object.freeze({
  NODE: 'node',
  GO: 'go',
  PYTHON: 'python',
  DART: 'dart',
});

module.exports = { ServerLang }; 