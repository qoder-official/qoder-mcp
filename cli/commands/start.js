const path = require('path');
const fs = require('fs-extra');
const yaml = require('js-yaml');
const { spawn } = require('child_process');

const { installDependencies } = require('../install_dependencies');
const { toPrettyName } = require('../utils');
const { projectRoot, serversRoot } = require('../context');
const { runNodeServer } = require('../runners/node_runner');
const { runGoServer } = require('../runners/go_runner');
const { runPythonServer } = require('../runners/python_runner');
const { runDartServer } = require('../runners/dart_runner');

/**
 * Registers the `start` command.
 * @param {import('commander').Command} program
 */
function registerStartCommand(program) {
  program
    .command('start <id>')
    .description('Builds and runs a specific MCP server.')
    .option('--stdio', 'Run in stdio mode for JSON-RPC communication')
    .option('--docker', 'Run the server inside Docker (legacy behaviour)')
    .option('--alias <name>', 'Run with a specific environment variable alias')
    .option(
      '--timeout <seconds>',
      'Maximum run time before the server is automatically stopped (0 = no timeout)',
      parseInt,
    )
    .option('--no-install', 'Do not install dependencies before running')
    .action(async (id, options) => {
      const isStdio = !!options.stdio;
      const alias = options.alias;
      const timeoutSec = options.timeout ?? 0;
      const log = (msg) => (isStdio ? console.error(msg) : console.log(msg));

      const serverDir = path.join(serversRoot, id);
      if (!(await fs.pathExists(serverDir))) {
        log(`Error: Server with ID '${id}' not found at ${serverDir}`);
        process.exit(1);
      }

      try {
        const mcpConfig = yaml.load(await fs.readFile(path.join(serverDir, 'mcp.yml'), 'utf8'));
        if (!options.noInstall) {
          await installDependencies({ id, config: mcpConfig }, log);
        }

        const executionCwd = mcpConfig.working_dir
          ? path.join(serverDir, mcpConfig.working_dir)
          : serverDir;
        const lang = mcpConfig.lang;
        const entry = mcpConfig.entry || mcpConfig.run || '';

        log(`🚀 Starting server '${id}'...`);

        // Build environment
        const mergedEnv = { ...process.env };
        const envPath = path.join(projectRoot, '.env');
        if (await fs.pathExists(envPath)) {
          const envContent = await fs.readFile(envPath, 'utf8');
          envContent.split(/\r?\n/).forEach((line) => {
            const trimmedLine = line.trim();
            if (trimmedLine && !trimmedLine.startsWith('#')) {
              const separatorIndex = trimmedLine.indexOf('=');
              if (separatorIndex > 0) {
                const key = trimmedLine.substring(0, separatorIndex).trim();
                const value = trimmedLine.substring(separatorIndex + 1).trim();
                mergedEnv[key] = mergedEnv[key] ?? value;
              }
            }
          });
        }

        // Alias handling
        if (alias && Array.isArray(mcpConfig.env)) {
          for (const envVar of mcpConfig.env) {
            if (envVar.key && envVar.as === 'env') {
              const aliasedKey = `${envVar.key}_${alias.toUpperCase()}`;
              if (mergedEnv[aliasedKey]) {
                mergedEnv[envVar.key] = mergedEnv[aliasedKey];
              }
            }
          }
        }

        // log file setup
        await fs.ensureDir(path.join(projectRoot, 'server-logs'));
        const logFilePath = path.join(projectRoot, 'server-logs', `${id}-${Date.now()}.log`);
        const logStream = fs.createWriteStream(logFilePath, { flags: 'a' });

        const mcpArgs = Array.isArray(mcpConfig.args) ? [...mcpConfig.args] : [];
        let child;
        if (lang === 'node') {
          child = await runNodeServer({ id, mcpConfig, executionCwd, isStdio, entry, mergedEnv, log });
        } else if (lang === 'go') {
          child = await runGoServer({ id, executionCwd, isStdio, mcpArgs, mergedEnv, log });
        } else if (lang === 'python') {
          child = await runPythonServer({ id, executionCwd, isStdio, entry, mcpArgs, mergedEnv, log });
        } else if (lang === 'dart') {
          child = await runDartServer({ executionCwd, entry, isStdio, mcpArgs, mergedEnv, log });
        } else {
          log(`Unsupported language: ${lang}`);
          process.exit(1);
        }

        const writeStdout = (chunk) => {
          logStream.write(chunk);
          if (!isStdio) process.stdout.write(chunk);
        };
        const writeStderr = (chunk) => {
          logStream.write(chunk);
          if (!isStdio) process.stderr.write(chunk);
        };

        if (child) {
          child.stdout.on('data', writeStdout);
          child.stderr.on('data', writeStderr);
          if (isStdio) {
            process.stdin.pipe(child.stdin);
            child.stdout.pipe(process.stdout);
            child.stderr.pipe(process.stderr);
          }
          child.on('close', (code) => {
            log(`Server '${id}' exited with code ${code}.`);
            logStream.end();
          });
          child.on('error', (err) => {
            log(`Error running server '${id}': ${err.message}`);
            logStream.end();
          });
          if (timeoutSec > 0) {
            setTimeout(() => {
              log(`⏳ Server '${id}' has run for ${timeoutSec}s. Stopping...`);
              child.kill();
            }, timeoutSec * 1000);
          }
        }
      } catch (error) {
        console.error(`Error starting server ${id}:`, error);
        process.exit(1);
      }
    });
}

module.exports = registerStartCommand; 