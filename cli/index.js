#!/usr/bin/env node
/*
 * Bootstrap file for the Qoder MCP command-line interface.
 *
 * For backwards-compatibility we temporarily delegate to the existing
 * monolithic implementation in the repository root (cli.js).  This allows us
 * to introduce the new folder-based architecture incrementally while keeping
 * the published entry-point stable.
 *
 * Going forward, individual commands will be moved into the `cli/commands`
 * sub-folder and wired up here via Commander, following the SOLID principles
 * for maintainability and scalability.
 */

// CLI bootstrap – registers command modules then parses argv.
const { Command } = require('commander');

const registerListCommand = require('./commands/list');
const registerSetupCommand = require('./commands/setup');
const registerStartCommand = require('./commands/start');

const program = new Command();
program.name('qoder-mcp').description('Qoder MCP server management CLI');

// Register sub-commands
registerListCommand(program);
registerSetupCommand(program);
registerStartCommand(program);

program.parse(process.argv); 