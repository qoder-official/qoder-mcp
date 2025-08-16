const { getServers } = require('../utils');

/**
 * Registers the `list` command with Commander.
 * @param {import('commander').Command} program
 */
function registerListCommand(program) {
  program
    .command('list')
    .description('List all available MCP servers')
    .action(async () => {
      try {
        const servers = await getServers();
        console.log('Available MCP servers:');
        servers.forEach((server) => {
          console.log(`- ${server.id} (${server.config.name || server.id})`);
        });
      } catch (error) {
        console.error('Error listing servers:', error);
        process.exit(1);
      }
    });
}

module.exports = registerListCommand; 