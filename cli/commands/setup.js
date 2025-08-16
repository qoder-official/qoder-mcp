const path = require('path');
const os = require('os');
const fs = require('fs-extra');
const shell = require('shelljs');

const { getServers, toPrettyName } = require('../utils');
const { installDependencies } = require('../install_dependencies');
const { serversRoot, runMcpBatTemplate, runMcpShTemplate } = require('../context');

/**
 * Registers the `setup` command.
 * @param {import('commander').Command} program
 */
function registerSetupCommand(program) {
  program
    .command('setup')
    .description('Installs dependencies for all servers. Optionally, configures servers for Cursor.')
    .option('--cursor-setup', 'Configure servers for Cursor by creating runner scripts and mcp.json.')
    .option('--flutter-essentials', 'Filters the --cursor-setup to only include essential Flutter development servers.')
    .argument('[aliases]', 'Optional comma-separated list of server aliases to use with --cursor-setup (e.g., figma-node:work,github-go:personal)')
    .action(async (aliasesStr, options) => {
      if (options.cursorSetup) {
        console.log('Configuring servers for Cursor...');
        try {
          const aliases = (aliasesStr || '').split(',').reduce((acc, curr) => {
            if (!curr) return acc;
            const [id, alias] = curr.split(':');
            if (id && alias) {
              if (!acc[id]) acc[id] = [];
              acc[id].push(alias);
            }
            return acc;
          }, {});

          const mcpJsonPath = path.join(os.homedir(), '.cursor', 'mcp.json');
          let servers = await getServers();

          if (options.flutterEssentials) {
            console.log('Filtering for Flutter Essentials...');
            const essentialServers = [
              'better-fetch-node',
              'compass-node',
              'desktop-commander-node',
              'duckduckgo-python',
              'figma-node',
              'postman-node',
              'mcp_flutter-dart',
              'flutter-mcp-python',
              'sequential-thinking-node',
            ];
            servers = servers.filter((server) => essentialServers.includes(server.id));
          }

          servers.sort((a, b) => {
            const nameA = a.config.name || toPrettyName(a.id);
            const nameB = b.config.name || toPrettyName(b.id);
            return nameA.localeCompare(nameB);
          });

          const mcpJson = { mcpServers: {} };

          const isWindows = os.platform() === 'win32';
          const runnerName = isWindows ? 'run-mcp.cmd' : 'run-mcp';
          const runnerPath = path.join(__dirname, '..', runnerName);

          for (const server of servers) {
            const config = server.config;
            const baseName = config.name || toPrettyName(server.id);
            const prettyName = `Qoder - ${baseName}`;

            mcpJson.mcpServers[prettyName] = {
              command: runnerPath,
              args: ['start', server.id, '--stdio'],
            };

            if (aliases[server.id]) {
              for (const alias of aliases[server.id]) {
                const aliasPrettyName = toPrettyName(alias);
                const aliasName = `${prettyName} (${aliasPrettyName})`;
                mcpJson.mcpServers[aliasName] = {
                  command: runnerPath,
                  args: ['start', server.id, '--stdio', `--alias=${alias}`],
                };
              }
            }
          }

          await fs.ensureDir(path.dirname(mcpJsonPath));
          await fs.writeJson(mcpJsonPath, mcpJson, { spaces: 2 });
          console.log(`✅ MCP JSON updated at ${mcpJsonPath}`);

          const runnerTemplate = isWindows ? runMcpBatTemplate : runMcpShTemplate;
          await fs.writeFile(runnerPath, runnerTemplate);
          if (!isWindows) {
            await fs.chmod(runnerPath, '755');
          }
          console.log(`✅ Created ${runnerName} runner script.`);
        } catch (error) {
          console.error('Error during setup:', error);
          process.exit(1);
        }
      } else {
        console.log('Installing dependencies for all servers...');
        const servers = await getServers();
        const log = (msg) => console.log(msg);
        for (const server of servers) {
          await installDependencies(server, log);
        }
        console.log('\nAll server dependencies are up to date.');
      }
    });
}

module.exports = registerSetupCommand; 