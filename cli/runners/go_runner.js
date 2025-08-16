const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs-extra');

/**
 * Locate the main.go file starting from a directory, recursing into sub-folders
 * under /cmd (preferred) or falling back to the project root.
 * @param {string} dir – directory in which to search recursively
 * @returns {Promise<string|null>} – absolute path to main.go or null
 */
async function findMainGo(dir) {
  const files = await fs.readdir(dir);
  for (const file of files) {
    const fullPath = path.join(dir, file);
    const stat = await fs.stat(fullPath);
    if (stat.isDirectory()) {
      const found = await findMainGo(fullPath);
      if (found) return found;
    } else if (file === 'main.go') {
      return fullPath;
    }
  }
  return null;
}

/**
 * Spawn a Go-based MCP server using `go run`.
 * @param {Object} params – named params
 * @param {string} params.id
 * @param {string} params.executionCwd
 * @param {boolean} params.isStdio
 * @param {Array<string>} params.mcpArgs – additional args defined in mcp.yml
 * @param {Object} params.mergedEnv
 * @param {(msg: string) => void} params.log
 * @returns {ChildProcess}
 */
async function runGoServer({ id, executionCwd, isStdio, mcpArgs = [], mergedEnv, log }) {
  // Discover main.go
  let goMainPath = '';
  const cmdDir = path.join(executionCwd, 'cmd');
  if (await fs.pathExists(cmdDir)) {
    goMainPath = await findMainGo(cmdDir);
  }
  if (!goMainPath && await fs.pathExists(path.join(executionCwd, 'main.go'))) {
    goMainPath = path.join(executionCwd, 'main.go');
  }
  if (!goMainPath) {
    log('❌ Unable to locate main.go for Go server');
    process.exit(1);
  }

  const cmd = 'go';
  const cmdArgs = ['run', goMainPath, ...mcpArgs];
  if (isStdio && id !== 'gitlab-go') {
    cmdArgs.push('stdio');
  }

  return spawn(cmd, cmdArgs, { cwd: executionCwd, env: mergedEnv, windowsHide: true });
}

module.exports = { runGoServer }; 