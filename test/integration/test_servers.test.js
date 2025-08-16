const { spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

const { ServerLang } = require('../../cli/constants');

const isWin = process.platform === 'win32';
const projectRoot = path.join(__dirname, '..', '..');
const serversDir = path.join(projectRoot, 'servers');

// -----------------------------------------------------------------------------
// Utility helpers
// -----------------------------------------------------------------------------
function getServersByLang() {
  const result = {
    [ServerLang.NODE]: [],
    [ServerLang.GO]: [],
    [ServerLang.PYTHON]: [],
    [ServerLang.DART]: [],
  };

  const allEntries = fs
    .readdirSync(serversDir)
    .filter((name) => fs.statSync(path.join(serversDir, name)).isDirectory());

  for (const id of allEntries) {
    try {
      const mcpPath = path.join(serversDir, id, 'mcp.yml');
      if (!fs.existsSync(mcpPath)) continue; // skip if no config
      const config = yaml.load(fs.readFileSync(mcpPath, 'utf8')) || {};
      if (!config.lang) continue; // skip if lang undefined
      if (!result[config.lang]) result[config.lang] = [];
      result[config.lang].push(id);
    } catch (_) {
      // ignore parsing errors for now
    }
  }
  return result;
}

function runSetup() {
  console.log('🛠️  Ensuring dependencies with "npm run setup" ...');
  const setupResult = spawnSync('npm', ['run', '--silent', 'setup'], {
    stdio: 'inherit',
    shell: isWin,
  });
  if (setupResult.status !== 0) {
    console.error(`Setup failed (exit ${setupResult.status}).`);
    process.exit(setupResult.status ?? 1);
  }
  console.log('✅ Dependencies ready.\n');
}

function testServer(id, timeoutSec) {
  console.log(`\n→ Testing ${id} (timeout ${timeoutSec}s)`);
  const res = spawnSync(
    'npm',
    [
      'run',
      '--silent',
      'start',
      '--',
      id,
      '--stdio',
      '--timeout',
      String(timeoutSec),
      '--no-install',
    ],
    {
      stdio: 'inherit',
      shell: isWin,
      timeout: timeoutSec * 1000 + 5000, // buffer
    },
  );
  if (res.status !== 0 && res.status !== null) {
    console.error(`❌ ${id} failed (exit ${res.status}). Stopping tests.`);
    process.exit(res.status ?? 1);
  }
}

// -----------------------------------------------------------------------------
// Main entry
// -----------------------------------------------------------------------------
(function main() {
  const timeoutSec = parseInt(process.env.TEST_SERVER_TIMEOUT_SEC || '15', 10);
  runSetup();

  const serversByLang = getServersByLang();

  for (const lang of Object.values(ServerLang)) {
    const list = serversByLang[lang] || [];
    if (list.length === 0) continue;
    console.log(`\n============================\n# Testing ${lang.toUpperCase()} servers\n============================`);
    for (const id of list) {
      testServer(id, timeoutSec);
    }
  }

  console.log('\n🎉 All servers tested successfully.');
})();