const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs-extra');

async function runDartServer({ executionCwd, entry, isStdio, mcpArgs = [], mergedEnv, log }) {
  const pubspecPath = path.join(executionCwd, 'pubspec.yaml');
  if (!(await fs.pathExists(pubspecPath))) {
    log('❌ pubspec.yaml not found. Cannot run Dart server.');
    process.exit(1);
  }
  const cmd = 'dart';
  const cmdArgs = ['run', entry, ...mcpArgs];
  if (isStdio) {
    cmdArgs.push('stdio');
  }
  return spawn(cmd, cmdArgs, { cwd: executionCwd, env: mergedEnv, windowsHide: true, shell: true });
}

module.exports = { runDartServer }; 