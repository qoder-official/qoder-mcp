const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs-extra');
const shell = require('shelljs');

/**
 * Launch a Node-based MCP server.
 *
 * @param {Object} params – named parameters
 * @param {string} params.id – server id (folder name)
 * @param {Object} params.mcpConfig – parsed mcp.yml config
 * @param {string} params.executionCwd – working directory where the server should be executed
 * @param {boolean} params.isStdio – whether to start the server in stdio mode
 * @param {string} params.entry – explicit entry command from mcp.yml (may be empty)
 * @param {Object} params.mergedEnv – environment variables for the child process
 * @param {(msg: string) => void} params.log – logger utility
 * @returns {ChildProcess} – spawned child process
 */
async function runNodeServer({ id, mcpConfig, executionCwd, isStdio, entry, mergedEnv, log }) {
  let cmd;
  let cmdArgs = [];
  const pkgJsonPath = path.join(executionCwd, 'package.json');
  if (!(await fs.pathExists(pkgJsonPath))) {
    log(`❌ No package.json found for '${id}'`);
    process.exit(1);
  }

  const pkg = await fs.readJson(pkgJsonPath);

  // 1. Build step (if present)
  if (pkg.scripts?.build) {
    log('🏗️  Running build script…');
    const buildRes = shell.exec('npm run build', { cwd: executionCwd, silent: true });
    if (buildRes.code !== 0) {
      log(`❌ Build failed for '${id}'.`);
      log(buildRes.stderr);
      process.exit(1);
    }
    log('✅ Build complete.');
  }

  // 2. Determine the command to run
  if (pkg.scripts?.start) {
    cmd = 'npm';
    cmdArgs = ['start'];
    // Inject user-defined args first (npm treats them after -- separator)
    if (Array.isArray(mcpConfig.args) && mcpConfig.args.length > 0) {
      cmdArgs.push('--', ...mcpConfig.args);
    }
    const shouldInjectStdio = isStdio && !mcpConfig.skip_stdio_arg;
    if (shouldInjectStdio) {
      const startScript = pkg.scripts.start;
      const stdioAlready = startScript.includes('--stdio') || (Array.isArray(mcpConfig.args) && mcpConfig.args.includes('--stdio'));
      if (!stdioAlready) {
        if (!cmdArgs.includes('--')) cmdArgs.push('--');
        cmdArgs.push('--stdio');
      }
    }
  } else if (entry) {
    const entryParts = entry.trim().split(/\s+/);
    let executable = entryParts[0];
    cmdArgs = entryParts.slice(1);

    // Prepend any user args from config
    if (Array.isArray(mcpConfig.args) && mcpConfig.args.length > 0) {
      cmdArgs.push(...mcpConfig.args);
    }

    if (executable.endsWith('.js')) {
      cmdArgs.unshift(executable);
      executable = 'node';
    } else if (executable.endsWith('.ts')) {
      cmdArgs.unshift(executable);
      executable = 'ts-node';
    }
    cmd = executable;

    if (cmd === 'ts-node') {
      // Ensure ts-node is resolved via npx so we don't require a global install
      cmdArgs.unshift('ts-node');
      cmd = 'npx';
    }

    if (isStdio && !mcpConfig.skip_stdio_arg && !cmdArgs.includes('--stdio')) {
      cmdArgs.push('--stdio');
    }
  } else {
    log(`❌ No "start" script in package.json and no "entry" in mcp.yml for '${id}'`);
    process.exit(1);
  }

  return spawn(cmd, cmdArgs, {
    cwd: executionCwd,
    env: mergedEnv,
    windowsHide: true,
    shell: true,
  });
}

module.exports = { runNodeServer }; 