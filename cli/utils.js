const fs = require('fs-extra');
const path = require('path');
const yaml = require('js-yaml');

// Paths relative to project root
const projectRoot = path.resolve(__dirname, '..');
const serversRoot = path.join(projectRoot, 'servers');

/**
 * Convert a server id such as "desktop-commander-node" into a more readable
 * name – e.g. "Desktop Commander".  The "-node", "-go" and "-python" suffixes
 * are stripped because they are implementation details rather than part of the
 * marketing name exposed to users.
 * @param {string} str – the raw server id
 */
function toPrettyName(str) {
  const sanitizedStr = str.replace(/-node$|-go$|-python$/, '');
  return sanitizedStr
    .split('-')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Scan the servers directory and return an array of objects with id + parsed
 * mcp.yml config for every embedded MCP server.
 * @returns {Promise<Array<{id: string, config: any}>>}
 */
async function getServers() {
  const servers = [];
  const dirs = await fs.readdir(serversRoot);
  for (const id of dirs) {
    const serverDir = path.join(serversRoot, id);
    const mcpYmlPath = path.join(serverDir, 'mcp.yml');
    if ((await fs.stat(serverDir)).isDirectory() && (await fs.pathExists(mcpYmlPath))) {
      const mcpConfig = yaml.load(await fs.readFile(mcpYmlPath, 'utf8'));
      servers.push({ id, config: mcpConfig });
    }
  }
  return servers;
}

module.exports = {
  toPrettyName,
  getServers,
}; 