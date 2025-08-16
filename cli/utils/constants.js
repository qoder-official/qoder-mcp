const path = require('path');

// Resolve project root two levels up from this utils directory
const projectRoot = path.resolve(__dirname, '..', '..');
const serversRoot = path.join(projectRoot, 'servers');
const dockerRoot = path.join(projectRoot, 'docker');

module.exports = {
  projectRoot,
  serversRoot,
  dockerRoot,
}; 