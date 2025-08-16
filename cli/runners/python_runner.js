const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs-extra');

/**
 * Build the command to run a Python MCP server.
 * Mirrors logic previously embedded in cli.js but encapsulated.
 */
async function runPythonServer({ id, executionCwd, isStdio, entry, mcpArgs = [], mergedEnv, log }) {
  const pyProjectTomlPath = path.join(executionCwd, 'pyproject.toml');
  const pythonEnv = { ...mergedEnv, PYTHONIOENCODING: 'UTF-8' };

  let cmd;
  let cmdArgs = [];

  if (await fs.pathExists(pyProjectTomlPath)) {
    log('🐍 Found pyproject.toml, attempting to run as a module.');
    const pyprojectContent = await fs.readFile(pyProjectTomlPath, 'utf8');
    const nameMatch = pyprojectContent.match(/name\s*=\s*"(.*?)"/);
    const scriptsMatch = pyprojectContent.match(/\[project\.scripts\]\s*([\s\S]*?)(?:\[|$)/);

    if (entry && scriptsMatch && scriptsMatch[1]) {
      const scriptLineMatch = scriptsMatch[1]
        .split(/\r?\n/)
        .find((line) => line.startsWith(entry));
      if (scriptLineMatch) {
        const moduleMatch = scriptLineMatch.match(/=\s*"(.*?):/);
        if (moduleMatch && moduleMatch[1]) {
          cmd = 'python';
          cmdArgs = ['-m', moduleMatch[1].trim()];
        }
      }
    }

    if (!cmd && nameMatch && nameMatch[1]) {
      const moduleToRun = nameMatch[1].replace(/-/g, '_');
      cmd = 'python';
      cmdArgs = ['-m', moduleToRun];
    }
  }

  if (!cmd && entry) {
    if (entry.endsWith('.py')) {
      cmd = 'python';
      cmdArgs = [entry];
    } else {
      // console script installed by pip
      cmd = entry;
      cmdArgs = [];
    }
  }

  if (!cmd) {
    log('❌ No "entry" in mcp.yml or runnable package structure found for Python server.');
    process.exit(1);
  }

  cmdArgs.push(...mcpArgs);
  if (isStdio) {
    if (id === 'flutter-mcp-python') {
      cmdArgs.push('start', '--transport', 'stdio');
    } else {
      cmdArgs.push('--stdio');
    }
  }

  return spawn(cmd, cmdArgs, { cwd: executionCwd, env: pythonEnv, windowsHide: true, shell: true });
}

module.exports = { runPythonServer }; 